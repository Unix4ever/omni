// Copyright (c) 2024 Sidero Labs, Inc.
//
// Use of this software is governed by the Business Source License
// included in the LICENSE file.

package omni

import (
	"context"
	"time"

	"github.com/cosi-project/runtime/pkg/controller"
	"github.com/cosi-project/runtime/pkg/controller/generic/qtransform"
	"github.com/cosi-project/runtime/pkg/safe"
	"github.com/cosi-project/runtime/pkg/state"
	"go.uber.org/zap"

	"github.com/siderolabs/omni/client/pkg/omni/resources"
	"github.com/siderolabs/omni/client/pkg/omni/resources/omni"
	"github.com/siderolabs/omni/internal/backend/runtime/omni/controllers/omni/internal/mappers"
)

// MachineProvisionController turns MachineProvision resources into a MachineRequestSets, scales them automatically on demand.
type MachineProvisionController = qtransform.QController[*omni.MachineProvision, *omni.MachineRequestSet]

var teardownGracePeriod = time.Second * 30

// NewMachineProvisionController instanciates the machine controller.
func NewMachineProvisionController() *MachineProvisionController {
	machinesToRemove := map[string]time.Time{}

	return qtransform.NewQController(
		qtransform.Settings[*omni.MachineProvision, *omni.MachineRequestSet]{
			Name: "MachineProvisionController",
			MapMetadataFunc: func(res *omni.MachineProvision) *omni.MachineRequestSet {
				return omni.NewMachineRequestSet(resources.DefaultNamespace, res.Metadata().ID())
			},
			UnmapMetadataFunc: func(res *omni.MachineRequestSet) *omni.MachineProvision {
				return omni.NewMachineProvision(resources.DefaultNamespace, res.Metadata().ID())
			},
			TransformFunc: func(ctx context.Context, r controller.Reader, logger *zap.Logger, provision *omni.MachineProvision, machineRequestSet *omni.MachineRequestSet) error {
				machines, err := safe.ReaderListAll[*machineStatusLabels](ctx, r)
				if err != nil {
					return err
				}

				machineRequestSet.TypedSpec().Value.ProviderId = provision.TypedSpec().Value.ProviderId
				machineRequestSet.TypedSpec().Value.Extensions = provision.TypedSpec().Value.Extensions
				machineRequestSet.TypedSpec().Value.KernelArgs = provision.TypedSpec().Value.KernelArgs
				machineRequestSet.TypedSpec().Value.MetaValues = provision.TypedSpec().Value.MetaValues
				machineRequestSet.TypedSpec().Value.TalosVersion = provision.TypedSpec().Value.TalosVersion
				machineRequestSet.TypedSpec().Value.Overlay = provision.TypedSpec().Value.Overlay

				var (
					readyToDelete     int
					idleMachinesCount int
				)

				for m := range machines.All() {
					_, available := m.Metadata().Labels().Get(omni.MachineStatusLabelAvailable)

					if !available {
						delete(machinesToRemove, m.Metadata().ID())

						continue
					}

					idleMachinesCount++

					if unallocatedTime, ok := machinesToRemove[m.Metadata().ID()]; ok {
						if time.Since(unallocatedTime) > 0 {
							readyToDelete++
						}

						delete(machinesToRemove, m.Metadata().ID())

						continue
					}

					timeSinceCreation := time.Since(m.Metadata().Created())

					var gracePeriod time.Duration

					if timeSinceCreation < teardownGracePeriod {
						gracePeriod = teardownGracePeriod - timeSinceCreation
					}

					machinesToRemove[m.Metadata().ID()] = time.Now().Add(
						provision.TypedSpec().Value.IdleMachineTeardownTimeout.AsDuration() + gracePeriod,
					)
				}

				// TODO: cleanup machinesToRemove

				extraMachines := idleMachinesCount - int(provision.TypedSpec().Value.IdleMachineCount)

				if extraMachines > 0 {
					if readyToDelete > extraMachines {
						readyToDelete = extraMachines
					}

					logger.Info("scale down", zap.Int("count", readyToDelete))

					machineRequestSet.TypedSpec().Value.MachineCount -= int32(readyToDelete)
				}

				pressure, err := safe.ReaderGetByID[*omni.MachineRequestSetPressure](ctx, r, provision.Metadata().ID())
				if err != nil {
					if state.IsNotFoundError(err) {
						return nil
					}

					return err
				}

				expectMachines := pressure.TypedSpec().Value.RequiredMachines + provision.TypedSpec().Value.IdleMachineCount

				delta := expectMachines - uint32(machineRequestSet.TypedSpec().Value.MachineCount)

				if delta > 0 {
					machineRequestSet.TypedSpec().Value.MachineCount = int32(expectMachines)

					logger.Info("scale up", zap.Uint32("count", delta))
				}

				if readyToDelete < extraMachines {
					logger.Info("waiting for timeout to scale down", zap.Int("count", extraMachines-readyToDelete))

					return controller.NewRequeueInterval(time.Second * 30)
				}

				return nil
			},
		},
		qtransform.WithExtraMappedInput(
			qtransform.MapperSameID[*omni.MachineRequestSetPressure, *omni.MachineProvision](),
		),
		qtransform.WithExtraMappedInput(
			mappers.MapExtractLabelValue[*machineStatusLabels, *omni.MachineProvision](omni.LabelMachineRequestSet),
		),
		qtransform.WithConcurrency(4),
	)
}
