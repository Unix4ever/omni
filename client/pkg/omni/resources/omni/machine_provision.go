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

// NewMachineProvision creates new MachineProvision state.
func NewMachineProvision(ns, id string) *MachineProvision {
	return typed.NewResource[MachineProvisionSpec, MachineProvisionExtension](
		resource.NewMetadata(ns, MachineProvisionType, id, resource.VersionUndefined),
		protobuf.NewResourceSpec(&specs.MachineProvisionSpec{}),
	)
}

// MachineProvisionType is the type of MachineProvision resource.
//
// tsgen:MachineProvisionType
const MachineProvisionType = resource.Type("MachineProvisions.omni.sidero.dev")

// MachineProvision resource is used to create an automatically scaled machine request set.
// It describes the config for the machine request and the config for the controller.
type MachineProvision = typed.Resource[MachineProvisionSpec, MachineProvisionExtension]

// MachineProvisionSpec wraps specs.MachineProvisionSpec.
type MachineProvisionSpec = protobuf.ResourceSpec[specs.MachineProvisionSpec, *specs.MachineProvisionSpec]

// MachineProvisionExtension providers auxiliary methods for MachineProvision resource.
type MachineProvisionExtension struct{}

// ResourceDefinition implements [typed.Extension] interface.
func (MachineProvisionExtension) ResourceDefinition() meta.ResourceDefinitionSpec {
	return meta.ResourceDefinitionSpec{
		Type:             MachineProvisionType,
		Aliases:          []resource.Type{},
		DefaultNamespace: resources.DefaultNamespace,
		PrintColumns: []meta.PrintColumn{
			{
				Name:     "ProviderID",
				JSONPath: "{.providerid}",
			},
		},
	}
}
