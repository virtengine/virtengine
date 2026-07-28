package providerharness

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/ptr"

	slurm_k8s "github.com/virtengine/virtengine/pkg/provider_daemon/slurm_k8s"
)

const (
	defaultSLURMNodeCPUs     int64 = 64
	defaultSLURMNodeMemoryMB int64 = 256000
	slurmControlCommand            = "scontrol"
)

// SLURMExecCall records an ExecInPod call made by the SLURM adapter.
type SLURMExecCall struct {
	Namespace string
	PodName   string
	Container string
	Command   []string
}

type slurmNode struct {
	Name     string
	CPUs     int64
	MemoryMB int64
	GPUs     int64
	GPUType  string
	State    string
}

// SLURMClusterHarness backs the SLURM-on-Kubernetes adapter with envtest resources.
type SLURMClusterHarness struct {
	cp *ControlPlane

	mu           sync.Mutex
	installErr   error
	upgradeErr   error
	uninstallErr error
	releases     map[string]*slurm_k8s.HelmRelease
	nodes        map[string]map[string]*slurmNode
	execCalls    []SLURMExecCall
}

// NewSLURMClusterHarness creates an envtest-backed Helm/Kubernetes harness for slurm_k8s tests.
func NewSLURMClusterHarness(cp *ControlPlane) *SLURMClusterHarness {
	return &SLURMClusterHarness{
		cp:       cp,
		releases: make(map[string]*slurm_k8s.HelmRelease),
		nodes:    make(map[string]map[string]*slurmNode),
	}
}

// SetInstallError configures the next install attempt to fail.
func (h *SLURMClusterHarness) SetInstallError(err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.installErr = err
}

// SetUpgradeError configures upgrade attempts to fail.
func (h *SLURMClusterHarness) SetUpgradeError(err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.upgradeErr = err
}

// SetUninstallError configures uninstall attempts to fail.
func (h *SLURMClusterHarness) SetUninstallError(err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.uninstallErr = err
}

// ExecCalls returns the recorded ExecInPod calls.
func (h *SLURMClusterHarness) ExecCalls() []SLURMExecCall {
	h.mu.Lock()
	defer h.mu.Unlock()

	calls := make([]SLURMExecCall, 0, len(h.execCalls))
	for _, call := range h.execCalls {
		calls = append(calls, SLURMExecCall{
			Namespace: call.Namespace,
			PodName:   call.PodName,
			Container: call.Container,
			Command:   append([]string(nil), call.Command...),
		})
	}
	return calls
}

// SetStatefulSetReadyReplicas adjusts the reported ready replica count for a component.
func (h *SLURMClusterHarness) SetStatefulSetReadyReplicas(
	ctx context.Context,
	namespace string,
	statefulSetName string,
	readyReplicas int32,
) error {
	statefulSet, err := h.cp.Clientset.AppsV1().StatefulSets(namespace).Get(ctx, statefulSetName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	replicas := int32(1)
	if statefulSet.Spec.Replicas != nil {
		replicas = *statefulSet.Spec.Replicas
	}

	statefulSet.Status.Replicas = replicas
	statefulSet.Status.ReadyReplicas = readyReplicas
	statefulSet.Status.CurrentReplicas = readyReplicas
	statefulSet.Status.UpdatedReplicas = readyReplicas
	if _, err := h.cp.Clientset.AppsV1().StatefulSets(namespace).UpdateStatus(ctx, statefulSet, metav1.UpdateOptions{}); err != nil {
		return err
	}

	component := strings.TrimPrefix(statefulSetName, releaseNameFromStatefulSet(statefulSetName)+"-")
	return h.applyComponentPods(ctx, namespace, releaseNameFromStatefulSet(statefulSetName), component, replicas, readyReplicas)
}

// SetNodeState sets the mocked SLURM state for a node.
func (h *SLURMClusterHarness) SetNodeState(namespace, releaseName, nodeName, state string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	key := releaseKey(namespace, releaseName)
	if h.nodes[key] == nil {
		h.nodes[key] = make(map[string]*slurmNode)
	}
	node, ok := h.nodes[key][nodeName]
	if !ok {
		node = &slurmNode{
			Name:     nodeName,
			CPUs:     defaultSLURMNodeCPUs,
			MemoryMB: defaultSLURMNodeMemoryMB,
		}
		h.nodes[key][nodeName] = node
	}
	node.State = normalizeNodeState(state)
}

// Install installs a release and materializes SLURM control-plane StatefulSets in envtest.
func (h *SLURMClusterHarness) Install(
	ctx context.Context,
	releaseName string,
	chartPath string,
	namespace string,
	values map[string]interface{},
) error {
	h.mu.Lock()
	installErr := h.installErr
	h.mu.Unlock()
	if installErr != nil {
		return installErr
	}

	if err := ensureNamespace(ctx, h.cp, namespace); err != nil {
		return err
	}

	computeReplicas := computeReplicaCount(values)
	if computeReplicas < 1 {
		computeReplicas = 1
	}

	if err := h.applyStatefulSet(ctx, namespace, releaseName, "controller", 1, 1); err != nil {
		return err
	}
	if err := h.applyStatefulSet(ctx, namespace, releaseName, "slurmdbd", 1, 1); err != nil {
		return err
	}
	if err := h.applyStatefulSet(ctx, namespace, releaseName, "compute", computeReplicas, computeReplicas); err != nil {
		return err
	}

	key := releaseKey(namespace, releaseName)
	h.mu.Lock()
	h.releases[key] = &slurm_k8s.HelmRelease{
		Name:       releaseName,
		Namespace:  namespace,
		Chart:      chartPath,
		Status:     "deployed",
		AppVersion: "envtest",
		Values:     cloneInterfaceMap(values),
	}
	h.ensureNodesLocked(key, computeReplicas)
	h.mu.Unlock()
	return nil
}

// Upgrade upgrades a release and refreshes the envtest objects.
func (h *SLURMClusterHarness) Upgrade(
	ctx context.Context,
	releaseName string,
	chartPath string,
	namespace string,
	values map[string]interface{},
) error {
	h.mu.Lock()
	upgradeErr := h.upgradeErr
	h.mu.Unlock()
	if upgradeErr != nil {
		return upgradeErr
	}

	if err := ensureNamespace(ctx, h.cp, namespace); err != nil {
		return err
	}

	computeReplicas := computeReplicaCount(values)
	if computeReplicas < 1 {
		computeReplicas = 1
	}

	if err := h.applyStatefulSet(ctx, namespace, releaseName, "controller", 1, 1); err != nil {
		return err
	}
	if err := h.applyStatefulSet(ctx, namespace, releaseName, "slurmdbd", 1, 1); err != nil {
		return err
	}
	if err := h.applyStatefulSet(ctx, namespace, releaseName, "compute", computeReplicas, computeReplicas); err != nil {
		return err
	}

	key := releaseKey(namespace, releaseName)
	h.mu.Lock()
	release := h.releases[key]
	if release == nil {
		release = &slurm_k8s.HelmRelease{Name: releaseName, Namespace: namespace}
		h.releases[key] = release
	}
	release.Chart = chartPath
	release.Status = "deployed"
	release.AppVersion = "envtest"
	release.Values = cloneInterfaceMap(values)
	h.ensureNodesLocked(key, computeReplicas)
	h.mu.Unlock()
	return nil
}

// Uninstall deletes a release and its envtest resources.
func (h *SLURMClusterHarness) Uninstall(ctx context.Context, releaseName, namespace string) error {
	h.mu.Lock()
	uninstallErr := h.uninstallErr
	h.mu.Unlock()
	if uninstallErr != nil {
		return uninstallErr
	}

	for _, name := range []string{
		releaseName + "-controller",
		releaseName + "-slurmdbd",
		releaseName + "-compute",
	} {
		err := h.cp.Clientset.AppsV1().StatefulSets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}

	podSelector := labels.Set(map[string]string{"app.kubernetes.io/instance": releaseName}).String()
	if err := h.cp.Clientset.CoreV1().Pods(namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: podSelector,
	}); err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	h.mu.Lock()
	delete(h.releases, releaseKey(namespace, releaseName))
	delete(h.nodes, releaseKey(namespace, releaseName))
	h.mu.Unlock()
	return nil
}

// GetRelease returns a single Helm release.
func (h *SLURMClusterHarness) GetRelease(ctx context.Context, releaseName, namespace string) (*slurm_k8s.HelmRelease, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	release := h.releases[releaseKey(namespace, releaseName)]
	if release == nil {
		return nil, fmt.Errorf("release %s not found", releaseName)
	}
	return cloneRelease(release), nil
}

// ListReleases lists the releases in a namespace.
func (h *SLURMClusterHarness) ListReleases(ctx context.Context, namespace string) ([]*slurm_k8s.HelmRelease, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	releases := make([]*slurm_k8s.HelmRelease, 0)
	for _, release := range h.releases {
		if namespace == "" || release.Namespace == namespace {
			releases = append(releases, cloneRelease(release))
		}
	}
	sort.Slice(releases, func(i, j int) bool {
		return releases[i].Name < releases[j].Name
	})
	return releases, nil
}

// GetStatefulSetStatus returns status derived from envtest StatefulSets.
func (h *SLURMClusterHarness) GetStatefulSetStatus(
	ctx context.Context,
	namespace string,
	name string,
) (*slurm_k8s.StatefulSetStatus, error) {
	statefulSet, err := h.cp.Clientset.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	replicas := int32(0)
	if statefulSet.Spec.Replicas != nil {
		replicas = *statefulSet.Spec.Replicas
	}
	conditions := make([]string, 0, len(statefulSet.Status.Conditions))
	for _, condition := range statefulSet.Status.Conditions {
		conditions = append(conditions, fmt.Sprintf("%s=%s", condition.Type, condition.Status))
	}
	sort.Strings(conditions)

	return &slurm_k8s.StatefulSetStatus{
		Name:            name,
		Replicas:        replicas,
		ReadyReplicas:   statefulSet.Status.ReadyReplicas,
		CurrentReplicas: statefulSet.Status.CurrentReplicas,
		UpdatedReplicas: statefulSet.Status.UpdatedReplicas,
		Conditions:      conditions,
	}, nil
}

// GetPodLogs returns a synthetic controller log line for operator-facing error messages.
func (h *SLURMClusterHarness) GetPodLogs(
	ctx context.Context,
	namespace string,
	podName string,
	containerName string,
	lines int,
) (string, error) {
	return fmt.Sprintf("%s/%s ready in %s", podName, containerName, namespace), nil
}

// ExecInPod emulates the SLURM control-plane commands that the adapter issues.
func (h *SLURMClusterHarness) ExecInPod(
	ctx context.Context,
	namespace string,
	podName string,
	containerName string,
	command []string,
) (string, error) {
	h.mu.Lock()
	h.execCalls = append(h.execCalls, SLURMExecCall{
		Namespace: namespace,
		PodName:   podName,
		Container: containerName,
		Command:   append([]string(nil), command...),
	})
	h.mu.Unlock()

	releaseName := releaseNameFromPod(podName)
	if releaseName == "" {
		return "", fmt.Errorf("unsupported pod %s", podName)
	}

	switch {
	case len(command) >= 2 && command[0] == slurmControlCommand && command[1] == "ping":
		return "Slurmctld(primary) at " + podName + " is UP", nil
	case len(command) >= 5 && command[0] == "sinfo" && command[1] == "-N" && command[2] == "-h" && command[3] == "-o":
		return h.renderSinfo(namespace, releaseName, command[4]), nil
	case len(command) >= 6 && command[0] == "sinfo" && command[1] == "-n" && command[3] == "-h" && command[4] == "-o":
		nodeName := command[2]
		return h.renderSingleNodeState(namespace, releaseName, nodeName, command[5]), nil
	case len(command) >= 4 && command[0] == slurmControlCommand && command[1] == "update":
		return h.handleScontrolUpdate(namespace, releaseName, command[2:])
	case len(command) >= 3 && command[0] == slurmControlCommand && command[1] == "create" && command[2] == "node":
		return h.handleScontrolCreateNode(namespace, releaseName, command[3:])
	case len(command) >= 3 && command[0] == slurmControlCommand && command[1] == "delete":
		return h.handleScontrolDeleteNode(namespace, releaseName, command[2:])
	default:
		return "", fmt.Errorf("unsupported slurm exec command: %v", command)
	}
}

func (h *SLURMClusterHarness) applyStatefulSet(
	ctx context.Context,
	namespace string,
	releaseName string,
	component string,
	replicas int32,
	readyReplicas int32,
) error {
	labels := map[string]string{
		"app.kubernetes.io/instance":  releaseName,
		"app.kubernetes.io/component": component,
	}
	name := fmt.Sprintf("%s-%s", releaseName, component)
	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    cloneStringMap(labels),
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:    ptr.To(replicas),
			ServiceName: name,
			Selector:    &metav1.LabelSelector{MatchLabels: cloneStringMap(labels)},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: cloneStringMap(labels)},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  componentContainerName(component),
						Image: "ghcr.io/virtengine/provider-harness:envtest",
					}},
				},
			},
		},
	}

	if _, err := h.cp.Clientset.AppsV1().StatefulSets(namespace).Create(ctx, statefulSet, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		}
		existing, getErr := h.cp.Clientset.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if getErr != nil {
			return getErr
		}
		existing.Labels = cloneStringMap(labels)
		existing.Spec.Replicas = ptr.To(replicas)
		existing.Spec.Selector = &metav1.LabelSelector{MatchLabels: cloneStringMap(labels)}
		existing.Spec.Template.Labels = cloneStringMap(labels)
		existing.Spec.Template.Spec.Containers = statefulSet.Spec.Template.Spec.Containers
		if _, err := h.cp.Clientset.AppsV1().StatefulSets(namespace).Update(ctx, existing, metav1.UpdateOptions{}); err != nil {
			return err
		}
	}

	current, err := h.cp.Clientset.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	current.Status.Replicas = replicas
	current.Status.ReadyReplicas = readyReplicas
	current.Status.CurrentReplicas = readyReplicas
	current.Status.UpdatedReplicas = readyReplicas
	if _, err := h.cp.Clientset.AppsV1().StatefulSets(namespace).UpdateStatus(ctx, current, metav1.UpdateOptions{}); err != nil {
		return err
	}

	return h.applyComponentPods(ctx, namespace, releaseName, component, replicas, readyReplicas)
}

func (h *SLURMClusterHarness) applyComponentPods(
	ctx context.Context,
	namespace string,
	releaseName string,
	component string,
	replicas int32,
	readyReplicas int32,
) error {
	selector := labels.Set(map[string]string{
		"app.kubernetes.io/instance":  releaseName,
		"app.kubernetes.io/component": component,
	}).String()
	existingPods, err := h.cp.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return err
	}

	expectedNames := make(map[string]struct{}, replicas)
	for i := int32(0); i < replicas; i++ {
		podName := fmt.Sprintf("%s-%s-%d", releaseName, component, i)
		expectedNames[podName] = struct{}{}
	}
	for _, pod := range existingPods.Items {
		if _, ok := expectedNames[pod.Name]; ok {
			continue
		}
		if err := h.cp.Clientset.CoreV1().Pods(namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}

	for i := int32(0); i < replicas; i++ {
		podName := fmt.Sprintf("%s-%s-%d", releaseName, component, i)
		ready := i < readyReplicas
		if err := h.applyPod(ctx, namespace, podName, releaseName, component, ready); err != nil {
			return err
		}
	}
	return nil
}

func (h *SLURMClusterHarness) applyPod(
	ctx context.Context,
	namespace string,
	podName string,
	releaseName string,
	component string,
	ready bool,
) error {
	labels := map[string]string{
		"app.kubernetes.io/instance":  releaseName,
		"app.kubernetes.io/component": component,
	}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
			Labels:    cloneStringMap(labels),
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:  componentContainerName(component),
				Image: "ghcr.io/virtengine/provider-harness:envtest",
			}},
		},
	}

	if _, err := h.cp.Clientset.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		}
	}

	current, err := h.cp.Clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	now := metav1.NewTime(time.Now().UTC())
	current.Status.Phase = corev1.PodRunning
	current.Status.StartTime = &now
	if !ready {
		current.Status.Phase = corev1.PodPending
	}
	current.Status.Conditions = []corev1.PodCondition{{
		Type:               corev1.PodReady,
		Status:             boolToConditionStatus(ready),
		LastTransitionTime: now,
	}}
	current.Status.ContainerStatuses = []corev1.ContainerStatus{{
		Name:  componentContainerName(component),
		Ready: ready,
		State: runningOrWaitingState(now, ready),
	}}
	if _, err := h.cp.Clientset.CoreV1().Pods(namespace).UpdateStatus(ctx, current, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}

func (h *SLURMClusterHarness) handleScontrolUpdate(
	namespace string,
	releaseName string,
	args []string,
) (string, error) {
	nodeName := argValue(args, "NodeName")
	state := argValue(args, "State")
	if nodeName == "" || state == "" {
		return "", fmt.Errorf("node update requires NodeName and State")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	key := releaseKey(namespace, releaseName)
	if h.nodes[key] == nil {
		h.nodes[key] = make(map[string]*slurmNode)
	}
	node, ok := h.nodes[key][nodeName]
	if !ok {
		node = &slurmNode{
			Name:     nodeName,
			CPUs:     defaultSLURMNodeCPUs,
			MemoryMB: defaultSLURMNodeMemoryMB,
		}
		h.nodes[key][nodeName] = node
	}
	node.State = normalizeNodeState(state)
	return "ok", nil
}

func (h *SLURMClusterHarness) handleScontrolCreateNode(
	namespace string,
	releaseName string,
	args []string,
) (string, error) {
	nodeName := argValue(args, "NodeName")
	if nodeName == "" {
		return "", fmt.Errorf("node creation requires NodeName")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	key := releaseKey(namespace, releaseName)
	if h.nodes[key] == nil {
		h.nodes[key] = make(map[string]*slurmNode)
	}
	node := &slurmNode{
		Name:     nodeName,
		CPUs:     parseInt64Arg(argValue(args, "CPUs"), defaultSLURMNodeCPUs),
		MemoryMB: parseInt64Arg(argValue(args, "RealMemory"), defaultSLURMNodeMemoryMB),
		State:    normalizeNodeState(argValue(args, "State")),
	}
	if node.State == "" {
		node.State = "future"
	}
	if gres := argValue(args, "Gres"); gres != "" {
		node.GPUType, node.GPUs = parseNodeGRES(gres)
	}
	h.nodes[key][nodeName] = node
	return "ok", nil
}

func (h *SLURMClusterHarness) handleScontrolDeleteNode(
	namespace string,
	releaseName string,
	args []string,
) (string, error) {
	nodeName := argValue(args, "NodeName")
	if nodeName == "" {
		return "", fmt.Errorf("node delete requires NodeName")
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.nodes[releaseKey(namespace, releaseName)], nodeName)
	return "ok", nil
}

func (h *SLURMClusterHarness) renderSinfo(namespace string, releaseName string, format string) string {
	h.mu.Lock()
	defer h.mu.Unlock()

	nodes := sortedNodes(h.nodes[releaseKey(namespace, releaseName)])
	lines := make([]string, 0, len(nodes))
	for _, node := range nodes {
		switch format {
		case "%n %c %m %G %t":
			lines = append(lines, fmt.Sprintf("%s %d %d %s %s", node.Name, node.CPUs, node.MemoryMB, renderNodeGRES(node), node.State))
		case "%n %t":
			lines = append(lines, fmt.Sprintf("%s %s", node.Name, node.State))
		case "%n":
			lines = append(lines, node.Name)
		}
	}
	return strings.Join(lines, "\n")
}

func (h *SLURMClusterHarness) renderSingleNodeState(
	namespace string,
	releaseName string,
	nodeName string,
	format string,
) string {
	h.mu.Lock()
	defer h.mu.Unlock()

	node := h.nodes[releaseKey(namespace, releaseName)][nodeName]
	if node == nil {
		return ""
	}
	if format == "%t" {
		return node.State
	}
	return ""
}

func (h *SLURMClusterHarness) ensureNodesLocked(key string, replicas int32) {
	if h.nodes[key] == nil {
		h.nodes[key] = make(map[string]*slurmNode)
	}

	expected := make(map[string]struct{}, replicas)
	for i := int32(0); i < replicas; i++ {
		name := fmt.Sprintf("compute-%d", i)
		expected[name] = struct{}{}
		if _, ok := h.nodes[key][name]; !ok {
			h.nodes[key][name] = &slurmNode{
				Name:     name,
				CPUs:     defaultSLURMNodeCPUs,
				MemoryMB: defaultSLURMNodeMemoryMB,
				State:    "idle",
			}
		}
	}

	for name := range h.nodes[key] {
		if _, ok := expected[name]; !ok {
			delete(h.nodes[key], name)
		}
	}
}

// RecordingReporter captures status and capacity reports for assertions.
type RecordingReporter struct {
	mu              sync.Mutex
	StatusReports   []slurm_k8s.ClusterStatusUpdate
	CapacityReports []slurm_k8s.ClusterCapacity
	JoinedNodes     []string
	RemovedNodes    []string
}

// ReportClusterStatus records cluster status updates.
func (r *RecordingReporter) ReportClusterStatus(
	ctx context.Context,
	clusterID string,
	status *slurm_k8s.ClusterStatusUpdate,
) error {
	if status == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.StatusReports = append(r.StatusReports, *status)
	return nil
}

// ReportCapacityUpdate records capacity updates.
func (r *RecordingReporter) ReportCapacityUpdate(
	ctx context.Context,
	clusterID string,
	capacity *slurm_k8s.ClusterCapacity,
) error {
	if capacity == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.CapacityReports = append(r.CapacityReports, *capacity)
	return nil
}

// ReportNodeJoin records node joins.
func (r *RecordingReporter) ReportNodeJoin(ctx context.Context, clusterID, nodeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.JoinedNodes = append(r.JoinedNodes, nodeID)
	return nil
}

// ReportNodeLeave records node leaves.
func (r *RecordingReporter) ReportNodeLeave(ctx context.Context, clusterID, nodeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.RemovedNodes = append(r.RemovedNodes, nodeID)
	return nil
}

func ensureNamespace(ctx context.Context, cp *ControlPlane, namespace string) error {
	if namespace == "" {
		return fmt.Errorf("namespace is required")
	}
	_, err := cp.Clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = cp.Clientset.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: namespace},
		}, metav1.CreateOptions{})
	}
	return err
}

func runningOrWaitingState(now metav1.Time, ready bool) corev1.ContainerState {
	if ready {
		return corev1.ContainerState{
			Running: &corev1.ContainerStateRunning{StartedAt: now},
		}
	}
	return corev1.ContainerState{
		Waiting: &corev1.ContainerStateWaiting{
			Reason:  "Pending",
			Message: "component not ready",
		},
	}
}

func cloneRelease(release *slurm_k8s.HelmRelease) *slurm_k8s.HelmRelease {
	if release == nil {
		return nil
	}
	return &slurm_k8s.HelmRelease{
		Name:       release.Name,
		Namespace:  release.Namespace,
		Chart:      release.Chart,
		Version:    release.Version,
		AppVersion: release.AppVersion,
		Status:     release.Status,
		Values:     cloneInterfaceMap(release.Values),
	}
}

func cloneInterfaceMap(values map[string]interface{}) map[string]interface{} {
	if len(values) == 0 {
		return nil
	}
	cloned := make(map[string]interface{}, len(values))
	for key, value := range values {
		cloned[key] = cloneInterfaceValue(value)
	}
	return cloned
}

func cloneInterfaceValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		return cloneInterfaceMap(typed)
	case []map[string]interface{}:
		cloned := make([]map[string]interface{}, 0, len(typed))
		for _, item := range typed {
			cloned = append(cloned, cloneInterfaceMap(item))
		}
		return cloned
	case []interface{}:
		cloned := make([]interface{}, 0, len(typed))
		for _, item := range typed {
			cloned = append(cloned, cloneInterfaceValue(item))
		}
		return cloned
	default:
		return typed
	}
}

func computeReplicaCount(values map[string]interface{}) int32 {
	compute, ok := values["compute"].(map[string]interface{})
	if !ok {
		return 1
	}
	return normalizeReplicaValue(compute["replicas"], 1)
}

func normalizeReplicaValue(value interface{}, fallback int32) int32 {
	switch typed := value.(type) {
	case int:
		return safeInt32FromInt(typed)
	case int32:
		return typed
	case int64:
		return safeInt32FromInt64(typed)
	case float64:
		return safeInt32FromFloat64(typed)
	default:
		return fallback
	}
}

func safeInt32FromInt(value int) int32 {
	if value < 0 {
		return 0
	}
	if value > int(^uint32(0)>>1) {
		return int32(^uint32(0) >> 1)
	}
	return int32(value)
}

func safeInt32FromInt64(value int64) int32 {
	if value < 0 {
		return 0
	}
	if value > int64(^uint32(0)>>1) {
		return int32(^uint32(0) >> 1)
	}
	return int32(value)
}

func safeInt32FromFloat64(value float64) int32 {
	if math.IsNaN(value) || value <= 0 {
		return 0
	}
	if value > float64(int32(^uint32(0)>>1)) {
		return int32(^uint32(0) >> 1)
	}
	return int32(value)
}

func componentContainerName(component string) string {
	switch component {
	case "controller":
		return "slurmctld"
	case "slurmdbd":
		return "slurmdbd"
	default:
		return "slurmd"
	}
}

func releaseNameFromPod(podName string) string {
	for _, suffix := range []string{"-controller-0", "-slurmdbd-0"} {
		if strings.HasSuffix(podName, suffix) {
			return strings.TrimSuffix(podName, suffix)
		}
	}
	if idx := strings.LastIndex(podName, "-compute-"); idx > 0 {
		return podName[:idx]
	}
	return ""
}

func releaseNameFromStatefulSet(name string) string {
	for _, suffix := range []string{"-controller", "-slurmdbd", "-compute"} {
		if strings.HasSuffix(name, suffix) {
			return strings.TrimSuffix(name, suffix)
		}
	}
	return name
}

func releaseKey(namespace, releaseName string) string {
	return namespace + "/" + releaseName
}

func argValue(args []string, prefix string) string {
	for _, arg := range args {
		if strings.HasPrefix(arg, prefix+"=") {
			return strings.TrimPrefix(arg, prefix+"=")
		}
	}
	return ""
}

func parseInt64Arg(value string, fallback int64) int64 {
	if value == "" {
		return fallback
	}
	var parsed int64
	if _, err := fmt.Sscanf(value, "%d", &parsed); err != nil {
		return fallback
	}
	return parsed
}

func parseNodeGRES(value string) (string, int64) {
	parts := strings.Split(value, ":")
	if len(parts) == 3 {
		return parts[1], parseInt64Arg(parts[2], 0)
	}
	if len(parts) == 2 {
		return "", parseInt64Arg(parts[1], 0)
	}
	return "", 0
}

func renderNodeGRES(node *slurmNode) string {
	if node == nil || node.GPUs == 0 {
		return "(null)"
	}
	if node.GPUType == "" {
		return fmt.Sprintf("gpu:%d", node.GPUs)
	}
	return fmt.Sprintf("gpu:%s:%d", node.GPUType, node.GPUs)
}

func normalizeNodeState(state string) string {
	switch strings.ToUpper(strings.TrimSpace(state)) {
	case "RESUME":
		return "idle"
	case "DRAIN":
		return "drained"
	case "DOWN":
		return "down"
	case "FUTURE":
		return "future"
	case "ALLOC", "ALLOCATED":
		return "allocated"
	case "MIXED":
		return "mixed"
	case "IDLE":
		return "idle"
	default:
		return strings.ToLower(strings.TrimSpace(state))
	}
}

func sortedNodes(nodes map[string]*slurmNode) []*slurmNode {
	list := make([]*slurmNode, 0, len(nodes))
	for _, node := range nodes {
		if node != nil {
			list = append(list, node)
		}
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Name < list[j].Name
	})
	return list
}
