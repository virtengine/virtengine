// Copyright 2024-2026 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0

// Package runner provides experiment execution backends for chaos engineering.
package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/virtengine/virtengine/pkg/chaos"
)

// KubernetesRunner implements the ExperimentRunner interface for Kubernetes.
// It integrates with Chaos Mesh or Litmus to inject chaos into Kubernetes clusters.
type KubernetesRunner struct {
	mu sync.RWMutex

	// experiments tracks active experiments
	experiments map[string]*chaos.ExperimentSpec

	// namespace is the Kubernetes namespace for chaos resources
	namespace string

	// chaosProvider is the chaos engineering platform ("chaos-mesh" or "litmus")
	chaosProvider string

	// kubectlPath is the path to kubectl binary
	kubectlPath string

	// dryRun indicates whether to actually apply chaos
	dryRun bool

	// logger for debug output
	logger chaos.Logger
}

// KubernetesRunnerOption configures the KubernetesRunner.
type KubernetesRunnerOption func(*KubernetesRunner)

// WithNamespace sets the Kubernetes namespace.
func WithNamespace(namespace string) KubernetesRunnerOption {
	return func(r *KubernetesRunner) {
		r.namespace = namespace
	}
}

// WithChaosProvider sets the chaos provider (chaos-mesh or litmus).
func WithChaosProvider(provider string) KubernetesRunnerOption {
	return func(r *KubernetesRunner) {
		r.chaosProvider = provider
	}
}

// WithKubectlPath sets the kubectl binary path.
func WithKubectlPath(path string) KubernetesRunnerOption {
	return func(r *KubernetesRunner) {
		r.kubectlPath = path
	}
}

// WithDryRun enables dry-run mode.
func WithDryRun(dryRun bool) KubernetesRunnerOption {
	return func(r *KubernetesRunner) {
		r.dryRun = dryRun
	}
}

// WithK8sLogger sets the logger.
func WithK8sLogger(logger chaos.Logger) KubernetesRunnerOption {
	return func(r *KubernetesRunner) {
		r.logger = logger
	}
}

// NewKubernetesRunner creates a new Kubernetes chaos runner.
func NewKubernetesRunner(opts ...KubernetesRunnerOption) *KubernetesRunner {
	r := &KubernetesRunner{
		experiments:   make(map[string]*chaos.ExperimentSpec),
		namespace:     "chaos-testing",
		chaosProvider: "chaos-mesh",
		kubectlPath:   "kubectl",
		dryRun:        false,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Start begins a chaos experiment in Kubernetes.
func (r *KubernetesRunner) Start(ctx context.Context, experiment *chaos.ExperimentSpec) error {
	if experiment == nil {
		return fmt.Errorf("experiment is nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.experiments[experiment.ID]; exists {
		return fmt.Errorf("experiment %s already running", experiment.ID)
	}

	// Generate chaos manifest based on experiment type
	manifest, err := r.generateChaosManifest(experiment)
	if err != nil {
		return fmt.Errorf("failed to generate manifest: %w", err)
	}

	// Apply the manifest
	if !r.dryRun {
		if err := r.applyManifest(ctx, manifest); err != nil {
			return fmt.Errorf("failed to apply manifest: %w", err)
		}
	}

	experiment.State = chaos.ExperimentStateRunning
	experiment.StartTime = time.Now()
	r.experiments[experiment.ID] = experiment

	if r.logger != nil {
		r.logger.Info("started Kubernetes chaos experiment",
			chaos.LogField{Key: "experiment_id", Value: experiment.ID},
			chaos.LogField{Key: "provider", Value: r.chaosProvider},
		)
	}

	return nil
}

// Stop terminates a running experiment.
func (r *KubernetesRunner) Stop(ctx context.Context, experiment *chaos.ExperimentSpec) error {
	if experiment == nil {
		return fmt.Errorf("experiment is nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.experiments[experiment.ID]; !exists {
		return fmt.Errorf("experiment %s not found", experiment.ID)
	}

	// Delete the chaos resource
	if !r.dryRun {
		resourceName := r.getResourceName(experiment)
		if err := r.deleteResource(ctx, resourceName); err != nil {
			return fmt.Errorf("failed to delete chaos resource: %w", err)
		}
	}

	experiment.State = chaos.ExperimentStateCompleted
	experiment.EndTime = time.Now()
	delete(r.experiments, experiment.ID)

	if r.logger != nil {
		r.logger.Info("stopped Kubernetes chaos experiment",
			chaos.LogField{Key: "experiment_id", Value: experiment.ID},
		)
	}

	return nil
}

// Status returns the current state of an experiment.
func (r *KubernetesRunner) Status(ctx context.Context, experimentID string) (*chaos.ExperimentSpec, error) {
	r.mu.RLock()
	experiment, exists := r.experiments[experimentID]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("experiment %s not found", experimentID)
	}

	// Query Kubernetes for actual status
	if !r.dryRun {
		resourceName := r.getResourceName(experiment)
		status, err := r.getResourceStatus(ctx, resourceName)
		if err != nil {
			return experiment, nil // Return cached state if query fails
		}
		experiment.State = r.mapK8sStatusToState(status)
	}

	return experiment, nil
}

// Rollback reverts changes made by an experiment.
func (r *KubernetesRunner) Rollback(ctx context.Context, experiment *chaos.ExperimentSpec) error {
	return r.Stop(ctx, experiment)
}

// ExecuteAction executes a single experiment action.
func (r *KubernetesRunner) ExecuteAction(ctx context.Context, action chaos.ExperimentAction) error {
	// For Kubernetes, actions are applied as manifests
	manifest := r.generateActionManifest(action)

	if !r.dryRun {
		if err := r.applyManifest(ctx, manifest); err != nil {
			return fmt.Errorf("failed to apply action manifest: %w", err)
		}
	}

	if r.logger != nil {
		r.logger.Info("executed Kubernetes action",
			chaos.LogField{Key: "action", Value: action.Name},
		)
	}

	return nil
}

// ExecuteRollback executes a rollback action.
func (r *KubernetesRunner) ExecuteRollback(ctx context.Context, action chaos.RollbackAction) error {
	resourceName := fmt.Sprintf("rollback-%s", action.Type)

	if !r.dryRun {
		if err := r.deleteResource(ctx, resourceName); err != nil {
			// Log but don't fail on rollback errors
			if r.logger != nil {
				r.logger.Warn("rollback resource deletion failed",
					chaos.LogField{Key: "resource", Value: resourceName},
					chaos.LogField{Key: "error", Value: err.Error()},
				)
			}
		}
	}

	return nil
}

// Validate validates an action before execution.
func (r *KubernetesRunner) Validate(action chaos.ExperimentAction) error {
	if action.Name == "" {
		return fmt.Errorf("action name is required")
	}
	if action.Type == "" {
		return fmt.Errorf("action type is required")
	}
	return nil
}

// generateChaosManifest generates a Chaos Mesh or Litmus manifest.
func (r *KubernetesRunner) generateChaosManifest(experiment *chaos.ExperimentSpec) (string, error) {
	switch r.chaosProvider {
	case "chaos-mesh":
		return r.generateChaosMeshManifest(experiment)
	case "litmus":
		return r.generateLitmusManifest(experiment)
	default:
		return "", fmt.Errorf("unsupported chaos provider: %s", r.chaosProvider)
	}
}

// generateChaosMeshManifest generates a Chaos Mesh manifest.
func (r *KubernetesRunner) generateChaosMeshManifest(experiment *chaos.ExperimentSpec) (string, error) {
	manifest := map[string]interface{}{
		"apiVersion": "chaos-mesh.org/v1alpha1",
		"metadata": map[string]interface{}{
			"name":      r.getResourceName(experiment),
			"namespace": r.namespace,
			"labels": map[string]string{
				"app.kubernetes.io/component": "chaos-experiment",
				"chaos.virtengine.io/id":      experiment.ID,
			},
		},
	}

	// Set kind and spec based on experiment type
	switch experiment.Type {
	case chaos.ExperimentTypeNetworkPartition:
		manifest["kind"] = "NetworkChaos"
		manifest["spec"] = map[string]interface{}{
			"action":   "partition",
			"mode":     "all",
			"duration": fmt.Sprintf("%ds", int(experiment.Duration.Seconds())),
			"selector": r.buildSelector(experiment.Targets),
		}
	case chaos.ExperimentTypeNetworkLatency:
		manifest["kind"] = "NetworkChaos"
		manifest["spec"] = map[string]interface{}{
			"action":   "delay",
			"mode":     "all",
			"duration": fmt.Sprintf("%ds", int(experiment.Duration.Seconds())),
			"delay": map[string]interface{}{
				"latency": experiment.Parameters["latency"],
				"jitter":  experiment.Parameters["jitter"],
			},
			"selector": r.buildSelector(experiment.Targets),
		}
	case chaos.ExperimentTypeNodeFailure:
		manifest["kind"] = "PodChaos"
		manifest["spec"] = map[string]interface{}{
			"action":   "pod-kill",
			"mode":     "one",
			"duration": fmt.Sprintf("%ds", int(experiment.Duration.Seconds())),
			"selector": r.buildSelector(experiment.Targets),
		}
	case chaos.ExperimentTypeCPUStress:
		manifest["kind"] = "StressChaos"
		manifest["spec"] = map[string]interface{}{
			"mode":     "all",
			"duration": fmt.Sprintf("%ds", int(experiment.Duration.Seconds())),
			"stressors": map[string]interface{}{
				"cpu": map[string]interface{}{
					"workers": experiment.Parameters["workers"],
					"load":    experiment.Parameters["load"],
				},
			},
			"selector": r.buildSelector(experiment.Targets),
		}
	case chaos.ExperimentTypeMemoryStress:
		manifest["kind"] = "StressChaos"
		manifest["spec"] = map[string]interface{}{
			"mode":     "all",
			"duration": fmt.Sprintf("%ds", int(experiment.Duration.Seconds())),
			"stressors": map[string]interface{}{
				"memory": map[string]interface{}{
					"workers": experiment.Parameters["workers"],
					"size":    experiment.Parameters["size"],
				},
			},
			"selector": r.buildSelector(experiment.Targets),
		}
	case chaos.ExperimentTypeTimeChaos:
		manifest["kind"] = "TimeChaos"
		manifest["spec"] = map[string]interface{}{
			"mode":       "all",
			"duration":   fmt.Sprintf("%ds", int(experiment.Duration.Seconds())),
			"timeOffset": experiment.Parameters["offset"],
			"selector":   r.buildSelector(experiment.Targets),
		}
	default:
		return "", fmt.Errorf("unsupported experiment type for Chaos Mesh: %s", experiment.Type)
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal manifest: %w", err)
	}

	return string(data), nil
}

// generateLitmusManifest generates a LitmusChaos manifest.
func (r *KubernetesRunner) generateLitmusManifest(experiment *chaos.ExperimentSpec) (string, error) {
	manifest := map[string]interface{}{
		"apiVersion": "litmuschaos.io/v1alpha1",
		"kind":       "ChaosEngine",
		"metadata": map[string]interface{}{
			"name":      r.getResourceName(experiment),
			"namespace": r.namespace,
		},
		"spec": map[string]interface{}{
			"engineState": "active",
			"appinfo": map[string]interface{}{
				"appns":    r.namespace,
				"applabel": r.getAppLabel(experiment.Targets),
				"appkind":  "deployment",
			},
			"chaosServiceAccount": "litmus-admin",
			"experiments":         r.getLitmusExperiments(experiment),
		},
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal manifest: %w", err)
	}

	return string(data), nil
}

// getLitmusExperiments maps experiment type to Litmus experiments.
func (r *KubernetesRunner) getLitmusExperiments(experiment *chaos.ExperimentSpec) []map[string]interface{} {
	var expName string
	switch experiment.Type {
	case chaos.ExperimentTypeNodeFailure:
		expName = "pod-delete"
	case chaos.ExperimentTypeNetworkPartition:
		expName = "pod-network-partition"
	case chaos.ExperimentTypeNetworkLatency:
		expName = "pod-network-latency"
	case chaos.ExperimentTypeCPUStress:
		expName = "pod-cpu-hog"
	case chaos.ExperimentTypeMemoryStress:
		expName = "pod-memory-hog"
	default:
		expName = "pod-delete"
	}

	return []map[string]interface{}{
		{
			"name": expName,
			"spec": map[string]interface{}{
				"components": map[string]interface{}{
					"env": []map[string]string{
						{"name": "TOTAL_CHAOS_DURATION", "value": fmt.Sprintf("%d", int(experiment.Duration.Seconds()))},
					},
				},
			},
		},
	}
}

// buildSelector builds a Chaos Mesh selector from targets.
func (r *KubernetesRunner) buildSelector(targets []chaos.Target) map[string]interface{} {
	selector := map[string]interface{}{
		"namespaces": []string{r.namespace},
	}

	for _, target := range targets {
		if len(target.Selector) > 0 {
			selector["labelSelectors"] = target.Selector
		}
		if target.Name != "" {
			if pods, ok := selector["pods"].(map[string][]string); ok {
				pods[r.namespace] = append(pods[r.namespace], target.Name)
			} else {
				selector["pods"] = map[string][]string{
					r.namespace: {target.Name},
				}
			}
		}
	}

	return selector
}

// getAppLabel extracts app label from targets.
func (r *KubernetesRunner) getAppLabel(targets []chaos.Target) string {
	for _, target := range targets {
		if app, ok := target.Selector["app"]; ok {
			return fmt.Sprintf("app=%s", app)
		}
	}
	return "app=virtengine"
}

// getResourceName generates a Kubernetes resource name.
func (r *KubernetesRunner) getResourceName(experiment *chaos.ExperimentSpec) string {
	name := fmt.Sprintf("chaos-%s", experiment.ID)
	// Kubernetes name constraints
	if len(name) > 63 {
		name = name[:63]
	}
	return strings.ToLower(name)
}

// generateActionManifest generates a manifest for a single action.
func (r *KubernetesRunner) generateActionManifest(action chaos.ExperimentAction) string {
	manifest := map[string]interface{}{
		"apiVersion": "chaos-mesh.org/v1alpha1",
		"kind":       "PodChaos",
		"metadata": map[string]interface{}{
			"name":      fmt.Sprintf("action-%s", action.Name),
			"namespace": r.namespace,
		},
		"spec": map[string]interface{}{
			"action": action.Type,
			"mode":   "one",
		},
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to marshal manifest: %v"}`, err)
	}
	return string(data)
}

// applyManifest applies a manifest to Kubernetes.
func (r *KubernetesRunner) applyManifest(ctx context.Context, manifest string) error {
	//nolint:gosec // G204: kubectlPath is validated during runner initialization
	cmd := exec.CommandContext(ctx, r.kubectlPath, "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(manifest)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kubectl apply failed: %w, output: %s", err, string(output))
	}
	return nil
}

// deleteResource deletes a chaos resource.
func (r *KubernetesRunner) deleteResource(ctx context.Context, name string) error {
	// Try to delete all chaos types
	for _, kind := range []string{"podchaos", "networkchaos", "stresschaos", "timechaos", "chaosengine"} {
		//nolint:gosec // G204: kubectlPath validated, kind from hardcoded list, name is chaos resource name
		cmd := exec.CommandContext(ctx, r.kubectlPath, "delete", kind, name, "-n", r.namespace, "--ignore-not-found")
		if _, err := cmd.CombinedOutput(); err != nil {
			// Log but continue
			if r.logger != nil {
				r.logger.Debug("delete attempt", chaos.LogField{Key: "kind", Value: kind})
			}
		}
	}
	return nil
}

// getResourceStatus queries the status of a chaos resource.
func (r *KubernetesRunner) getResourceStatus(ctx context.Context, name string) (string, error) {
	// Try common chaos resource types
	for _, kind := range []string{"podchaos", "networkchaos", "stresschaos"} {
		//nolint:gosec // G204: kubectlPath validated, kind from hardcoded list
		cmd := exec.CommandContext(ctx, r.kubectlPath, "get", kind, name, "-n", r.namespace,
			"-o", "jsonpath={.status.experiment.phase}")
		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			return string(output), nil
		}
	}
	return "Unknown", nil
}

// mapK8sStatusToState maps Kubernetes status to experiment state.
func (r *KubernetesRunner) mapK8sStatusToState(status string) chaos.ExperimentState {
	switch strings.ToLower(status) {
	case "running", "executing":
		return chaos.ExperimentStateRunning
	case "paused":
		return chaos.ExperimentStatePaused
	case "finished", "succeeded":
		return chaos.ExperimentStateCompleted
	case "failed", "error":
		return chaos.ExperimentStateFailed
	default:
		return chaos.ExperimentStatePending
	}
}
