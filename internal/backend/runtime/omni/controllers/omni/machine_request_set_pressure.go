// Copyright (c) 2024 Sidero Labs, Inc.
//
// Use of this software is governed by the Business Source License
// included in the LICENSE file.

package omni

import (
	"context"

	"github.com/cosi-project/runtime/pkg/controller"
	"github.com/cosi-project/runtime/pkg/controller/generic/qtransform"
	"github.com/cosi-project/runtime/pkg/resource"
	"github.com/cosi-project/runtime/pkg/safe"
	"github.com/cosi-project/runtime/pkg/state"
	"go.uber.org/zap"

	"github.com/siderolabs/omni/client/pkg/omni/resources/omni"
	"github.com/siderolabs/omni/internal/backend/runtime/omni/controllers/omni/internal/mappers"
)

const machineRequestSetPressureControllerName = "MachineRequestSetPressureController"

// MachineRequestSetPressureController manages MachineRequestSetPressure resource lifecycle.
//
// MachineRequestSetPressureController calculates requested machines for each machine request set.
type MachineRequestSetPressureController = qtransform.QController[*omni.MachineRequestSet, *omni.MachineRequestSetPressure]

// NewMachineRequestSetPressureController initializes MachineRequestSetPressureController.
func NewMachineRequestSetPressureController() *MachineRequestSetPressureController {
	return qtransform.NewQController(
		qtransform.Settings[*omni.MachineRequestSet, *omni.MachineRequestSetPressure]{
			Name: machineRequestSetPressureControllerName,
			MapMetadataFunc: func(res *omni.MachineRequestSet) *omni.MachineRequestSetPressure {
				return omni.NewMachineRequestSetPressure(res.Metadata().Namespace(), res.Metadata().ID())
			},
			UnmapMetadataFunc: func(res *omni.MachineRequestSetPressure) *omni.MachineRequestSet {
				return omni.NewMachineRequestSet(res.Metadata().Namespace(), res.Metadata().ID())
			},
			TransformExtraOutputFunc: func(ctx context.Context, r controller.ReaderWriter, _ *zap.Logger, mrs *omni.MachineRequestSet, mrsp *omni.MachineRequestSetPressure) error {
				msrmList, err := safe.ReaderListAll[*omni.MachineSetStatus](ctx, r, state.WithLabelQuery(resource.LabelEqual(omni.LabelMachineRequestSet, mrs.Metadata().ID())))
				if err != nil {
					return err
				}

				total := uint32(0)

				err = msrmList.ForEachErr(func(mss *omni.MachineSetStatus) error {
					if mss.Metadata().Phase() == resource.PhaseTearingDown || mrs.Metadata().Phase() == resource.PhaseTearingDown {
						return r.RemoveFinalizer(ctx, mss.Metadata(), machineRequestSetPressureControllerName)
					}

					if mss.TypedSpec().Value.MachineRequestSet == nil {
						return nil
					}

					total += mss.TypedSpec().Value.MachineRequestSet.MachineCount

					if !mss.Metadata().Finalizers().Has(machineRequestSetPressureControllerName) {
						return r.AddFinalizer(ctx, mss.Metadata(), machineRequestSetPressureControllerName)
					}

					return nil
				})
				if err != nil {
					return err
				}

				mrsp.TypedSpec().Value.RequiredMachines = total

				return nil
			},
		},
		qtransform.WithExtraMappedInput(
			mappers.MapExtractLabelValue[*omni.MachineSetStatus, *omni.MachineRequestSet](omni.LabelMachineRequestSet),
		),
	)
}