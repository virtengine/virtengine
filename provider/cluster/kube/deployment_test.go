package kube

import (
	"context"
	"fmt"
	"github.com/virtengine/virtengine/testutil"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/virtengine/virtengine/sdl"
	mtypes "github.com/virtengine/virtengine/x/market/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

const (
	randDSeq uint64 = 1
	randGSeq uint32 = 2
	randOSeq uint32 = 3
)

func TestDeploy(t *testing.T) {
	t.Skip()
	ctx := context.Background()

	owner := ed25519.GenPrivKey().PubKey().Address()
	provider := ed25519.GenPrivKey().PubKey().Address()

	leaseID := mtypes.LeaseID{
		Owner:    sdk.AccAddress(owner).String(),
		DSeq:     randDSeq,
		GSeq:     randGSeq,
		OSeq:     randOSeq,
		Provider: sdk.AccAddress(provider).String(),
	}

	sdl, err := sdl.ReadFile("../../../_run/kube/deployment.yaml")
	require.NoError(t, err)

	mani, err := sdl.Manifest()
	require.NoError(t, err)

	log := testutil.Logger(t)
	client, err := NewClient(log, "lease", NewDefaultSettings())
	assert.NoError(t, err)

	err = client.Deploy(ctx, leaseID, &mani.GetGroups()[0])
	assert.NoError(t, err)
}

func TestDeploySetsEnvironmentVariables(t *testing.T) {
	log := testutil.Logger(t)
	const fakeHostname = "ahostname.dev"
	settings := Settings{
		ClusterPublicHostname: fakeHostname,
	}
	lid := testutil.LeaseID(t)
	sdl, err := sdl.ReadFile("../../../_run/kube/deployment.yaml")
	require.NoError(t, err)

	mani, err := sdl.Manifest()
	require.NoError(t, err)
	service := mani.GetGroups()[0].Services[0]
	deploymentBuilder := newDeploymentBuilder(log, settings, lid, &mani.GetGroups()[0], &service)
	require.NotNil(t, deploymentBuilder)

	container := deploymentBuilder.container()
	require.NotNil(t, container)

	env := make(map[string]string)
	for _, entry := range container.Env {
		env[entry.Name] = entry.Value
	}

	value, ok := env[envVarVirtEngineClusterPublicHostname]
	require.True(t, ok)
	require.Equal(t, fakeHostname, value)

	value, ok = env[envVarVirtEngineDeploymentSequence]
	require.True(t, ok)
	require.Equal(t, fmt.Sprintf("%d", lid.GetDSeq()), value)

	value, ok = env[envVarVirtEngineGroupSequence]
	require.True(t, ok)
	require.Equal(t, fmt.Sprintf("%d", lid.GetGSeq()), value)

	value, ok = env[envVarVirtEngineOrderSequence]
	require.True(t, ok)
	require.Equal(t, fmt.Sprintf("%d", lid.GetOSeq()), value)

	value, ok = env[envVarVirtEngineOwner]
	require.True(t, ok)
	require.Equal(t, lid.Owner, value)

	value, ok = env[envVarVirtEngineProvider]
	require.True(t, ok)
	require.Equal(t, lid.Provider, value)

}
