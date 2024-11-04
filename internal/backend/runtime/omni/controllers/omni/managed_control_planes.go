// Copyright (c) 2024 Sidero Labs, Inc.
//
// Use of this software is governed by the Business Source License
// included in the LICENSE file.

package omni

import (
	"context"
	"errors"

	"github.com/cosi-project/runtime/pkg/controller"
	"github.com/cosi-project/runtime/pkg/controller/generic/qtransform"
	"github.com/cosi-project/runtime/pkg/safe"
	"github.com/siderolabs/gen/optional"
	"github.com/siderolabs/gen/xerrors"
	"go.uber.org/zap"

	"github.com/siderolabs/omni/client/pkg/omni/resources"
	"github.com/siderolabs/omni/client/pkg/omni/resources/omni"
	"github.com/siderolabs/omni/internal/backend/runtime/omni/controllers/omni/internal/mappers"
)

// ManagedControlPlaneController creates omni.ClusterSecrets for each input omni.Cluster.
//
// ManagedControlPlaneController generates and stores cluster wide secrets.
type ManagedControlPlaneController = qtransform.QController[*omni.MachineSet, *omni.MachineRequestSet]

// ProviderConfig defines the infra provider configuration for the managed control planes.
type ProviderConfig struct {
	ID   string
	Data string
}

// NewManagedControlPlaneController instantiates the talosconfig controller.
func NewManagedControlPlaneController(defaultProvider ProviderConfig) *ManagedControlPlaneController {
	return qtransform.NewQController(
		qtransform.Settings[*omni.MachineSet, *omni.MachineRequestSet]{
			Name: "ManagedControlPlaneController",
			MapMetadataOptionalFunc: func(machineSet *omni.MachineSet) optional.Optional[*omni.MachineRequestSet] {
				if _, controlplane := machineSet.Metadata().Labels().Get(omni.LabelControlPlaneRole); !controlplane {
					return optional.None[*omni.MachineRequestSet]()
				}

				return optional.Some(omni.NewMachineRequestSet(resources.DefaultNamespace, machineSet.Metadata().ID()))
			},
			UnmapMetadataFunc: func(machineRequestSet *omni.MachineRequestSet) *omni.MachineSet {
				return omni.NewMachineSet(resources.DefaultNamespace, machineRequestSet.Metadata().ID())
			},
			TransformFunc: func(ctx context.Context, r controller.Reader, _ *zap.Logger, machineSet *omni.MachineSet, machineRequestSet *omni.MachineRequestSet) error {
				if !machineSet.TypedSpec().Value.Managed.Enable {
					return xerrors.NewTaggedf[qtransform.DestroyOutputTag]("control plane is not managed")
				}

				clusterName, ok := machineSet.Metadata().Labels().Get(omni.LabelCluster)
				if !ok {
					return errors.New("cluster name label is missing from the machine set")
				}

				cluster, err := safe.ReaderGetByID[*omni.Cluster](ctx, r, clusterName)
				if err != nil {
					return err
				}

				machineRequestSet.TypedSpec().Value.MachineCount = 3
				machineRequestSet.TypedSpec().Value.TalosVersion = cluster.TypedSpec().Value.TalosVersion
				machineRequestSet.TypedSpec().Value.ProviderId = defaultProvider.ID
				machineRequestSet.TypedSpec().Value.ProviderData = defaultProvider.Data

				machineRequestSet.Metadata().Annotations().Set(omni.LabelNoManualAllocation, "")

				return nil
			},
		},
		qtransform.WithOutputKind(controller.OutputShared),
		qtransform.WithExtraMappedInput(
			mappers.MapClusterResourceToLabeledResources[*omni.Cluster, *omni.MachineSet](),
		),
	)
}
