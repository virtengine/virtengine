package keeper

import (
	"context"
	"fmt"
	"math"
	"strings"
	"text/template"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/hpc/types"
)

// CreateWorkloadTemplate handles creation of a workload template.
func (ms msgServer) CreateWorkloadTemplate(goCtx context.Context, msg *types.MsgCreateWorkloadTemplate) (*types.MsgCreateWorkloadTemplateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg == nil || msg.Template == nil {
		return nil, types.ErrInvalidWorkloadTemplate.Wrap("template is required")
	}

	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, types.ErrInvalidWorkloadTemplate.Wrap("invalid creator address")
	}

	template := *msg.Template
	if template.Publisher == "" {
		template.Publisher = creator.String()
	}
	if template.Publisher != creator.String() {
		return nil, types.ErrUnauthorized.Wrap("publisher must match creator")
	}

	if err := ms.keeper.CreateWorkloadTemplate(ctx, &template); err != nil {
		return nil, err
	}

	return &types.MsgCreateWorkloadTemplateResponse{TemplateID: template.TemplateID}, nil
}

// UpdateWorkloadTemplate handles updating a workload template.
func (ms msgServer) UpdateWorkloadTemplate(goCtx context.Context, msg *types.MsgUpdateWorkloadTemplate) (*types.MsgUpdateWorkloadTemplateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg == nil || msg.Template == nil {
		return nil, types.ErrInvalidWorkloadTemplate.Wrap("template is required")
	}

	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, types.ErrInvalidWorkloadTemplate.Wrap("invalid creator address")
	}

	template := *msg.Template
	if template.Publisher == "" {
		template.Publisher = creator.String()
	}
	if template.Publisher != creator.String() {
		return nil, types.ErrUnauthorized.Wrap("publisher must match creator")
	}

	if err := ms.keeper.UpdateWorkloadTemplate(ctx, &template); err != nil {
		return nil, err
	}

	return &types.MsgUpdateWorkloadTemplateResponse{}, nil
}

// ApproveWorkloadTemplate handles approving a workload template.
func (ms msgServer) ApproveWorkloadTemplate(goCtx context.Context, msg *types.MsgApproveWorkloadTemplate) (*types.MsgApproveWorkloadTemplateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, types.ErrUnauthorized.Wrap("invalid authority address")
	}

	if err := ms.keeper.ApproveWorkloadTemplate(ctx, msg.TemplateID, msg.Version, authority); err != nil {
		return nil, err
	}

	return &types.MsgApproveWorkloadTemplateResponse{}, nil
}

// RejectWorkloadTemplate handles rejecting a workload template.
func (ms msgServer) RejectWorkloadTemplate(goCtx context.Context, msg *types.MsgRejectWorkloadTemplate) (*types.MsgRejectWorkloadTemplateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, types.ErrUnauthorized.Wrap("invalid authority address")
	}

	if err := ms.keeper.RejectWorkloadTemplate(ctx, msg.TemplateID, msg.Version, msg.Reason, authority); err != nil {
		return nil, err
	}

	return &types.MsgRejectWorkloadTemplateResponse{}, nil
}

// DeprecateWorkloadTemplate handles deprecating a workload template.
func (ms msgServer) DeprecateWorkloadTemplate(goCtx context.Context, msg *types.MsgDeprecateWorkloadTemplate) (*types.MsgDeprecateWorkloadTemplateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, types.ErrUnauthorized.Wrap("invalid authority address")
	}

	if err := ms.keeper.DeprecateWorkloadTemplate(ctx, msg.TemplateID, msg.Version, authority); err != nil {
		return nil, err
	}

	return &types.MsgDeprecateWorkloadTemplateResponse{}, nil
}

// RevokeWorkloadTemplate handles revoking a workload template.
func (ms msgServer) RevokeWorkloadTemplate(goCtx context.Context, msg *types.MsgRevokeWorkloadTemplate) (*types.MsgRevokeWorkloadTemplateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, types.ErrUnauthorized.Wrap("invalid authority address")
	}

	if err := ms.keeper.RevokeWorkloadTemplate(ctx, msg.TemplateID, msg.Version, msg.Reason, authority); err != nil {
		return nil, err
	}

	return &types.MsgRevokeWorkloadTemplateResponse{}, nil
}

// SubmitJobFromTemplate handles submitting a job from a template.
func (ms msgServer) SubmitJobFromTemplate(goCtx context.Context, msg *types.MsgSubmitJobFromTemplate) (*types.MsgSubmitJobFromTemplateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, types.ErrInvalidJob.Wrap("invalid creator address")
	}

	template, found := ms.keeper.GetWorkloadTemplateByVersion(ctx, msg.TemplateID, msg.Version)
	if !found {
		return nil, types.ErrWorkloadTemplateNotFound
	}
	if !template.ApprovalStatus.CanBeUsed() {
		return nil, types.ErrWorkloadTemplateNotApproved
	}

	offering, err := ms.selectOfferingForTemplate(ctx, &template)
	if err != nil {
		return nil, err
	}

	resolved, err := resolveTemplateResources(&template, msg)
	if err != nil {
		return nil, err
	}

	workloadSpec, err := resolveTemplateWorkloadSpec(&template, msg.Parameters)
	if err != nil {
		return nil, err
	}

	job := &types.HPCJob{
		CustomerAddress: creator.String(),
		OfferingID:      offering.OfferingID,
		QueueName:       selectQueueName(&offering),
		WorkloadSpec:    workloadSpec,
		Resources: types.JobResources{
			Nodes:           resolved.Nodes,
			CPUCoresPerNode: resolved.CPUsPerNode,
			MemoryGBPerNode: resolved.MemoryGBPerNode,
			GPUsPerNode:     resolved.GPUsPerNode,
			GPUType:         resolved.GPUType,
			StorageGB:       resolved.StorageGB,
		},
		MaxRuntimeSeconds: resolved.RuntimeMinutes * 60,
		State:             types.JobStatePending,
	}

	if err := ms.keeper.SubmitJob(ctx, job); err != nil {
		return nil, err
	}

	// Emit template usage event.
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"workload_template_used",
			sdk.NewAttribute("template_id", template.TemplateID),
			sdk.NewAttribute("version", template.Version),
			sdk.NewAttribute("job_id", job.JobID),
			sdk.NewAttribute("creator", msg.Creator),
		),
	)

	return &types.MsgSubmitJobFromTemplateResponse{JobId: job.JobID}, nil
}

func (ms msgServer) selectOfferingForTemplate(ctx sdk.Context, template *types.WorkloadTemplate) (types.HPCOffering, error) {
	var selected *types.HPCOffering

	ms.keeper.WithOfferings(ctx, func(offering types.HPCOffering) bool {
		if !offering.Active {
			return false
		}

		if offering.SupportsCustomWorkloads {
			selected = &offering
			return true
		}

		for _, wl := range offering.PreconfiguredWorkloads {
			if wl.WorkloadID == template.TemplateID && (wl.Version == "" || wl.Version == template.Version) {
				selected = &offering
				return true
			}
		}

		return false
	})

	if selected == nil {
		return types.HPCOffering{}, types.ErrOfferingNotFound.Wrap("no offering supports this template")
	}

	return *selected, nil
}

type resolvedTemplateResources struct {
	Nodes           int32
	CPUsPerNode     int32
	MemoryMBPerNode int64
	MemoryGBPerNode int32
	GPUsPerNode     int32
	GPUType         string
	RuntimeMinutes  int64
	StorageGB       int32
}

func resolveTemplateResources(template *types.WorkloadTemplate, msg *types.MsgSubmitJobFromTemplate) (*resolvedTemplateResources, error) {
	res := &resolvedTemplateResources{
		Nodes:           template.Resources.DefaultNodes,
		CPUsPerNode:     template.Resources.DefaultCPUsPerNode,
		MemoryMBPerNode: template.Resources.DefaultMemoryMBPerNode,
		GPUsPerNode:     template.Resources.DefaultGPUsPerNode,
		RuntimeMinutes:  template.Resources.DefaultRuntimeMinutes,
		StorageGB:       template.Resources.StorageGBRequired,
	}

	if msg != nil {
		if msg.ResourceOverrides != nil {
			applyOverrides(res, msg.ResourceOverrides)
		} else {
			applyLegacyOverrides(res, msg)
		}
	}

	if err := validateResourceOverrides(template, res); err != nil {
		return nil, err
	}

	// Convert MB to GB for job resources (rounded up).
	if res.MemoryMBPerNode > 0 {
		resGB := (res.MemoryMBPerNode + 1023) / 1024
		if resGB > math.MaxInt32 {
			return nil, types.ErrInvalidWorkloadResources.Wrap("memory override too large")
		}
		res.MemoryGBPerNode = int32(resGB) //nolint:gosec // resGB bounded by MaxInt32 above
		res.MemoryMBPerNode = resGB * 1024
	}

	if res.RuntimeMinutes < 1 {
		return nil, types.ErrInvalidJob.Wrap("runtime_minutes must be >= 1")
	}

	if res.Nodes < 1 {
		return nil, types.ErrInvalidJob.Wrap("nodes must be >= 1")
	}

	return res, nil
}

func applyOverrides(res *resolvedTemplateResources, overrides *types.ResourceOverrides) {
	if overrides.Nodes > 0 {
		res.Nodes = overrides.Nodes
	}
	if overrides.CpusPerNode > 0 {
		res.CPUsPerNode = overrides.CpusPerNode
	}
	if overrides.MemoryMbPerNode > 0 {
		res.MemoryMBPerNode = overrides.MemoryMbPerNode
	}
	if overrides.GpusPerNode > 0 {
		res.GPUsPerNode = overrides.GpusPerNode
	}
	if overrides.RuntimeMinutes > 0 {
		res.RuntimeMinutes = overrides.RuntimeMinutes
	}
}

func applyLegacyOverrides(res *resolvedTemplateResources, msg *types.MsgSubmitJobFromTemplate) {
	if msg.Nodes > 0 {
		res.Nodes = msg.Nodes
	}
	if msg.CPUs > 0 {
		res.CPUsPerNode = msg.CPUs
	}
	if msg.MemoryMB > 0 {
		res.MemoryMBPerNode = msg.MemoryMB
	}
	if msg.Runtime > 0 {
		res.RuntimeMinutes = msg.Runtime
	}
}

func validateResourceOverrides(template *types.WorkloadTemplate, res *resolvedTemplateResources) error {
	r := template.Resources

	if res.Nodes < r.MinNodes || res.Nodes > r.MaxNodes {
		return types.ErrInvalidWorkloadResources.Wrap("nodes override out of bounds")
	}
	if res.CPUsPerNode < r.MinCPUsPerNode || res.CPUsPerNode > r.MaxCPUsPerNode {
		return types.ErrInvalidWorkloadResources.Wrap("cpus_per_node override out of bounds")
	}
	if res.MemoryMBPerNode < r.MinMemoryMBPerNode || res.MemoryMBPerNode > r.MaxMemoryMBPerNode {
		return types.ErrInvalidWorkloadResources.Wrap("memory_mb_per_node override out of bounds")
	}
	if res.GPUsPerNode < r.MinGPUsPerNode || res.GPUsPerNode > r.MaxGPUsPerNode {
		return types.ErrInvalidWorkloadResources.Wrap("gpus_per_node override out of bounds")
	}
	if res.RuntimeMinutes < r.MinRuntimeMinutes || res.RuntimeMinutes > r.MaxRuntimeMinutes {
		return types.ErrInvalidWorkloadResources.Wrap("runtime_minutes override out of bounds")
	}

	if len(r.GPUTypes) > 0 {
		res.GPUType = r.GPUTypes[0]
	}

	return nil
}

func resolveTemplateWorkloadSpec(template *types.WorkloadTemplate, params map[string]string) (types.JobWorkloadSpec, error) {
	env := make(map[string]string)

	for _, variable := range template.Environment {
		value := variable.Value
		if value == "" && variable.ValueTemplate != "" {
			substituted, err := applyTemplate(variable.ValueTemplate, params)
			if err != nil {
				return types.JobWorkloadSpec{}, err
			}
			value = substituted
		}
		if value != "" {
			env[variable.Name] = value
		}
	}

	for k, v := range params {
		key := "VE_PARAM_" + strings.ToUpper(strings.ReplaceAll(k, "-", "_"))
		if _, exists := env[key]; !exists {
			env[key] = v
		}
	}

	args := make([]string, 0, len(template.Entrypoint.DefaultArgs))
	args = append(args, template.Entrypoint.DefaultArgs...)
	if template.Entrypoint.ArgTemplate != "" {
		substituted, err := applyTemplate(template.Entrypoint.ArgTemplate, params)
		if err != nil {
			return types.JobWorkloadSpec{}, err
		}
		if strings.TrimSpace(substituted) != "" {
			args = append(args, strings.Fields(substituted)...)
		}
	}

	return types.JobWorkloadSpec{
		ContainerImage:          template.Runtime.ContainerImage,
		Command:                 template.Entrypoint.Command,
		Arguments:               args,
		Environment:             env,
		WorkingDirectory:        template.Entrypoint.WorkingDirectory,
		PreconfiguredWorkloadID: formatTemplateVersionedID(template.TemplateID, template.Version),
		IsPreconfigured:         true,
	}, nil
}

func applyTemplate(templateStr string, params map[string]string) (string, error) {
	if templateStr == "" || len(params) == 0 {
		return templateStr, nil
	}
	tmpl, err := template.New("subst").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}
	var sb strings.Builder
	if err := tmpl.Execute(&sb, params); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	return sb.String(), nil
}

func selectQueueName(offering *types.HPCOffering) string {
	if offering == nil {
		return ""
	}
	if len(offering.QueueOptions) > 0 {
		if offering.QueueOptions[0].PartitionName != "" {
			return offering.QueueOptions[0].PartitionName
		}
		if offering.QueueOptions[0].DisplayName != "" {
			return offering.QueueOptions[0].DisplayName
		}
	}
	return ""
}

func formatTemplateVersionedID(templateID, version string) string {
	if templateID == "" {
		return ""
	}
	if version == "" {
		return templateID
	}
	return templateID + "@" + version
}
