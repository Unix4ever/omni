// Copyright (c) 2024 Sidero Labs, Inc.
//
// Use of this software is governed by the Business Source License
// included in the LICENSE file.

package omni_test

import (
	"testing"

	"github.com/cosi-project/runtime/pkg/resource"
	"github.com/cosi-project/runtime/pkg/resource/rtestutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/siderolabs/omni/client/api/omni/specs"
	"github.com/siderolabs/omni/client/pkg/omni/resources"
	"github.com/siderolabs/omni/client/pkg/omni/resources/omni"
	omnictrl "github.com/siderolabs/omni/internal/backend/runtime/omni/controllers/omni"
)

type ManagedControlPlaneSuite struct {
	OmniSuite
}

func (suite *ManagedControlPlaneSuite) TestReconcile() {
	suite.startRuntime()

	suite.Require().NoError(suite.runtime.RegisterQController(omnictrl.NewClusterMachineConfigController(nil, 8090)))
	suite.Require().NoError(suite.runtime.RegisterQController(omnictrl.NewManagedControlPlaneController(omnictrl.ProviderConfig{
		ID:   "talemu",
		Data: "{}",
	})))

	id := "ms"

	cluster := omni.NewCluster(resources.DefaultNamespace, "cluster")

	cluster.TypedSpec().Value.TalosVersion = "v1.8.2"

	machineSet := omni.NewMachineSet(resources.DefaultNamespace, id)
	machineSet.Metadata().Labels().Set(omni.LabelControlPlaneRole, "")
	machineSet.Metadata().Labels().Set(omni.LabelCluster, cluster.Metadata().ID())

	machineSet.TypedSpec().Value.Managed = &specs.MachineSetSpec_Managed{
		Enable: true,
	}

	suite.Require().NoError(suite.state.Create(suite.ctx, cluster))
	suite.Require().NoError(suite.state.Create(suite.ctx, machineSet))

	rtestutils.AssertResources(suite.ctx, suite.T(), suite.state, []string{id},
		func(machineRequestSet *omni.MachineRequestSet, assert *assert.Assertions) {
			assert.Equal(cluster.TypedSpec().Value.TalosVersion, machineRequestSet.TypedSpec().Value.TalosVersion)
			assert.EqualValues(3, machineRequestSet.TypedSpec().Value.MachineCount)
		},
	)

	rtestutils.Destroy[*omni.MachineSet](suite.ctx, suite.T(), suite.state, []resource.ID{id})

	rtestutils.AssertNoResource[*omni.MachineRequestSet](suite.ctx, suite.T(), suite.state, id)
}

func TestManagedControlPlaneSuite(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(ManagedControlPlaneSuite))
}
