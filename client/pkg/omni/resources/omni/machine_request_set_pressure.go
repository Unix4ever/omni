// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package omni

import (
	"github.com/cosi-project/runtime/pkg/resource"
	"github.com/cosi-project/runtime/pkg/resource/meta"
	"github.com/cosi-project/runtime/pkg/resource/protobuf"
	"github.com/cosi-project/runtime/pkg/resource/typed"

	"github.com/siderolabs/omni/client/api/omni/specs"
	"github.com/siderolabs/omni/client/pkg/omni/resources"
)

// NewMachineRequestSetPressure creates new MachineRequestSetPressure.
func NewMachineRequestSetPressure(ns, id string) *MachineRequestSetPressure {
	return typed.NewResource[MachineRequestSetPressureSpec, MachineRequestSetPressureExtension](
		resource.NewMetadata(ns, MachineRequestSetPressureType, id, resource.VersionUndefined),
		protobuf.NewResourceSpec(&specs.MachineRequestSetPressureSpec{}),
	)
}

// MachineRequestSetPressureType is the type of MachineRequestSetPressure resource.
//
// tsgen:MachineRequestSetPressureType
const MachineRequestSetPressureType = resource.Type("MachineRequestSetPressurees.omni.sidero.dev")

// MachineRequestSetPressure resource describes the additional number of machines which might be required in the machine request set.
type MachineRequestSetPressure = typed.Resource[MachineRequestSetPressureSpec, MachineRequestSetPressureExtension]

// MachineRequestSetPressureSpec wraps specs.MachineRequestSetPressureSpec.
type MachineRequestSetPressureSpec = protobuf.ResourceSpec[specs.MachineRequestSetPressureSpec, *specs.MachineRequestSetPressureSpec]

// MachineRequestSetPressureExtension providers auxiliary methods for MachineRequestSetPressure resource.
type MachineRequestSetPressureExtension struct{}

// ResourceDefinition implements [typed.Extension] interface.
func (MachineRequestSetPressureExtension) ResourceDefinition() meta.ResourceDefinitionSpec {
	return meta.ResourceDefinitionSpec{
		Type:             MachineRequestSetPressureType,
		Aliases:          []resource.Type{},
		DefaultNamespace: resources.DefaultNamespace,
		PrintColumns:     []meta.PrintColumn{},
	}
}
