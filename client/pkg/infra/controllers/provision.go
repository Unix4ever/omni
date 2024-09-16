// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package controllers

import (
	"context"
	"slices"

	"github.com/cosi-project/runtime/pkg/controller"
	"github.com/cosi-project/runtime/pkg/controller/generic"
	"github.com/cosi-project/runtime/pkg/resource"
	"github.com/cosi-project/runtime/pkg/resource/protobuf"
	"github.com/cosi-project/runtime/pkg/safe"
	"github.com/cosi-project/runtime/pkg/state"
	"github.com/siderolabs/gen/optional"
	"go.uber.org/zap"

	"github.com/siderolabs/omni/client/api/omni/specs"
	infrares "github.com/siderolabs/omni/client/pkg/infra/internal/resources"
	"github.com/siderolabs/omni/client/pkg/infra/provision"
	"github.com/siderolabs/omni/client/pkg/omni/resources"
	"github.com/siderolabs/omni/client/pkg/omni/resources/infra"
	"github.com/siderolabs/omni/client/pkg/omni/resources/omni"
	"github.com/siderolabs/omni/client/pkg/omni/resources/siderolink"
)

const currentStepAnnotation = "infra." + omni.SystemLabelPrefix + "step"

// ProvisionController is the generic controller that operates the Provisioner.
type ProvisionController[T generic.ResourceWithRD] struct {
	generic.NamedController
	provisioner  provision.Provisioner[T]
	imageFactory provision.FactoryClient
	providerID   string
	concurrency  uint
}

// NewProvisionController creates new ProvisionController.
func NewProvisionController[T generic.ResourceWithRD](providerID string, provisioner provision.Provisioner[T], concurrency uint,
	imageFactory provision.FactoryClient,
) *ProvisionController[T] {
	return &ProvisionController[T]{
		NamedController: generic.NamedController{
			ControllerName: providerID + ".ProvisionController",
		},
		providerID:   providerID,
		provisioner:  provisioner,
		concurrency:  concurrency,
		imageFactory: imageFactory,
	}
}

// Settings implements controller.QController interface.
func (ctrl *ProvisionController[T]) Settings() controller.QSettings {
	var t T

	return controller.QSettings{
		Inputs: []controller.Input{
			{
				Namespace: resources.InfraProviderNamespace,
				Type:      infra.MachineRequestType,
				Kind:      controller.InputQPrimary,
			},
			{
				Namespace: resources.InfraProviderNamespace,
				Type:      infra.MachineRequestStatusType,
				Kind:      controller.InputQMappedDestroyReady,
			},
			{
				Namespace: resources.DefaultNamespace,
				Type:      siderolink.ConnectionParamsType,
				Kind:      controller.InputQMapped,
				ID:        optional.Some(siderolink.ConfigID),
			},
			{
				Namespace: t.ResourceDefinition().DefaultNamespace,
				Type:      t.ResourceDefinition().Type,
				Kind:      controller.InputQMappedDestroyReady,
			},
		},
		Outputs: []controller.Output{
			{
				Kind: controller.OutputExclusive,
				Type: infra.MachineRequestStatusType,
			},
			{
				Kind: controller.OutputShared,
				Type: t.ResourceDefinition().Type,
			},
		},
		Concurrency: optional.Some(ctrl.concurrency),
	}
}

// MapInput implements controller.QController interface.
func (ctrl *ProvisionController[T]) MapInput(_ context.Context, _ *zap.Logger,
	_ controller.QRuntime, ptr resource.Pointer,
) ([]resource.Pointer, error) {
	if ptr.Type() == siderolink.ConnectionParamsType {
		return nil, nil
	}

	return []resource.Pointer{
		infra.NewMachineRequest(ptr.ID()).Metadata(),
	}, nil
}

// Reconcile implements controller.QController interface.
func (ctrl *ProvisionController[T]) Reconcile(ctx context.Context,
	logger *zap.Logger, r controller.QRuntime, ptr resource.Pointer,
) error {
	machineRequest, err := safe.ReaderGet[*infra.MachineRequest](ctx, r, infra.NewMachineRequest(ptr.ID()).Metadata())
	if err != nil {
		if state.IsNotFoundError(err) {
			return nil
		}

		return err
	}

	if machineRequest.Metadata().Phase() == resource.PhaseTearingDown {
		return ctrl.reconcileTearingDown(ctx, r, logger, machineRequest)
	}

	if !machineRequest.Metadata().Finalizers().Has(ctrl.Name()) {
		if err = r.AddFinalizer(ctx, machineRequest.Metadata(), ctrl.Name()); err != nil {
			return err
		}
	}

	machineRequestStatus, err := ctrl.initializeStatus(ctx, r, logger, machineRequest)
	if err != nil {
		return err
	}

	return safe.WriterModify(ctx, r, machineRequestStatus, func(res *infra.MachineRequestStatus) error {
		return ctrl.reconcileRunning(ctx, r, logger, machineRequest, res)
	})
}

func (ctrl *ProvisionController[T]) reconcileRunning(ctx context.Context, r controller.QRuntime, logger *zap.Logger,
	machineRequest *infra.MachineRequest, machineRequestStatus *infra.MachineRequestStatus,
) error {
	connectionParams, err := ctrl.getConnectionArgs(ctx, r)
	if err != nil {
		return err
	}

	var t T

	md := resource.NewMetadata(infrares.ResourceNamespace(ctrl.providerID), t.ResourceDefinition().Type, machineRequest.Metadata().ID(), resource.VersionUndefined)

	var res resource.Resource

	res, err = r.Get(ctx, md)
	if err != nil && !state.IsNotFoundError(err) {
		return err
	}

	if res == nil {
		res, err = protobuf.CreateResource(t.ResourceDefinition().Type)
		if err != nil {
			return err
		}

		*res.Metadata() = md

		// initialize empty spec
		if r, ok := res.Spec().(interface {
			UnmarshalJSON(bytes []byte) error
		}); ok {
			if err = r.UnmarshalJSON([]byte("{}")); err != nil {
				return err
			}
		}
	}

	// nothing to do as the machine was already provisioned
	if machineRequestStatus.TypedSpec().Value.Stage == specs.MachineRequestStatusSpec_PROVISIONED {
		return nil
	}

	steps := ctrl.provisioner.ProvisionSteps()

	initialStep, ok := res.Metadata().Annotations().Get(currentStepAnnotation)

	var initialStepIndex int

	if ok {
		if index := slices.IndexFunc(steps, func(step provision.Step[T]) bool {
			return step.Name() == initialStep
		}); index != -1 {
			initialStepIndex = index
		}
	}

	for _, step := range steps[initialStepIndex:] {
		if initialStep != "" && step.Name() != initialStep {
			continue
		}

		initialStep = ""

		logger.Info("running provision step", zap.String("step", step.Name()))

		res.Metadata().Annotations().Set(currentStepAnnotation, step.Name())

		if err = safe.WriterModify(ctx, r, res.(T), func(st T) error { //nolint:forcetypeassert
			return step.Run(ctx, logger, provision.NewContext(
				machineRequest,
				machineRequestStatus,
				st,
				connectionParams,
				ctrl.imageFactory,
			))
		}); err != nil {
			logger.Error("machine provision failed", zap.Error(err), zap.String("step", step.Name()))

			machineRequestStatus.TypedSpec().Value.Error = err.Error()
			machineRequestStatus.TypedSpec().Value.Stage = specs.MachineRequestStatusSpec_FAILED

			return nil
		}

		if err = safe.WriterModify(ctx, r, machineRequestStatus, func(res *infra.MachineRequestStatus) error {
			res.TypedSpec().Value = machineRequestStatus.TypedSpec().Value

			return nil
		}); err != nil {
			return err
		}
	}

	machineRequestStatus.TypedSpec().Value.Stage = specs.MachineRequestStatusSpec_PROVISIONED

	*machineRequestStatus.Metadata().Labels() = *machineRequest.Metadata().Labels()

	return nil
}

func (ctrl *ProvisionController[T]) initializeStatus(ctx context.Context, r controller.QRuntime, logger *zap.Logger, machineRequest *infra.MachineRequest) (*infra.MachineRequestStatus, error) {
	mrs, err := safe.ReaderGetByID[*infra.MachineRequestStatus](ctx, r, machineRequest.Metadata().ID())
	if err != nil && !state.IsNotFoundError(err) {
		return nil, err
	}

	if mrs != nil {
		return mrs, nil
	}

	return safe.WriterModifyWithResult(ctx, r, infra.NewMachineRequestStatus(machineRequest.Metadata().ID()), func(res *infra.MachineRequestStatus) error {
		if res.TypedSpec().Value.Stage == specs.MachineRequestStatusSpec_UNKNOWN {
			res.TypedSpec().Value.Stage = specs.MachineRequestStatusSpec_PROVISIONING
			*res.Metadata().Labels() = *machineRequest.Metadata().Labels()

			logger.Info("machine provision started", zap.String("request_id", machineRequest.Metadata().ID()))
		}

		return nil
	})
}

func (ctrl *ProvisionController[T]) getConnectionArgs(ctx context.Context, r controller.QRuntime) (string, error) {
	connectionParams, err := safe.ReaderGetByID[*siderolink.ConnectionParams](ctx, r, siderolink.ConfigID)
	if err != nil {
		return "", err
	}

	return siderolink.GetConnectionArgsForProvider(connectionParams, ctrl.providerID)
}

func (ctrl *ProvisionController[T]) reconcileTearingDown(ctx context.Context, r controller.QRuntime, logger *zap.Logger, machineRequest *infra.MachineRequest) error {
	t, err := safe.ReaderGetByID[T](ctx, r, machineRequest.Metadata().ID())
	if err != nil && !state.IsNotFoundError(err) {
		return err
	}

	resources := []resource.Metadata{
		resource.NewMetadata(t.ResourceDefinition().DefaultNamespace, t.ResourceDefinition().Type, machineRequest.Metadata().ID(), resource.VersionUndefined),
		*infra.NewMachineRequestStatus(machineRequest.Metadata().ID()).Metadata(),
	}

	for _, md := range resources {
		ready, err := r.Teardown(ctx, md)
		if err != nil {
			if state.IsNotFoundError(err) {
				continue
			}

			return err
		}

		if !ready {
			return nil
		}

		err = r.Destroy(ctx, md)
		if err != nil {
			if state.IsNotFoundError(err) {
				continue
			}

			return err
		}
	}

	if err = ctrl.provisioner.Deprovision(ctx, logger, t, machineRequest); err != nil {
		return err
	}

	logger.Info("machine deprovisioned", zap.String("request_id", machineRequest.Metadata().ID()))

	return r.RemoveFinalizer(ctx, machineRequest.Metadata(), ctrl.Name())
}
