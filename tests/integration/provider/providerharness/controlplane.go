package providerharness

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	provider_daemon "github.com/virtengine/virtengine/pkg/provider_daemon"
)

const providerHarnessTimeout = 5 * time.Second

// ControlPlane wraps an envtest-backed Kubernetes API server for provider tests.
type ControlPlane struct {
	Env       *envtest.Environment
	Config    *rest.Config
	Clientset *kubernetes.Clientset
}

// StartControlPlane starts a local envtest control plane and downloads binaries when needed.
func StartControlPlane() (*ControlPlane, error) {
	assetsDir, err := envtest.SetupEnvtestDefaultBinaryAssetsDirectory()
	if err != nil {
		return nil, fmt.Errorf("resolve envtest assets directory: %w", err)
	}

	environment := &envtest.Environment{
		DownloadBinaryAssets:     true,
		BinaryAssetsDirectory:    assetsDir,
		ControlPlaneStartTimeout: 60 * time.Second,
		ControlPlaneStopTimeout:  60 * time.Second,
	}

	config, err := environment.Start()
	if err != nil {
		return nil, fmt.Errorf("start envtest control plane: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		_ = environment.Stop()
		return nil, fmt.Errorf("create kubernetes clientset: %w", err)
	}

	return &ControlPlane{
		Env:       environment,
		Config:    config,
		Clientset: clientset,
	}, nil
}

// Stop tears down the envtest control plane.
func (cp *ControlPlane) Stop() error {
	if cp == nil || cp.Env == nil {
		return nil
	}
	err := cp.Env.Stop()
	if err != nil &&
		strings.Contains(err.Error(), "unable to signal for process") &&
		strings.Contains(strings.ToLower(err.Error()), "not supported by windows") {
		return nil
	}
	return err
}

// NewKubernetesClient returns a provider-daemon Kubernetes client backed by the envtest API server.
func (cp *ControlPlane) NewKubernetesClient() provider_daemon.KubernetesClient {
	return &EnvtestKubernetesClient{clientset: cp.Clientset}
}

// NamespaceExists reports whether a namespace exists.
func (cp *ControlPlane) NamespaceExists(ctx context.Context, namespace string) (bool, error) {
	_, err := cp.Clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return false, nil
	}
	return err == nil, err
}

// WaitForNamespaceDeleted waits until a namespace has been deleted.
func (cp *ControlPlane) WaitForNamespaceDeleted(ctx context.Context, namespace string) error {
	deadline := time.Now().Add(providerHarnessTimeout)
	for time.Now().Before(deadline) {
		_, err := cp.Clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("namespace %s still exists after cleanup", namespace)
}

// ReplaceDeploymentPods recreates the pods selected by a deployment with the requested runtime state.
func (cp *ControlPlane) ReplaceDeploymentPods(
	ctx context.Context,
	namespace string,
	deploymentName string,
	phase corev1.PodPhase,
	ready bool,
	podMessage string,
	containerState string,
	containerMessage string,
) error {
	deployment, err := cp.Clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get deployment %s: %w", deploymentName, err)
	}

	selector := labels.Set(deployment.Spec.Selector.MatchLabels).String()
	existingPods, err := cp.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return fmt.Errorf("list deployment pods for %s: %w", deploymentName, err)
	}
	for _, pod := range existingPods.Items {
		if err := cp.Clientset.CoreV1().Pods(namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("delete pod %s: %w", pod.Name, err)
		}
	}

	replicas := int32(1)
	if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas > 0 {
		replicas = *deployment.Spec.Replicas
	}

	for i := int32(0); i < replicas; i++ {
		podName := fmt.Sprintf("%s-pod-%d", deploymentName, i)
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      podName,
				Namespace: namespace,
				Labels:    cloneStringMap(deployment.Spec.Template.Labels),
			},
			Spec: deployment.Spec.Template.Spec,
		}
		if _, err := cp.Clientset.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("create pod %s: %w", podName, err)
		}

		status := corev1.PodStatus{
			Phase:   phase,
			PodIP:   fmt.Sprintf("10.42.0.%d", i+10),
			Message: podMessage,
		}
		now := metav1.NewTime(time.Now().UTC())
		status.StartTime = &now
		status.Conditions = append(status.Conditions, corev1.PodCondition{
			Type:               corev1.PodReady,
			Status:             boolToConditionStatus(ready),
			LastTransitionTime: now,
		})

		for _, container := range deployment.Spec.Template.Spec.Containers {
			containerStatus := corev1.ContainerStatus{
				Name:         container.Name,
				Ready:        ready,
				RestartCount: 0,
			}
			switch strings.ToLower(containerState) {
			case "terminated", "failed":
				containerStatus.State = corev1.ContainerState{
					Terminated: &corev1.ContainerStateTerminated{
						ExitCode:    1,
						Reason:      "Error",
						Message:     containerMessage,
						FinishedAt:  now,
						StartedAt:   now,
						ContainerID: "docker://providerharness",
					},
				}
			case "waiting":
				containerStatus.State = corev1.ContainerState{
					Waiting: &corev1.ContainerStateWaiting{
						Reason:  "Pending",
						Message: containerMessage,
					},
				}
			default:
				containerStatus.State = corev1.ContainerState{
					Running: &corev1.ContainerStateRunning{
						StartedAt: now,
					},
				}
			}
			status.ContainerStatuses = append(status.ContainerStatuses, containerStatus)
		}

		pod.Status = status
		if _, err := cp.Clientset.CoreV1().Pods(namespace).UpdateStatus(ctx, pod, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("update status for pod %s: %w", podName, err)
		}
	}

	return nil
}

// EnvtestKubernetesClient implements provider-daemon Kubernetes operations against envtest.
type EnvtestKubernetesClient struct {
	clientset kubernetes.Interface
}

// CreateNamespace creates a namespace or accepts the existing namespace.
func (c *EnvtestKubernetesClient) CreateNamespace(ctx context.Context, name string, labels map[string]string) error {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: cloneStringMap(labels),
		},
	}

	_, err := c.clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

// DeleteNamespace deletes a namespace and clears finalizers so envtest cleanup is deterministic.
func (c *EnvtestKubernetesClient) DeleteNamespace(ctx context.Context, name string) error {
	deletePolicy := metav1.DeletePropagationBackground
	_ = c.clientset.AppsV1().Deployments(name).DeleteCollection(ctx, metav1.DeleteOptions{PropagationPolicy: &deletePolicy}, metav1.ListOptions{})
	_ = c.clientset.CoreV1().Pods(name).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{})
	_ = c.clientset.CoreV1().Secrets(name).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{})
	_ = c.clientset.CoreV1().PersistentVolumeClaims(name).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{})
	_ = c.clientset.NetworkingV1().NetworkPolicies(name).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{})
	if services, err := c.clientset.CoreV1().Services(name).List(ctx, metav1.ListOptions{}); err == nil {
		for _, service := range services.Items {
			_ = c.clientset.CoreV1().Services(name).Delete(ctx, service.Name, metav1.DeleteOptions{})
		}
	}

	err := c.clientset.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	deadline := time.Now().Add(providerHarnessTimeout)
	for time.Now().Before(deadline) {
		namespace, err := c.clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if len(namespace.Spec.Finalizers) > 0 {
			updated := namespace.DeepCopy()
			updated.Spec.Finalizers = nil
			if _, err := c.clientset.CoreV1().Namespaces().Finalize(ctx, updated, metav1.UpdateOptions{}); err != nil && !apierrors.IsNotFound(err) {
				return err
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("namespace %s not deleted before timeout", name)
}

// CreateDeployment creates or updates a deployment.
func (c *EnvtestKubernetesClient) CreateDeployment(ctx context.Context, namespace string, spec *provider_daemon.K8sDeploymentSpec) error {
	deployment := buildDeployment(spec)
	_, err := c.clientset.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		return c.UpdateDeployment(ctx, namespace, spec)
	}
	return err
}

// UpdateDeployment updates the deployment spec.
func (c *EnvtestKubernetesClient) UpdateDeployment(ctx context.Context, namespace string, spec *provider_daemon.K8sDeploymentSpec) error {
	existing, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, spec.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if spec.Replicas > 0 || existing.Spec.Replicas != nil {
		existing.Spec.Replicas = ptr.To(spec.Replicas)
	}
	if len(spec.Labels) > 0 {
		existing.Labels = cloneStringMap(spec.Labels)
		existing.Spec.Selector = &metav1.LabelSelector{MatchLabels: cloneStringMap(spec.Labels)}
		existing.Spec.Template.Labels = cloneStringMap(spec.Labels)
	}
	if len(spec.Annotations) > 0 {
		existing.Annotations = cloneStringMap(spec.Annotations)
		existing.Spec.Template.Annotations = cloneStringMap(spec.Annotations)
	}
	if len(spec.Containers) > 0 {
		existing.Spec.Template.Spec.Containers = buildContainers(spec.Containers)
	}
	if len(spec.Volumes) > 0 {
		existing.Spec.Template.Spec.Volumes = buildVolumes(spec.Volumes)
	}
	_, err = c.clientset.AppsV1().Deployments(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	return err
}

// DeleteDeployment deletes a deployment.
func (c *EnvtestKubernetesClient) DeleteDeployment(ctx context.Context, namespace, name string) error {
	err := c.clientset.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}

// CreateService creates or updates a service.
func (c *EnvtestKubernetesClient) CreateService(ctx context.Context, namespace string, spec *provider_daemon.K8sServiceSpec) error {
	service := buildService(spec)
	_, err := c.clientset.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	if !apierrors.IsAlreadyExists(err) {
		return err
	}

	existing, err := c.clientset.CoreV1().Services(namespace).Get(ctx, spec.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	service.ResourceVersion = existing.ResourceVersion
	service.Spec.ClusterIP = existing.Spec.ClusterIP
	service.Spec.ClusterIPs = append([]string(nil), existing.Spec.ClusterIPs...)
	service.Spec.IPFamilies = append([]corev1.IPFamily(nil), existing.Spec.IPFamilies...)
	service.Spec.IPFamilyPolicy = existing.Spec.IPFamilyPolicy
	_, err = c.clientset.CoreV1().Services(namespace).Update(ctx, service, metav1.UpdateOptions{})
	return err
}

// DeleteService deletes a service.
func (c *EnvtestKubernetesClient) DeleteService(ctx context.Context, namespace, name string) error {
	err := c.clientset.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}

// CreateSecret creates or updates a secret.
func (c *EnvtestKubernetesClient) CreateSecret(ctx context.Context, namespace, name string, data map[string][]byte) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Type:       corev1.SecretTypeOpaque,
		Data:       cloneByteMap(data),
	}
	_, err := c.clientset.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	if !apierrors.IsAlreadyExists(err) {
		return err
	}
	existing, err := c.clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	secret.ResourceVersion = existing.ResourceVersion
	_, err = c.clientset.CoreV1().Secrets(namespace).Update(ctx, secret, metav1.UpdateOptions{})
	return err
}

// DeleteSecret deletes a secret.
func (c *EnvtestKubernetesClient) DeleteSecret(ctx context.Context, namespace, name string) error {
	err := c.clientset.CoreV1().Secrets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}

// CreatePVC creates a persistent volume claim when it does not already exist.
func (c *EnvtestKubernetesClient) CreatePVC(ctx context.Context, namespace string, spec *provider_daemon.K8sPVCSpec) error {
	pvc := buildPVC(spec)
	_, err := c.clientset.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, pvc, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

// DeletePVC deletes a PVC.
func (c *EnvtestKubernetesClient) DeletePVC(ctx context.Context, namespace, name string) error {
	err := c.clientset.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}

// GetPodStatus returns pod status for the deployment-selected pods.
func (c *EnvtestKubernetesClient) GetPodStatus(ctx context.Context, namespace, deploymentName string) ([]provider_daemon.PodStatus, error) {
	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	selector := labels.Set(deployment.Spec.Selector.MatchLabels).String()
	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, err
	}

	result := make([]provider_daemon.PodStatus, 0, len(pods.Items))
	for _, pod := range pods.Items {
		status := provider_daemon.PodStatus{
			Name:    pod.Name,
			Phase:   string(pod.Status.Phase),
			Ready:   podReady(pod),
			Message: pod.Status.Message,
		}
		if pod.Status.StartTime != nil {
			status.StartTime = pod.Status.StartTime.Time
		}
		for _, container := range pod.Status.ContainerStatuses {
			containerStatus := provider_daemon.ContainerStatus{
				Name:         container.Name,
				Ready:        container.Ready,
				RestartCount: container.RestartCount,
			}
			switch {
			case container.State.Terminated != nil:
				containerStatus.State = "terminated"
				containerStatus.Message = container.State.Terminated.Message
			case container.State.Waiting != nil:
				containerStatus.State = "waiting"
				containerStatus.Message = container.State.Waiting.Message
			case container.State.Running != nil:
				containerStatus.State = "running"
			default:
				containerStatus.State = "unknown"
			}
			status.Containers = append(status.Containers, containerStatus)
		}
		result = append(result, status)
	}

	return result, nil
}

// ApplyNetworkPolicy creates or updates a network policy.
func (c *EnvtestKubernetesClient) ApplyNetworkPolicy(ctx context.Context, namespace string, spec *provider_daemon.K8sNetworkPolicySpec) error {
	policy := buildNetworkPolicy(spec)
	_, err := c.clientset.NetworkingV1().NetworkPolicies(namespace).Create(ctx, policy, metav1.CreateOptions{})
	if !apierrors.IsAlreadyExists(err) {
		return err
	}
	existing, err := c.clientset.NetworkingV1().NetworkPolicies(namespace).Get(ctx, spec.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	policy.ResourceVersion = existing.ResourceVersion
	_, err = c.clientset.NetworkingV1().NetworkPolicies(namespace).Update(ctx, policy, metav1.UpdateOptions{})
	return err
}

// GetServiceEndpoints returns the ClusterIP-style endpoints for the service.
func (c *EnvtestKubernetesClient) GetServiceEndpoints(ctx context.Context, namespace, serviceName string) ([]provider_daemon.WorkloadEndpoint, error) {
	service, err := c.clientset.CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	internalAddress := service.Spec.ClusterIP
	if internalAddress == "" || internalAddress == "None" {
		internalAddress = fmt.Sprintf("%s.%s.svc.cluster.local", service.Name, namespace)
	}

	externalAddress := ""
	if len(service.Status.LoadBalancer.Ingress) > 0 {
		ingress := service.Status.LoadBalancer.Ingress[0]
		if ingress.IP != "" {
			externalAddress = ingress.IP
		} else {
			externalAddress = ingress.Hostname
		}
	} else if len(service.Spec.ExternalIPs) > 0 {
		externalAddress = service.Spec.ExternalIPs[0]
	}

	endpoints := make([]provider_daemon.WorkloadEndpoint, 0, len(service.Spec.Ports))
	for _, port := range service.Spec.Ports {
		endpoints = append(endpoints, provider_daemon.WorkloadEndpoint{
			Service:         service.Name,
			Port:            port.Port,
			Protocol:        string(port.Protocol),
			ExternalAddress: externalAddress,
			InternalAddress: internalAddress,
		})
	}
	return endpoints, nil
}

func buildDeployment(spec *provider_daemon.K8sDeploymentSpec) *appsv1.Deployment {
	replicas := spec.Replicas
	labels := cloneStringMap(spec.Labels)
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        spec.Name,
			Labels:      labels,
			Annotations: cloneStringMap(spec.Annotations),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(replicas),
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: cloneStringMap(spec.Annotations),
				},
				Spec: corev1.PodSpec{
					Containers: buildContainers(spec.Containers),
					Volumes:    buildVolumes(spec.Volumes),
				},
			},
		},
	}
}

func buildContainers(specs []provider_daemon.K8sContainerSpec) []corev1.Container {
	containers := make([]corev1.Container, 0, len(specs))
	for _, spec := range specs {
		container := corev1.Container{
			Name:         spec.Name,
			Image:        spec.Image,
			Command:      append([]string(nil), spec.Command...),
			Args:         append([]string(nil), spec.Args...),
			Env:          buildEnv(spec.Env),
			Ports:        buildContainerPorts(spec.Ports),
			Resources:    buildResources(spec.Resources),
			VolumeMounts: buildVolumeMounts(spec.VolumeMounts),
		}
		if spec.LivenessProbe != nil {
			container.LivenessProbe = buildProbe(spec.LivenessProbe)
		}
		if spec.ReadinessProbe != nil {
			container.ReadinessProbe = buildProbe(spec.ReadinessProbe)
		}
		if spec.SecurityContext != nil {
			container.SecurityContext = &corev1.SecurityContext{
				RunAsNonRoot:             ptr.To(spec.SecurityContext.RunAsNonRoot),
				RunAsUser:                ptr.To(spec.SecurityContext.RunAsUser),
				ReadOnlyRootFilesystem:   ptr.To(spec.SecurityContext.ReadOnlyRoot),
				AllowPrivilegeEscalation: ptr.To(false),
			}
		}
		containers = append(containers, container)
	}
	return containers
}

func buildEnv(values map[string]string) []corev1.EnvVar {
	env := make([]corev1.EnvVar, 0, len(values))
	for name, value := range values {
		env = append(env, corev1.EnvVar{Name: name, Value: value})
	}
	sort.Slice(env, func(i, j int) bool {
		return env[i].Name < env[j].Name
	})
	return env
}

func buildContainerPorts(specs []provider_daemon.K8sPortSpec) []corev1.ContainerPort {
	ports := make([]corev1.ContainerPort, 0, len(specs))
	for _, spec := range specs {
		protocol := corev1.Protocol(strings.ToUpper(spec.Protocol))
		if protocol == "" {
			protocol = corev1.ProtocolTCP
		}
		ports = append(ports, corev1.ContainerPort{
			Name:          spec.Name,
			ContainerPort: spec.ContainerPort,
			Protocol:      protocol,
		})
	}
	return ports
}

func buildResources(spec provider_daemon.K8sResourceSpec) corev1.ResourceRequirements {
	requests := corev1.ResourceList{}
	limits := corev1.ResourceList{}
	if spec.CPURequest != "" {
		requests[corev1.ResourceCPU] = resource.MustParse(spec.CPURequest)
	}
	if spec.MemoryRequest != "" {
		requests[corev1.ResourceMemory] = resource.MustParse(spec.MemoryRequest)
	}
	if spec.CPULimit != "" {
		limits[corev1.ResourceCPU] = resource.MustParse(spec.CPULimit)
	}
	if spec.MemoryLimit != "" {
		limits[corev1.ResourceMemory] = resource.MustParse(spec.MemoryLimit)
	}
	if spec.GPULimit != "" {
		limits[corev1.ResourceName("nvidia.com/gpu")] = resource.MustParse(spec.GPULimit)
	}
	return corev1.ResourceRequirements{
		Requests: requests,
		Limits:   limits,
	}
}

func buildVolumeMounts(specs []provider_daemon.K8sVolumeMountSpec) []corev1.VolumeMount {
	mounts := make([]corev1.VolumeMount, 0, len(specs))
	for _, spec := range specs {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      spec.Name,
			MountPath: spec.MountPath,
			ReadOnly:  spec.ReadOnly,
			SubPath:   spec.SubPath,
		})
	}
	return mounts
}

func buildVolumes(specs []provider_daemon.K8sVolumeSpec) []corev1.Volume {
	volumes := make([]corev1.Volume, 0, len(specs))
	for _, spec := range specs {
		volume := corev1.Volume{Name: spec.Name}
		switch {
		case spec.PVCName != "":
			volume.PersistentVolumeClaim = &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: spec.PVCName,
			}
		case spec.SecretRef != "":
			volume.Secret = &corev1.SecretVolumeSource{
				SecretName: spec.SecretRef,
			}
		default:
			volume.EmptyDir = &corev1.EmptyDirVolumeSource{}
		}
		volumes = append(volumes, volume)
	}
	return volumes
}

func buildProbe(spec *provider_daemon.K8sProbeSpec) *corev1.Probe {
	if spec == nil {
		return nil
	}
	probe := &corev1.Probe{
		InitialDelaySeconds: spec.InitialDelaySeconds,
		PeriodSeconds:       spec.PeriodSeconds,
		TimeoutSeconds:      spec.TimeoutSeconds,
		FailureThreshold:    spec.FailureThreshold,
		SuccessThreshold:    spec.SuccessThreshold,
	}
	if spec.HTTPGet != nil {
		scheme := corev1.URIScheme(strings.ToUpper(spec.HTTPGet.Scheme))
		if scheme == "" {
			scheme = corev1.URISchemeHTTP
		}
		probe.HTTPGet = &corev1.HTTPGetAction{
			Path:   spec.HTTPGet.Path,
			Port:   intstr.FromInt32(spec.HTTPGet.Port),
			Scheme: scheme,
		}
	} else if spec.TCPSocket != nil {
		probe.TCPSocket = &corev1.TCPSocketAction{
			Port: intstr.FromInt32(spec.TCPSocket.Port),
		}
	} else if spec.Exec != nil {
		probe.Exec = &corev1.ExecAction{Command: append([]string(nil), spec.Exec.Command...)}
	}
	return probe
}

func buildService(spec *provider_daemon.K8sServiceSpec) *corev1.Service {
	servicePorts := make([]corev1.ServicePort, 0, len(spec.Ports))
	for _, port := range spec.Ports {
		protocol := corev1.Protocol(strings.ToUpper(port.Protocol))
		if protocol == "" {
			protocol = corev1.ProtocolTCP
		}
		servicePorts = append(servicePorts, corev1.ServicePort{
			Name:       port.Name,
			Port:       port.Port,
			TargetPort: intstr.FromInt32(port.TargetPort),
			Protocol:   protocol,
			NodePort:   port.NodePort,
		})
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        spec.Name,
			Labels:      cloneStringMap(spec.Labels),
			Annotations: cloneStringMap(spec.Annotations),
		},
		Spec: corev1.ServiceSpec{
			Selector: cloneStringMap(spec.Selector),
			Ports:    servicePorts,
			Type:     corev1.ServiceType(spec.Type),
		},
	}
}

func buildPVC(spec *provider_daemon.K8sPVCSpec) *corev1.PersistentVolumeClaim {
	accessModes := make([]corev1.PersistentVolumeAccessMode, 0, len(spec.AccessModes))
	for _, mode := range spec.AccessModes {
		accessModes = append(accessModes, corev1.PersistentVolumeAccessMode(mode))
	}
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: spec.Name},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes:      accessModes,
			StorageClassName: ptr.To(spec.StorageClass),
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(spec.Size),
				},
			},
		},
	}
}

func buildNetworkPolicy(spec *provider_daemon.K8sNetworkPolicySpec) *networkingv1.NetworkPolicy {
	policy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: spec.Name},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{MatchLabels: cloneStringMap(spec.PodSelector)},
		},
	}
	for _, policyType := range spec.PolicyTypes {
		policy.Spec.PolicyTypes = append(policy.Spec.PolicyTypes, networkingv1.PolicyType(policyType))
	}
	for _, ingress := range spec.IngressRules {
		rule := networkingv1.NetworkPolicyIngressRule{}
		for _, port := range ingress.Ports {
			portNumber := intstr.FromInt32(port.Port)
			protocol := corev1.Protocol(strings.ToUpper(port.Protocol))
			rule.Ports = append(rule.Ports, networkingv1.NetworkPolicyPort{
				Port:     &portNumber,
				Protocol: &protocol,
			})
		}
		if len(ingress.FromSelector) > 0 {
			rule.From = append(rule.From, networkingv1.NetworkPolicyPeer{
				PodSelector: &metav1.LabelSelector{MatchLabels: cloneStringMap(ingress.FromSelector)},
			})
		}
		policy.Spec.Ingress = append(policy.Spec.Ingress, rule)
	}
	for _, egress := range spec.EgressRules {
		rule := networkingv1.NetworkPolicyEgressRule{}
		for _, port := range egress.Ports {
			portNumber := intstr.FromInt32(port.Port)
			protocol := corev1.Protocol(strings.ToUpper(port.Protocol))
			rule.Ports = append(rule.Ports, networkingv1.NetworkPolicyPort{
				Port:     &portNumber,
				Protocol: &protocol,
			})
		}
		if len(egress.ToSelector) > 0 {
			rule.To = append(rule.To, networkingv1.NetworkPolicyPeer{
				PodSelector: &metav1.LabelSelector{MatchLabels: cloneStringMap(egress.ToSelector)},
			})
		}
		policy.Spec.Egress = append(policy.Spec.Egress, rule)
	}
	return policy
}

func boolToConditionStatus(value bool) corev1.ConditionStatus {
	if value {
		return corev1.ConditionTrue
	}
	return corev1.ConditionFalse
}

func podReady(pod corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func cloneByteMap(values map[string][]byte) map[string][]byte {
	if len(values) == 0 {
		return nil
	}
	cloned := make(map[string][]byte, len(values))
	for key, value := range values {
		cloned[key] = append([]byte(nil), value...)
	}
	return cloned
}
