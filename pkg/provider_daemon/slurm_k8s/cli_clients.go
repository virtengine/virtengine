// Package slurm_k8s implements SLURM cluster deployment on Kubernetes.
package slurm_k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/virtengine/virtengine/pkg/security"
	"gopkg.in/yaml.v3"
)

// HelmCLIConfig configures Helm CLI usage.
type HelmCLIConfig struct {
	Binary     string   `json:"binary" yaml:"binary"`
	Kubeconfig string   `json:"kubeconfig" yaml:"kubeconfig"`
	ExtraArgs  []string `json:"extra_args" yaml:"extra_args"`
}

// HelmCLIClient implements HelmClient using the Helm CLI.
type HelmCLIClient struct {
	config HelmCLIConfig
}

// NewHelmCLIClient creates a new Helm CLI client.
func NewHelmCLIClient(config HelmCLIConfig) *HelmCLIClient {
	if config.Binary == "" {
		config.Binary = "helm"
	}

	// Validate helm executable path
	if validatedPath, err := security.ResolveAndValidateExecutable("kubernetes", config.Binary); err == nil {
		config.Binary = validatedPath
	}

	return &HelmCLIClient{config: config}
}

// Install installs a Helm chart.
func (h *HelmCLIClient) Install(ctx context.Context, releaseName, chartPath, namespace string, values map[string]interface{}) error {
	args := []string{"upgrade", "--install", releaseName, chartPath, "--namespace", namespace, "--create-namespace"}
	return h.runWithValues(ctx, args, values)
}

// Upgrade upgrades a Helm release.
func (h *HelmCLIClient) Upgrade(ctx context.Context, releaseName, chartPath, namespace string, values map[string]interface{}) error {
	args := []string{"upgrade", releaseName, chartPath, "--namespace", namespace}
	return h.runWithValues(ctx, args, values)
}

// Uninstall uninstalls a Helm release.
func (h *HelmCLIClient) Uninstall(ctx context.Context, releaseName, namespace string) error {
	args := []string{"uninstall", releaseName, "--namespace", namespace}
	_, err := h.run(ctx, args)
	return err
}

// GetRelease gets a Helm release.
func (h *HelmCLIClient) GetRelease(ctx context.Context, releaseName, namespace string) (*HelmRelease, error) {
	args := []string{"status", releaseName, "--namespace", namespace, "--output", "json"}
	out, err := h.run(ctx, args)
	if err != nil {
		return nil, err
	}

	var status struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Chart     string `json:"chart"`
		AppVer    string `json:"app_version"`
		Info      struct {
			Status string `json:"status"`
		} `json:"info"`
	}
	if err := json.Unmarshal(out, &status); err != nil {
		return nil, fmt.Errorf("decode helm status: %w", err)
	}

	return &HelmRelease{
		Name:       status.Name,
		Namespace:  status.Namespace,
		Chart:      status.Chart,
		AppVersion: status.AppVer,
		Status:     status.Info.Status,
	}, nil
}

// ListReleases lists Helm releases in a namespace.
func (h *HelmCLIClient) ListReleases(ctx context.Context, namespace string) ([]*HelmRelease, error) {
	args := []string{"list", "--namespace", namespace, "--output", "json"}
	out, err := h.run(ctx, args)
	if err != nil {
		return nil, err
	}

	var releases []struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Chart     string `json:"chart"`
		AppVer    string `json:"app_version"`
		Status    string `json:"status"`
	}
	if err := json.Unmarshal(out, &releases); err != nil {
		return nil, fmt.Errorf("decode helm list: %w", err)
	}

	result := make([]*HelmRelease, 0, len(releases))
	for _, rel := range releases {
		result = append(result, &HelmRelease{
			Name:       rel.Name,
			Namespace:  rel.Namespace,
			Chart:      rel.Chart,
			AppVersion: rel.AppVer,
			Status:     rel.Status,
		})
	}
	return result, nil
}

func (h *HelmCLIClient) runWithValues(ctx context.Context, args []string, values map[string]interface{}) error {
	if len(values) == 0 {
		_, err := h.run(ctx, args)
		return err
	}

	data, err := yaml.Marshal(values)
	if err != nil {
		return fmt.Errorf("marshal helm values: %w", err)
	}

	dir, err := os.MkdirTemp("", "slurm-helm-values-*")
	if err != nil {
		return fmt.Errorf("create helm temp dir: %w", err)
	}
	defer os.RemoveAll(dir)

	valuesPath := filepath.Join(dir, "values.yaml")
	if err := os.WriteFile(valuesPath, data, 0o600); err != nil {
		return fmt.Errorf("write helm values: %w", err)
	}

	args = append(args, "--values", valuesPath)
	_, err = h.run(ctx, args)
	return err
}

func (h *HelmCLIClient) run(ctx context.Context, args []string) ([]byte, error) {
	args = append(args, h.config.ExtraArgs...)

	// Validate command using security package
	validator := security.NewCommandValidator(
		[]string{"helm"},
		5*time.Minute,
	)
	if err := validator.ValidateCommand(h.config.Binary, args...); err != nil {
		return nil, fmt.Errorf("command validation failed: %w", err)
	}

	//nolint:gosec // G204: Command and arguments validated by security.CommandValidator
	cmd := exec.CommandContext(ctx, h.config.Binary, args...)
	if h.config.Kubeconfig != "" {
		cmd.Env = append(os.Environ(), "KUBECONFIG="+h.config.Kubeconfig)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("helm %s failed: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}

	return stdout.Bytes(), nil
}

// KubeCLIConfig configures kubectl usage.
type KubeCLIConfig struct {
	Binary     string   `json:"binary" yaml:"binary"`
	Kubeconfig string   `json:"kubeconfig" yaml:"kubeconfig"`
	ExtraArgs  []string `json:"extra_args" yaml:"extra_args"`
}

// KubeCLIStatusChecker implements KubernetesStatusChecker using kubectl.
type KubeCLIStatusChecker struct {
	config KubeCLIConfig
}

// NewKubeCLIStatusChecker creates a new kubectl-based status checker.
func NewKubeCLIStatusChecker(config KubeCLIConfig) *KubeCLIStatusChecker {
	if config.Binary == "" {
		config.Binary = "kubectl"
	}

	// Validate kubectl executable path
	if validatedPath, err := security.ResolveAndValidateExecutable("kubernetes", config.Binary); err == nil {
		config.Binary = validatedPath
	}

	return &KubeCLIStatusChecker{config: config}
}

// GetStatefulSetStatus gets StatefulSet status.
func (k *KubeCLIStatusChecker) GetStatefulSetStatus(ctx context.Context, namespace, name string) (*StatefulSetStatus, error) {
	args := []string{"get", "statefulset", name, "--namespace", namespace, "--output", "json"}
	out, err := k.run(ctx, args)
	if err != nil {
		return nil, err
	}

	var payload struct {
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
		Spec struct {
			Replicas int32 `json:"replicas"`
		} `json:"spec"`
		Status struct {
			ReadyReplicas   int32 `json:"readyReplicas"`
			CurrentReplicas int32 `json:"currentReplicas"`
			UpdatedReplicas int32 `json:"updatedReplicas"`
		} `json:"status"`
	}

	if err := json.Unmarshal(out, &payload); err != nil {
		return nil, fmt.Errorf("decode statefulset status: %w", err)
	}

	return &StatefulSetStatus{
		Name:            payload.Metadata.Name,
		Replicas:        payload.Spec.Replicas,
		ReadyReplicas:   payload.Status.ReadyReplicas,
		CurrentReplicas: payload.Status.CurrentReplicas,
		UpdatedReplicas: payload.Status.UpdatedReplicas,
	}, nil
}

// GetPodLogs gets pod logs.
func (k *KubeCLIStatusChecker) GetPodLogs(ctx context.Context, namespace, podName, containerName string, lines int) (string, error) {
	args := []string{"logs", podName, "--namespace", namespace}
	if containerName != "" {
		args = append(args, "--container", containerName)
	}
	if lines > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", lines))
	}

	out, err := k.run(ctx, args)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// ExecInPod executes a command in a pod.
func (k *KubeCLIStatusChecker) ExecInPod(ctx context.Context, namespace, podName, containerName string, command []string) (string, error) {
	args := []string{"exec", podName, "--namespace", namespace}
	if containerName != "" {
		args = append(args, "--container", containerName)
	}
	args = append(args, "--")
	args = append(args, command...)

	out, err := k.run(ctx, args)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (k *KubeCLIStatusChecker) run(ctx context.Context, args []string) ([]byte, error) {
	args = append(args, k.config.ExtraArgs...)

	// Validate command using security package
	validator := security.NewCommandValidator(
		[]string{"kubectl"},
		5*time.Minute,
	)
	if err := validator.ValidateCommand(k.config.Binary, args...); err != nil {
		return nil, fmt.Errorf("command validation failed: %w", err)
	}

	//nolint:gosec // G204: Command and arguments validated by security.CommandValidator
	cmd := exec.CommandContext(ctx, k.config.Binary, args...)
	if k.config.Kubeconfig != "" {
		cmd.Env = append(os.Environ(), "KUBECONFIG="+k.config.Kubeconfig)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("kubectl %s failed: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}

	return stdout.Bytes(), nil
}
