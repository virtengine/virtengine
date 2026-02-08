//go:build e2e.integration

package hpc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/app"
	sdktestutil "github.com/virtengine/virtengine/sdk/go/testutil"
	hpckeeper "github.com/virtengine/virtengine/x/hpc/keeper"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// TestWorkloadTemplateLifecycle validates template creation, approval, and usage.
func TestWorkloadTemplateLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	app := app.Setup(
		app.WithChainID("virtengine-hpc-integration-1"),
		app.WithGenesis(func(cdc codec.Codec) app.GenesisState {
			return app.GenesisStateWithValSet(cdc)
		}),
	)

	baseTime := time.Unix(1_700_000_000, 0).UTC()
	ctx := app.NewUncachedContext(false, cmtproto.Header{
		Height: 1,
		Time:   baseTime,
	})

	msgServer := hpckeeper.NewMsgServerImpl(app.Keepers.VirtEngine.HPC)
	queryServer := hpckeeper.NewQueryServerImpl(app.Keepers.VirtEngine.HPC)

	creator := sdktestutil.AccAddress(t)
	authority := app.Keepers.VirtEngine.HPC.GetAuthority() // x/gov module account

	// Step 1: Create a workload template
	template := &hpctypes.WorkloadTemplate{
		TemplateID:  "test-mpi-template",
		Name:        "MPI Test Template",
		Version:     "1.0.0",
		Description: "Test MPI workload template for integration testing",
		Type:        hpctypes.WorkloadTypeMPI,
		Runtime: hpctypes.WorkloadRuntime{
			RuntimeType:       "native",
			MPIImplementation: "openmpi",
			RequiredModules:   []string{"gcc/11.2.0", "openmpi/4.1.1"},
		},
		Resources: hpctypes.WorkloadResourceSpec{
			MinNodes:               1,
			MaxNodes:               10,
			DefaultNodes:           2,
			MinCPUsPerNode:         1,
			MaxCPUsPerNode:         64,
			DefaultCPUsPerNode:     16,
			MinMemoryMBPerNode:     2048,
			MaxMemoryMBPerNode:     131072,
			DefaultMemoryMBPerNode: 16384,
			MinRuntimeMinutes:      1,
			MaxRuntimeMinutes:      2880,
			DefaultRuntimeMinutes:  120,
			NetworkRequired:        true,
		},
		Security: hpctypes.WorkloadSecuritySpec{
			SandboxLevel:       "basic",
			AllowNetworkAccess: true,
			AllowHostMounts:    false,
		},
		Entrypoint: hpctypes.WorkloadEntrypoint{
			Command:     "/usr/bin/mpi-benchmark",
			DefaultArgs: []string{"--iterations", "1000"},
			UseMPIRun:   true,
			MPIRunArgs:  []string{"--bind-to", "core"},
		},
		Environment: []hpctypes.EnvironmentVariable{
			{
				Name:     "OMP_NUM_THREADS",
				Value:    "1",
				Required: false,
			},
		},
		ParameterSchema: []hpctypes.ParameterDefinition{
			{
				Name:        "iterations",
				Type:        "int",
				Description: "Number of iterations to run",
				Default:     "1000",
				Required:    false,
			},
		},
		Publisher:      creator.String(),
		ApprovalStatus: hpctypes.WorkloadApprovalPending,
		CreatedAt:      baseTime,
		UpdatedAt:      baseTime,
		BlockHeight:    1,
	}

	createMsg := hpctypes.NewMsgCreateWorkloadTemplate(creator.String(), template)
	_, err := msgServer.CreateWorkloadTemplate(ctx, createMsg)
	require.NoError(t, err, "failed to create workload template")

	// Step 2: Query the template (should exist but not be approved)
	getReq := &hpctypes.QueryGetWorkloadTemplateRequest{
		TemplateId: "test-mpi-template",
		Version:    "1.0.0",
	}
	getResp, err := queryServer.GetWorkloadTemplate(sdk.WrapSDKContext(ctx), getReq)
	require.NoError(t, err, "failed to query template")
	require.NotNil(t, getResp.Template, "template not found")
	require.Equal(t, hpctypes.WorkloadApprovalPending, getResp.Template.ApprovalStatus, "template should be pending")

	// Step 3: Approve the template (requires authority/governance)
	approveMsg := hpctypes.NewMsgApproveWorkloadTemplate(authority, "test-mpi-template", "1.0.0")
	_, err = msgServer.ApproveWorkloadTemplate(ctx, approveMsg)
	require.NoError(t, err, "failed to approve template")

	// Step 4: Query approved template
	getResp, err = queryServer.GetWorkloadTemplate(sdk.WrapSDKContext(ctx), getReq)
	require.NoError(t, err, "failed to query template after approval")
	require.Equal(t, hpctypes.WorkloadApprovalApproved, getResp.Template.ApprovalStatus, "template should be approved")
	require.NotNil(t, getResp.Template.ApprovedAt, "approved_at should be set")

	// Step 5: List approved templates
	listReq := &hpctypes.QueryListApprovedWorkloadTemplatesRequest{}
	listResp, err := queryServer.ListApprovedWorkloadTemplates(sdk.WrapSDKContext(ctx), listReq)
	require.NoError(t, err, "failed to list approved templates")
	require.GreaterOrEqual(t, len(listResp.Templates), 1, "should have at least one approved template")

	// Step 6: Submit a job using the template
	jobMsg := &hpctypes.MsgSubmitJobFromTemplate{
		Creator:    creator.String(),
		TemplateId: "test-mpi-template",
		Version:    "1.0.0",
		Parameters: map[string]string{
			"iterations": "2000",
		},
		ResourceOverrides: &hpctypes.ResourceOverrides{
			Nodes:           4,
			CpusPerNode:     32,
			MemoryMbPerNode: 32768,
			RuntimeMinutes:  180,
		},
	}
	jobResp, err := msgServer.SubmitJobFromTemplate(ctx, jobMsg)
	require.NoError(t, err, "failed to submit job from template")
	require.NotEmpty(t, jobResp.JobId, "job ID should not be empty")

	// Step 7: Query template usage statistics
	usageReq := &hpctypes.QueryWorkloadTemplateUsageRequest{
		TemplateId: "test-mpi-template",
		Version:    "1.0.0",
	}
	usageResp, err := queryServer.WorkloadTemplateUsage(sdk.WrapSDKContext(ctx), usageReq)
	require.NoError(t, err, "failed to query template usage")
	require.GreaterOrEqual(t, usageResp.TotalUses, uint64(1), "template should have at least 1 use")
}

// TestWorkloadTemplateVersioning validates template versioning and updates.
func TestWorkloadTemplateVersioning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	app := app.Setup(
		app.WithChainID("virtengine-hpc-integration-1"),
		app.WithGenesis(func(cdc codec.Codec) app.GenesisState {
			return app.GenesisStateWithValSet(cdc)
		}),
	)

	baseTime := time.Unix(1_700_000_000, 0).UTC()
	ctx := app.NewUncachedContext(false, cmtproto.Header{
		Height: 1,
		Time:   baseTime,
	})

	msgServer := hpckeeper.NewMsgServerImpl(app.Keepers.VirtEngine.HPC)
	queryServer := hpckeeper.NewQueryServerImpl(app.Keepers.VirtEngine.HPC)

	creator := sdktestutil.AccAddress(t)

	// Create v1.0.0
	template := &hpctypes.WorkloadTemplate{
		TemplateID:  "versioned-template",
		Name:        "Versioned Template",
		Version:     "1.0.0",
		Description: "Version 1.0.0",
		Type:        hpctypes.WorkloadTypeBatch,
		Runtime: hpctypes.WorkloadRuntime{
			RuntimeType: "native",
		},
		Resources: hpctypes.WorkloadResourceSpec{
			MinNodes:               1,
			MaxNodes:               5,
			DefaultNodes:           1,
			MinCPUsPerNode:         1,
			MaxCPUsPerNode:         16,
			DefaultCPUsPerNode:     4,
			MinMemoryMBPerNode:     1024,
			MaxMemoryMBPerNode:     32768,
			DefaultMemoryMBPerNode: 8192,
			MinRuntimeMinutes:      1,
			MaxRuntimeMinutes:      480,
			DefaultRuntimeMinutes:  60,
		},
		Security: hpctypes.WorkloadSecuritySpec{
			SandboxLevel:       "basic",
			AllowNetworkAccess: true,
		},
		Entrypoint: hpctypes.WorkloadEntrypoint{
			Command: "/usr/bin/processor",
		},
		Publisher:      creator.String(),
		ApprovalStatus: hpctypes.WorkloadApprovalPending,
		CreatedAt:      baseTime,
		UpdatedAt:      baseTime,
		BlockHeight:    1,
	}

	createMsg := hpctypes.NewMsgCreateWorkloadTemplate(creator.String(), template)
	_, err := msgServer.CreateWorkloadTemplate(ctx, createMsg)
	require.NoError(t, err, "failed to create v1.0.0")

	// Create v1.1.0 with enhanced resources
	template.Version = "1.1.0"
	template.Description = "Version 1.1.0 with more resources"
	template.Resources.MaxNodes = 10
	template.Resources.DefaultNodes = 2
	template.UpdatedAt = baseTime.Add(time.Hour)

	createMsg = hpctypes.NewMsgCreateWorkloadTemplate(creator.String(), template)
	_, err = msgServer.CreateWorkloadTemplate(ctx, createMsg)
	require.NoError(t, err, "failed to create v1.1.0")

	// Query both versions
	getV1Req := &hpctypes.QueryGetWorkloadTemplateRequest{
		TemplateId: "versioned-template",
		Version:    "1.0.0",
	}
	getV1Resp, err := queryServer.GetWorkloadTemplate(sdk.WrapSDKContext(ctx), getV1Req)
	require.NoError(t, err, "failed to get v1.0.0")
	require.Equal(t, "1.0.0", getV1Resp.Template.Version)
	require.Equal(t, int32(5), getV1Resp.Template.Resources.MaxNodes)

	getV2Req := &hpctypes.QueryGetWorkloadTemplateRequest{
		TemplateId: "versioned-template",
		Version:    "1.1.0",
	}
	getV2Resp, err := queryServer.GetWorkloadTemplate(sdk.WrapSDKContext(ctx), getV2Req)
	require.NoError(t, err, "failed to get v1.1.0")
	require.Equal(t, "1.1.0", getV2Resp.Template.Version)
	require.Equal(t, int32(10), getV2Resp.Template.Resources.MaxNodes)

	// List all versions of the template
	listReq := &hpctypes.QueryListWorkloadTemplatesRequest{
		TemplateId: "versioned-template",
	}
	listResp, err := queryServer.ListWorkloadTemplates(sdk.WrapSDKContext(ctx), listReq)
	require.NoError(t, err, "failed to list template versions")
	require.Len(t, listResp.Templates, 2, "should have 2 versions")
}

// TestWorkloadTemplateGovernance validates governance actions on templates.
func TestWorkloadTemplateGovernance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	app := app.Setup(
		app.WithChainID("virtengine-hpc-integration-1"),
		app.WithGenesis(func(cdc codec.Codec) app.GenesisState {
			return app.GenesisStateWithValSet(cdc)
		}),
	)

	baseTime := time.Unix(1_700_000_000, 0).UTC()
	ctx := app.NewUncachedContext(false, cmtproto.Header{
		Height: 1,
		Time:   baseTime,
	})

	msgServer := hpckeeper.NewMsgServerImpl(app.Keepers.VirtEngine.HPC)
	queryServer := hpckeeper.NewQueryServerImpl(app.Keepers.VirtEngine.HPC)

	creator := sdktestutil.AccAddress(t)
	authority := app.Keepers.VirtEngine.HPC.GetAuthority()

	// Create a template
	template := &hpctypes.WorkloadTemplate{
		TemplateID:  "governance-test",
		Name:        "Governance Test Template",
		Version:     "1.0.0",
		Description: "Template for governance testing",
		Type:        hpctypes.WorkloadTypeBatch,
		Runtime: hpctypes.WorkloadRuntime{
			RuntimeType: "native",
		},
		Resources: hpctypes.WorkloadResourceSpec{
			MinNodes:               1,
			MaxNodes:               5,
			DefaultNodes:           1,
			MinCPUsPerNode:         1,
			MaxCPUsPerNode:         16,
			DefaultCPUsPerNode:     4,
			MinMemoryMBPerNode:     1024,
			MaxMemoryMBPerNode:     32768,
			DefaultMemoryMBPerNode: 8192,
			MinRuntimeMinutes:      1,
			MaxRuntimeMinutes:      480,
			DefaultRuntimeMinutes:  60,
		},
		Security: hpctypes.WorkloadSecuritySpec{
			SandboxLevel:       "basic",
			AllowNetworkAccess: true,
		},
		Entrypoint: hpctypes.WorkloadEntrypoint{
			Command: "/usr/bin/test",
		},
		Publisher:      creator.String(),
		ApprovalStatus: hpctypes.WorkloadApprovalPending,
		CreatedAt:      baseTime,
		UpdatedAt:      baseTime,
		BlockHeight:    1,
	}

	createMsg := hpctypes.NewMsgCreateWorkloadTemplate(creator.String(), template)
	_, err := msgServer.CreateWorkloadTemplate(ctx, createMsg)
	require.NoError(t, err, "failed to create template")

	// Test approval
	approveMsg := hpctypes.NewMsgApproveWorkloadTemplate(authority, "governance-test", "1.0.0")
	_, err = msgServer.ApproveWorkloadTemplate(ctx, approveMsg)
	require.NoError(t, err, "failed to approve template")

	// Verify approval
	getReq := &hpctypes.QueryGetWorkloadTemplateRequest{
		TemplateId: "governance-test",
		Version:    "1.0.0",
	}
	getResp, err := queryServer.GetWorkloadTemplate(sdk.WrapSDKContext(ctx), getReq)
	require.NoError(t, err)
	require.Equal(t, hpctypes.WorkloadApprovalApproved, getResp.Template.ApprovalStatus)

	// Test deprecation
	deprecateMsg := hpctypes.NewMsgDeprecateWorkloadTemplate(authority, "governance-test", "1.0.0", "Superseded by v2.0")
	_, err = msgServer.DeprecateWorkloadTemplate(ctx, deprecateMsg)
	require.NoError(t, err, "failed to deprecate template")

	getResp, err = queryServer.GetWorkloadTemplate(sdk.WrapSDKContext(ctx), getReq)
	require.NoError(t, err)
	require.Equal(t, hpctypes.WorkloadApprovalDeprecated, getResp.Template.ApprovalStatus)

	// Create another template for revocation test
	template.TemplateID = "revoke-test"
	template.Version = "1.0.0"
	template.ApprovalStatus = hpctypes.WorkloadApprovalPending
	createMsg = hpctypes.NewMsgCreateWorkloadTemplate(creator.String(), template)
	_, err = msgServer.CreateWorkloadTemplate(ctx, createMsg)
	require.NoError(t, err, "failed to create revoke-test template")

	// Approve it first
	approveMsg = hpctypes.NewMsgApproveWorkloadTemplate(authority, "revoke-test", "1.0.0")
	_, err = msgServer.ApproveWorkloadTemplate(ctx, approveMsg)
	require.NoError(t, err, "failed to approve revoke-test template")

	// Test revocation
	revokeMsg := hpctypes.NewMsgRevokeWorkloadTemplate(authority, "revoke-test", "1.0.0", "Security vulnerability found")
	_, err = msgServer.RevokeWorkloadTemplate(ctx, revokeMsg)
	require.NoError(t, err, "failed to revoke template")

	getReq.TemplateId = "revoke-test"
	getResp, err = queryServer.GetWorkloadTemplate(sdk.WrapSDKContext(ctx), getReq)
	require.NoError(t, err)
	require.Equal(t, hpctypes.WorkloadApprovalRevoked, getResp.Template.ApprovalStatus)
}

// TestWorkloadTemplateSearch validates template search functionality.
func TestWorkloadTemplateSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	app := app.Setup(
		app.WithChainID("virtengine-hpc-integration-1"),
		app.WithGenesis(func(cdc codec.Codec) app.GenesisState {
			return app.GenesisStateWithValSet(cdc)
		}),
	)

	baseTime := time.Unix(1_700_000_000, 0).UTC()
	ctx := app.NewUncachedContext(false, cmtproto.Header{
		Height: 1,
		Time:   baseTime,
	})

	msgServer := hpckeeper.NewMsgServerImpl(app.Keepers.VirtEngine.HPC)
	queryServer := hpckeeper.NewQueryServerImpl(app.Keepers.VirtEngine.HPC)

	creator := sdktestutil.AccAddress(t)

	// Create multiple templates with different types and tags
	templates := []struct {
		id   string
		name string
		typ  hpctypes.WorkloadType
		tags []string
	}{
		{"mpi-template", "MPI Workload", hpctypes.WorkloadTypeMPI, []string{"mpi", "parallel", "hpc"}},
		{"gpu-template", "GPU Workload", hpctypes.WorkloadTypeGPU, []string{"gpu", "cuda", "ml"}},
		{"batch-template", "Batch Processing", hpctypes.WorkloadTypeBatch, []string{"batch", "processing"}},
	}

	for _, tmpl := range templates {
		template := &hpctypes.WorkloadTemplate{
			TemplateID:  tmpl.id,
			Name:        tmpl.name,
			Version:     "1.0.0",
			Description: "Test template for " + tmpl.name,
			Type:        tmpl.typ,
			Runtime: hpctypes.WorkloadRuntime{
				RuntimeType: "native",
			},
			Resources: hpctypes.WorkloadResourceSpec{
				MinNodes:               1,
				MaxNodes:               10,
				DefaultNodes:           2,
				MinCPUsPerNode:         1,
				MaxCPUsPerNode:         32,
				DefaultCPUsPerNode:     8,
				MinMemoryMBPerNode:     2048,
				MaxMemoryMBPerNode:     65536,
				DefaultMemoryMBPerNode: 8192,
				MinRuntimeMinutes:      1,
				MaxRuntimeMinutes:      1440,
				DefaultRuntimeMinutes:  60,
			},
			Security: hpctypes.WorkloadSecuritySpec{
				SandboxLevel:       "basic",
				AllowNetworkAccess: true,
			},
			Entrypoint: hpctypes.WorkloadEntrypoint{
				Command: "/usr/bin/app",
			},
			Tags:           tmpl.tags,
			Publisher:      creator.String(),
			ApprovalStatus: hpctypes.WorkloadApprovalPending,
			CreatedAt:      baseTime,
			UpdatedAt:      baseTime,
			BlockHeight:    1,
		}

		createMsg := hpctypes.NewMsgCreateWorkloadTemplate(creator.String(), template)
		_, err := msgServer.CreateWorkloadTemplate(ctx, createMsg)
		require.NoError(t, err, "failed to create template: %s", tmpl.id)
	}

	// Test list by type
	listByTypeReq := &hpctypes.QueryListWorkloadTemplatesByTypeRequest{
		Type: string(hpctypes.WorkloadTypeMPI),
	}
	listByTypeResp, err := queryServer.ListWorkloadTemplatesByType(sdk.WrapSDKContext(ctx), listByTypeReq)
	require.NoError(t, err, "failed to list by type")
	require.GreaterOrEqual(t, len(listByTypeResp.Templates), 1, "should have at least 1 MPI template")

	// Test list by publisher
	listByPubReq := &hpctypes.QueryListWorkloadTemplatesByPublisherRequest{
		Publisher: creator.String(),
	}
	listByPubResp, err := queryServer.ListWorkloadTemplatesByPublisher(sdk.WrapSDKContext(ctx), listByPubReq)
	require.NoError(t, err, "failed to list by publisher")
	require.GreaterOrEqual(t, len(listByPubResp.Templates), 3, "should have at least 3 templates from creator")

	// Test search by tag
	searchReq := &hpctypes.QuerySearchWorkloadTemplatesRequest{
		Query: "gpu",
	}
	searchResp, err := queryServer.SearchWorkloadTemplates(sdk.WrapSDKContext(ctx), searchReq)
	require.NoError(t, err, "failed to search templates")
	require.GreaterOrEqual(t, len(searchResp.Templates), 1, "should find at least 1 template with 'gpu' tag")
}
