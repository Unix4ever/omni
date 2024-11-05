// Copyright (c) 2024 Sidero Labs, Inc.
//
// Use of this software is governed by the Business Source License
// included in the LICENSE file.

package omni

import (
	"context"
	"errors"
	"fmt"

	"github.com/cosi-project/runtime/pkg/controller"
	"github.com/cosi-project/runtime/pkg/controller/generic"
	"github.com/cosi-project/runtime/pkg/controller/generic/qtransform"
	"github.com/cosi-project/runtime/pkg/resource"
	"github.com/cosi-project/runtime/pkg/safe"
	"github.com/cosi-project/runtime/pkg/state"
	"github.com/siderolabs/gen/optional"
	"github.com/siderolabs/gen/xerrors"
	"go.uber.org/zap"

	"github.com/siderolabs/omni/client/pkg/omni/resources"
	"github.com/siderolabs/omni/client/pkg/omni/resources/omni"
)

// ManagedConfigPatchesControllerName is the name of the ManagedConfigPatchesController.
const ManagedConfigPatchesControllerName = "ManagedConfigPatchesController"

// ManagedConfigPatchesController processes additional teardown steps for a machine leaving a machine set.
type ManagedConfigPatchesController struct {
	generic.NamedController
}

// NewManagedConfigPatchesController initializes ManagedConfigPatchesController.
func NewManagedConfigPatchesController() *ManagedConfigPatchesController {
	return &ManagedConfigPatchesController{
		NamedController: generic.NamedController{
			ControllerName: ManagedConfigPatchesControllerName,
		},
	}
}

// Settings implements controller.QController interface.
func (ctrl *ManagedConfigPatchesController) Settings() controller.QSettings {
	return controller.QSettings{
		Inputs: []controller.Input{
			{
				Namespace: resources.DefaultNamespace,
				Type:      omni.MachineSetType,
				Kind:      controller.InputQPrimary,
			},
			{
				Namespace: resources.DefaultNamespace,
				Type:      omni.ConfigPatchType,
				Kind:      controller.InputQMappedDestroyReady,
			},
		},
		Outputs: []controller.Output{
			{
				Type: omni.ConfigPatchType,
				Kind: controller.OutputShared,
			},
		},
		Concurrency: optional.Some[uint](4),
	}
}

// MapInput implements controller.QController interface.
func (ctrl *ManagedConfigPatchesController) MapInput(ctx context.Context, _ *zap.Logger, r controller.QRuntime, ptr resource.Pointer) ([]resource.Pointer, error) {
	if ptr.Type() == omni.ConfigPatchType {
		configPatch, err := safe.ReaderGet[*omni.ConfigPatch](ctx, r, ptr)
		if err != nil {
			if state.IsNotFoundError(err) {
				return nil, nil
			}

			return nil, err
		}

		if _, managed := configPatch.Metadata().Labels().Get(omni.LabelManaged); !managed {
			return nil, nil
		}

		clusterName, ok := configPatch.Metadata().Labels().Get(omni.LabelCluster)
		if !ok {
			return nil, nil
		}

		machineSets, err := safe.ReaderListAll[*omni.MachineSet](ctx, r,
			state.WithLabelQuery(
				resource.LabelEqual(omni.LabelCluster, clusterName),
				resource.LabelExists(omni.LabelControlPlaneRole),
			),
		)
		if err != nil {
			return nil, err
		}

		return safe.ToSlice(machineSets, func(r *omni.MachineSet) resource.Pointer { return r.Metadata() }), nil
	}

	return nil, fmt.Errorf("unexpected resource type %q", ptr.Type())
}

// Reconcile implements controller.QController interface.
func (ctrl *ManagedConfigPatchesController) Reconcile(ctx context.Context, logger *zap.Logger, r controller.QRuntime, ptr resource.Pointer) error {
	machineSet, err := safe.ReaderGetByID[*omni.MachineSet](ctx, r, ptr.ID())
	if err != nil {
		if state.IsNotFoundError(err) {
			return nil
		}

		return err
	}

	if _, isControlPlane := machineSet.Metadata().Labels().Get(omni.LabelControlPlaneRole); !isControlPlane {
		return nil
	}

	if machineSet.TypedSpec().Value.Managed == nil || !machineSet.TypedSpec().Value.Managed.Enable {
		return nil
	}

	if machineSet.Metadata().Phase() == resource.PhaseTearingDown {
		if err = ctrl.reconcileTearingDown(ctx, r, machineSet); err != nil {
			if xerrors.TagIs[qtransform.SkipReconcileTag](err) {
				return nil
			}

			return err
		}

		return r.RemoveFinalizer(ctx, machineSet.Metadata(), ctrl.Name())
	}

	if !machineSet.Metadata().Finalizers().Has(ctrl.Name()) {
		return r.AddFinalizer(ctx, machineSet.Metadata(), ctrl.Name())
	}

	if err := ctrl.reconcileRunning(ctx, r, machineSet); err != nil {
		return err
	}

	return nil
}

func (ctrl *ManagedConfigPatchesController) reconcileRunning(ctx context.Context, r controller.QRuntime,
	machineSet *omni.MachineSet,
) error {
	clusterName, ok := machineSet.Metadata().Labels().Get(omni.LabelCluster)
	if !ok {
		return errors.New("cluster name label is missing in the machine set")
	}

	return safe.WriterModify(ctx, r, omni.NewConfigPatch(resources.DefaultNamespace, "950-%s"+clusterName+"-kubespan"),
		func(patch *omni.ConfigPatch) error {
			patch.Metadata().Labels().Set(omni.LabelCluster, clusterName)
			patch.Metadata().Labels().Set(omni.LabelManaged, "")
			patch.Metadata().Labels().Set(omni.LabelSystemPatch, "")

			patch.Metadata().Annotations().Set(omni.ConfigPatchName, "KubeSpan")
			patch.Metadata().Annotations().Set(omni.ConfigPatchDescription, "Forcefully enable KubeSpan for managed clusters")

			patch.TypedSpec().Value.SetUncompressedData([]byte(`machine:
  network:
    kubespan:
      enabled: true
`))
			return nil
		})
}

func (ctrl *ManagedConfigPatchesController) reconcileTearingDown(ctx context.Context, r controller.QRuntime,
	machineSet *omni.MachineSet,
) error {
	clusterName, ok := machineSet.Metadata().Labels().Get(omni.LabelCluster)
	if !ok {
		return errors.New("cluster name label is missing in the machine set")
	}

	patches, err := safe.ReaderListAll[*omni.ConfigPatch](ctx, r,
		state.WithLabelQuery(
			resource.LabelEqual(omni.LabelCluster, clusterName),
			resource.LabelExists(omni.LabelManaged),
		),
	)
	if err != nil {
		return err
	}

	destroyReady := true

	for patch := range patches.All() {
		var ready bool

		ready, err = r.Teardown(ctx, patch.Metadata())
		if err != nil {
			if state.IsNotFoundError(err) {
				continue
			}

			return err
		}

		if !ready {
			destroyReady = false

			continue
		}

		err = r.Destroy(ctx, patch.Metadata())
		if err != nil {
			if state.IsNotFoundError(err) {
				continue
			}

			return err
		}
	}

	if !destroyReady {
		return xerrors.NewTaggedf[qtransform.SkipReconcileTag]("the resource is not ready to be destroyed")
	}

	return nil
}
