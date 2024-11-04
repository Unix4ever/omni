// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package models

import (
	"errors"
	"fmt"

	"github.com/cosi-project/runtime/pkg/resource"
	"github.com/hashicorp/go-multierror"

	"github.com/siderolabs/omni/client/pkg/omni/resources/omni"
)

// KindControlPlane is ControlPlane model kind.
const KindControlPlane = "ControlPlane"

// ControlPlane describes Cluster controlplane nodes.
type ControlPlane struct {
	MachineSet `yaml:",inline"`
	Managed    Managed `yaml:"managed,omitempty"`
}

// Managed describes managed control planes mode.
type Managed struct {
	Enable bool `yaml:"enable"`
}

// Validate the model.
func (controlplane *ControlPlane) Validate() error {
	var multiErr error

	if controlplane.Name != "" {
		multiErr = multierror.Append(multiErr, fmt.Errorf("custom name is not allowed in the controlplane"))
	}

	if controlplane.Managed.Enable {
		if controlplane.MachineClass != nil {
			multiErr = multierror.Append(multiErr, errors.New("setting machine class is not allowed in managed mode"))
		}

		if controlplane.Machines != nil {
			multiErr = multierror.Append(multiErr, errors.New("setting machines is not allowed in managed mode"))
		}

		if controlplane.SystemExtensions.SystemExtensions != nil {
			multiErr = multierror.Append(multiErr, errors.New("setting system extensions is not allowed in managed mode"))
		}
	}

	if controlplane.BootstrapSpec != nil {
		if controlplane.BootstrapSpec.ClusterUUID == "" {
			multiErr = multierror.Append(multiErr, fmt.Errorf("clusterUUID is required in bootstrapSpec"))
		}

		if controlplane.BootstrapSpec.Snapshot == "" {
			multiErr = multierror.Append(multiErr, fmt.Errorf("snapshot is required in bootstrapSpec"))
		}
	}

	if controlplane.UpdateStrategy != nil {
		multiErr = multierror.Append(multiErr, fmt.Errorf("updateStrategy is not allowed in the controlplane"))
	}

	if controlplane.DeleteStrategy != nil {
		multiErr = multierror.Append(multiErr, fmt.Errorf("deleteStrategy is not allowed in the controlplane"))
	}

	multiErr = joinErrors(multiErr, controlplane.MachineSet.Validate(), controlplane.Machines.Validate(), controlplane.Patches.Validate())

	if multiErr != nil {
		return fmt.Errorf("controlplane is invalid: %w", multiErr)
	}

	return nil
}

// Translate the model.
func (controlplane *ControlPlane) Translate(ctx TranslateContext) ([]resource.Resource, error) {
	return controlplane.translate(ctx, omni.ControlPlanesIDSuffix, omni.LabelControlPlaneRole, controlplane.Managed)
}

func init() {
	register[ControlPlane](KindControlPlane)
}
