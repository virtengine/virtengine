package main

import (
	"context"
	"fmt"
	"io"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/retry"
	ctrlconfig "sigs.k8s.io/controller-runtime/pkg/client/config"

	provider_daemon "github.com/virtengine/virtengine/pkg/provider_daemon"
)

type kubernetesAPIClient struct {
	clientset  kubernetes.Interface
	restConfig *rest.Config
}

func newKubernetesAPIClient(kubeconfig string) (*kubernetesAPIClient, error) {
	cfg, err := loadKubernetesRESTConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("create kubernetes client: %w", err)
	}

	return &kubernetesAPIClient{
		clientset:  clientset,
		restConfig: cfg,
	}, nil
}

func loadKubernetesRESTConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("load kubeconfig %s: %w", kubeconfig, err)
		}
		cfg.QPS = 20
		cfg.Burst = 40
		return cfg, nil
	}

	cfg, err := ctrlconfig.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("load kubernetes configuration: %w", err)
	}
	cfg.QPS = 20
	cfg.Burst = 40
	return cfg, nil
}

func (c *kubernetesAPIClient) CreateNamespace(ctx context.Context, name string, labels map[string]string) error {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}

	_, err := c.clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if k8serrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func (c *kubernetesAPIClient) DeleteNamespace(ctx context.Context, name string) error {
	err := c.clientset.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
	if k8serrors.IsNotFound(err) {
		return nil
	}
	return err
}

func (c *kubernetesAPIClient) CreateDeployment(ctx context.Context, namespace string, spec *provider_daemon.K8sDeploymentSpec) error {
	deployment := buildKubernetesDeployment(spec)

	_, err := c.clientset.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if !k8serrors.IsAlreadyExists(err) {
		return err
	}

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current, getErr := c.clientset.AppsV1().Deployments(namespace).Get(ctx, spec.Name, metav1.GetOptions{})
		if getErr != nil {
			return getErr
		}
		current.Labels = deployment.Labels
		current.Annotations = deployment.Annotations
		current.Spec = deployment.Spec
		_, updateErr := c.clientset.AppsV1().Deployments(namespace).Update(ctx, current, metav1.UpdateOptions{})
		return updateErr
	})
}

func (c *kubernetesAPIClient) UpdateDeployment(ctx context.Context, namespace string, spec *provider_daemon.K8sDeploymentSpec) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, spec.Name, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return c.CreateDeployment(ctx, namespace, spec)
			}
			return err
		}

		replicas := spec.Replicas
		current.Spec.Replicas = &replicas
		if len(spec.Containers) > 0 || len(spec.Volumes) > 0 || len(spec.Labels) > 0 || len(spec.Annotations) > 0 {
			desired := buildKubernetesDeployment(spec)
			current.Labels = desired.Labels
			current.Annotations = desired.Annotations
			current.Spec.Template = desired.Spec.Template
			current.Spec.Selector = desired.Spec.Selector
		}

		_, err = c.clientset.AppsV1().Deployments(namespace).Update(ctx, current, metav1.UpdateOptions{})
		return err
	})
}

func (c *kubernetesAPIClient) DeleteDeployment(ctx context.Context, namespace, name string) error {
	err := c.clientset.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if k8serrors.IsNotFound(err) {
		return nil
	}
	return err
}

func (c *kubernetesAPIClient) CreateService(ctx context.Context, namespace string, spec *provider_daemon.K8sServiceSpec) error {
	service := buildKubernetesService(spec)
	_, err := c.clientset.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	if k8serrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func (c *kubernetesAPIClient) DeleteService(ctx context.Context, namespace, name string) error {
	err := c.clientset.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if k8serrors.IsNotFound(err) {
		return nil
	}
	return err
}

func (c *kubernetesAPIClient) CreateSecret(ctx context.Context, namespace, name string, data map[string][]byte) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: data,
		Type: corev1.SecretTypeOpaque,
	}

	_, err := c.clientset.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	if !k8serrors.IsAlreadyExists(err) {
		return err
	}

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current, getErr := c.clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
		if getErr != nil {
			return getErr
		}
		current.Data = data
		current.Type = corev1.SecretTypeOpaque
		_, updateErr := c.clientset.CoreV1().Secrets(namespace).Update(ctx, current, metav1.UpdateOptions{})
		return updateErr
	})
}

func (c *kubernetesAPIClient) DeleteSecret(ctx context.Context, namespace, name string) error {
	err := c.clientset.CoreV1().Secrets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if k8serrors.IsNotFound(err) {
		return nil
	}
	return err
}

func (c *kubernetesAPIClient) CreatePVC(ctx context.Context, namespace string, spec *provider_daemon.K8sPVCSpec) error {
	pvc := buildKubernetesPVC(spec)
	_, err := c.clientset.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, pvc, metav1.CreateOptions{})
	if k8serrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func (c *kubernetesAPIClient) DeletePVC(ctx context.Context, namespace, name string) error {
	err := c.clientset.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if k8serrors.IsNotFound(err) {
		return nil
	}
	return err
}

func (c *kubernetesAPIClient) GetPodStatus(ctx context.Context, namespace, deploymentName string) ([]provider_daemon.PodStatus, error) {
	pods, err := c.listPodsForDeployment(ctx, namespace, deploymentName)
	if err != nil {
		return nil, err
	}

	statuses := make([]provider_daemon.PodStatus, 0, len(pods))
	for _, pod := range pods {
		statuses = append(statuses, provider_daemon.PodStatus{
			Name:       pod.Name,
			Phase:      string(pod.Status.Phase),
			Ready:      podReady(pod),
			Restarts:   podRestartCount(pod),
			Message:    podMessage(pod),
			StartTime:  podStartTime(pod),
			Containers: containerStatusesFromPod(pod),
		})
	}
	return statuses, nil
}

func (c *kubernetesAPIClient) ApplyNetworkPolicy(ctx context.Context, namespace string, spec *provider_daemon.K8sNetworkPolicySpec) error {
	policy := buildKubernetesNetworkPolicy(spec)
	_, err := c.clientset.NetworkingV1().NetworkPolicies(namespace).Create(ctx, policy, metav1.CreateOptions{})
	if k8serrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func (c *kubernetesAPIClient) GetServiceEndpoints(ctx context.Context, namespace, serviceName string) ([]provider_daemon.WorkloadEndpoint, error) {
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

func (c *kubernetesAPIClient) ResolveDeploymentPods(ctx context.Context, namespace, deploymentName string) ([]deploymentPod, error) {
	pods, err := c.listPodsForDeployment(ctx, namespace, deploymentName)
	if err != nil {
		return nil, err
	}

	result := make([]deploymentPod, 0, len(pods))
	for _, pod := range pods {
		containers := make([]string, 0, len(pod.Spec.Containers))
		for _, container := range pod.Spec.Containers {
			containers = append(containers, container.Name)
		}
		result = append(result, deploymentPod{
			Name:       pod.Name,
			Phase:      string(pod.Status.Phase),
			Ready:      podReady(pod),
			Containers: containers,
			CreatedAt:  pod.CreationTimestamp.Time,
		})
	}
	return result, nil
}

func (c *kubernetesAPIClient) StreamPodLogs(
	ctx context.Context,
	namespace, podName, containerName string,
	tailLines int64,
	follow bool,
) (io.ReadCloser, error) {
	options := &corev1.PodLogOptions{
		Container:  containerName,
		Follow:     follow,
		Timestamps: true,
	}
	if tailLines > 0 {
		options.TailLines = &tailLines
	}
	return c.clientset.CoreV1().Pods(namespace).GetLogs(podName, options).Stream(ctx)
}

func (c *kubernetesAPIClient) ExecInPod(
	ctx context.Context,
	namespace, podName, containerName string,
	command []string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	tty bool,
	terminalResize <-chan remotecommand.TerminalSize,
) error {
	req := c.clientset.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: containerName,
		Command:   command,
		Stdin:     stdin != nil,
		Stdout:    stdout != nil,
		Stderr:    stderr != nil,
		TTY:       tty,
	}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(c.restConfig, "POST", req.URL())
	if err != nil {
		return err
	}

	streamOptions := remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Tty:    tty,
	}
	if tty && terminalResize != nil {
		streamOptions.TerminalSizeQueue = channelTerminalSizeQueue{ch: terminalResize}
	}

	return executor.StreamWithContext(ctx, streamOptions)
}

func (c *kubernetesAPIClient) listPodsForDeployment(ctx context.Context, namespace, deploymentName string) ([]corev1.Pod, error) {
	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	selector := labels.Set(deployment.Spec.Selector.MatchLabels).String()
	list, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

type channelTerminalSizeQueue struct {
	ch <-chan remotecommand.TerminalSize
}

func (q channelTerminalSizeQueue) Next() *remotecommand.TerminalSize {
	size, ok := <-q.ch
	if !ok {
		return nil
	}
	return &size
}

func buildKubernetesDeployment(spec *provider_daemon.K8sDeploymentSpec) *appsv1.Deployment {
	replicas := spec.Replicas
	labels := cloneStringMap(spec.Labels)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        spec.Name,
			Labels:      labels,
			Annotations: cloneStringMap(spec.Annotations),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: cloneStringMap(spec.Annotations),
				},
				Spec: corev1.PodSpec{
					Containers: buildKubernetesContainers(spec.Containers),
					Volumes:    buildKubernetesVolumes(spec.Volumes),
				},
			},
		},
	}
}

func buildKubernetesContainers(specs []provider_daemon.K8sContainerSpec) []corev1.Container {
	containers := make([]corev1.Container, 0, len(specs))
	for _, spec := range specs {
		container := corev1.Container{
			Name:            spec.Name,
			Image:           spec.Image,
			Command:         append([]string(nil), spec.Command...),
			Args:            append([]string(nil), spec.Args...),
			Env:             buildKubernetesEnv(spec.Env),
			Ports:           buildKubernetesContainerPorts(spec.Ports),
			Resources:       buildKubernetesResources(spec.Resources),
			VolumeMounts:    buildKubernetesVolumeMounts(spec.VolumeMounts),
			LivenessProbe:   buildKubernetesProbe(spec.LivenessProbe),
			ReadinessProbe:  buildKubernetesProbe(spec.ReadinessProbe),
			SecurityContext: buildKubernetesSecurityContext(spec.SecurityContext),
		}
		for _, secretName := range spec.EnvFromSecrets {
			container.EnvFrom = append(container.EnvFrom, corev1.EnvFromSource{
				SecretRef: &corev1.SecretEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: secretName},
				},
			})
		}
		containers = append(containers, container)
	}
	return containers
}

func buildKubernetesEnv(values map[string]string) []corev1.EnvVar {
	env := make([]corev1.EnvVar, 0, len(values))
	for name, value := range values {
		env = append(env, corev1.EnvVar{Name: name, Value: value})
	}
	return env
}

func buildKubernetesContainerPorts(specs []provider_daemon.K8sPortSpec) []corev1.ContainerPort {
	ports := make([]corev1.ContainerPort, 0, len(specs))
	for _, spec := range specs {
		ports = append(ports, corev1.ContainerPort{
			Name:          spec.Name,
			ContainerPort: spec.ContainerPort,
			Protocol:      corev1.Protocol(spec.Protocol),
		})
	}
	return ports
}

func buildKubernetesResources(spec provider_daemon.K8sResourceSpec) corev1.ResourceRequirements {
	requests := corev1.ResourceList{}
	limits := corev1.ResourceList{}

	if spec.CPURequest != "" {
		requests[corev1.ResourceCPU] = parseQuantity(spec.CPURequest)
	}
	if spec.MemoryRequest != "" {
		requests[corev1.ResourceMemory] = parseQuantity(spec.MemoryRequest)
	}
	if spec.CPULimit != "" {
		limits[corev1.ResourceCPU] = parseQuantity(spec.CPULimit)
	}
	if spec.MemoryLimit != "" {
		limits[corev1.ResourceMemory] = parseQuantity(spec.MemoryLimit)
	}
	if spec.GPULimit != "" {
		limits[corev1.ResourceName("nvidia.com/gpu")] = parseQuantity(spec.GPULimit)
	}

	return corev1.ResourceRequirements{
		Requests: requests,
		Limits:   limits,
	}
}

func buildKubernetesVolumeMounts(specs []provider_daemon.K8sVolumeMountSpec) []corev1.VolumeMount {
	mounts := make([]corev1.VolumeMount, 0, len(specs))
	for _, spec := range specs {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      spec.Name,
			MountPath: spec.MountPath,
			SubPath:   spec.SubPath,
			ReadOnly:  spec.ReadOnly,
		})
	}
	return mounts
}

func buildKubernetesProbe(spec *provider_daemon.K8sProbeSpec) *corev1.Probe {
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
		probe.HTTPGet = &corev1.HTTPGetAction{
			Path: spec.HTTPGet.Path,
			Port: intstr.FromInt32(spec.HTTPGet.Port),
		}
		if spec.HTTPGet.Scheme != "" {
			probe.HTTPGet.Scheme = corev1.URIScheme(spec.HTTPGet.Scheme)
		}
	}
	if spec.Exec != nil {
		probe.Exec = &corev1.ExecAction{Command: append([]string(nil), spec.Exec.Command...)}
	}
	if spec.TCPSocket != nil {
		probe.TCPSocket = &corev1.TCPSocketAction{
			Port: intstr.FromInt32(spec.TCPSocket.Port),
		}
	}
	return probe
}

func buildKubernetesSecurityContext(spec *provider_daemon.K8sSecurityContext) *corev1.SecurityContext {
	if spec == nil {
		return nil
	}

	runAsNonRoot := spec.RunAsNonRoot
	readOnlyRootFS := spec.ReadOnlyRoot
	runAsUser := spec.RunAsUser

	return &corev1.SecurityContext{
		RunAsNonRoot:           &runAsNonRoot,
		RunAsUser:              &runAsUser,
		ReadOnlyRootFilesystem: &readOnlyRootFS,
	}
}

func buildKubernetesVolumes(specs []provider_daemon.K8sVolumeSpec) []corev1.Volume {
	volumes := make([]corev1.Volume, 0, len(specs))
	for _, spec := range specs {
		volume := corev1.Volume{Name: spec.Name}
		switch {
		case spec.EmptyDir:
			volume.EmptyDir = &corev1.EmptyDirVolumeSource{}
		case spec.SecretRef != "":
			volume.Secret = &corev1.SecretVolumeSource{
				SecretName: spec.SecretRef,
			}
		case spec.PVCName != "":
			volume.PersistentVolumeClaim = &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: spec.PVCName,
			}
		}
		volumes = append(volumes, volume)
	}
	return volumes
}

func buildKubernetesService(spec *provider_daemon.K8sServiceSpec) *corev1.Service {
	ports := make([]corev1.ServicePort, 0, len(spec.Ports))
	for _, port := range spec.Ports {
		ports = append(ports, corev1.ServicePort{
			Name:       port.Name,
			Port:       port.Port,
			TargetPort: intstr.FromInt32(port.TargetPort),
			Protocol:   corev1.Protocol(port.Protocol),
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
			Ports:    ports,
			Type:     corev1.ServiceType(spec.Type),
		},
	}
}

func buildKubernetesPVC(spec *provider_daemon.K8sPVCSpec) *corev1.PersistentVolumeClaim {
	storageClass := spec.StorageClass
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: spec.Name,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: buildPVCAccessModes(spec.AccessModes),
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: parseQuantity(spec.Size),
				},
			},
			StorageClassName: &storageClass,
		},
	}
}

func buildPVCAccessModes(modes []string) []corev1.PersistentVolumeAccessMode {
	accessModes := make([]corev1.PersistentVolumeAccessMode, 0, len(modes))
	for _, mode := range modes {
		accessModes = append(accessModes, corev1.PersistentVolumeAccessMode(mode))
	}
	return accessModes
}

func buildKubernetesNetworkPolicy(spec *provider_daemon.K8sNetworkPolicySpec) *networkingv1.NetworkPolicy {
	policy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: spec.Name,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: cloneStringMap(spec.PodSelector),
			},
			PolicyTypes: buildKubernetesPolicyTypes(spec.PolicyTypes),
		},
	}

	for _, ingress := range spec.IngressRules {
		rule := networkingv1.NetworkPolicyIngressRule{
			Ports: buildKubernetesNetworkPolicyPorts(ingress.Ports),
		}
		if len(ingress.FromSelector) > 0 {
			rule.From = append(rule.From, networkingv1.NetworkPolicyPeer{
				PodSelector: &metav1.LabelSelector{MatchLabels: cloneStringMap(ingress.FromSelector)},
			})
		}
		policy.Spec.Ingress = append(policy.Spec.Ingress, rule)
	}

	for _, egress := range spec.EgressRules {
		rule := networkingv1.NetworkPolicyEgressRule{
			Ports: buildKubernetesNetworkPolicyPorts(egress.Ports),
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

func buildKubernetesPolicyTypes(types []string) []networkingv1.PolicyType {
	result := make([]networkingv1.PolicyType, 0, len(types))
	for _, policyType := range types {
		result = append(result, networkingv1.PolicyType(policyType))
	}
	return result
}

func buildKubernetesNetworkPolicyPorts(specs []provider_daemon.K8sNetworkPolicyPort) []networkingv1.NetworkPolicyPort {
	ports := make([]networkingv1.NetworkPolicyPort, 0, len(specs))
	for _, spec := range specs {
		port := intstr.FromInt32(spec.Port)
		protocol := corev1.Protocol(spec.Protocol)
		ports = append(ports, networkingv1.NetworkPolicyPort{
			Protocol: &protocol,
			Port:     &port,
		})
	}
	return ports
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

func parseQuantity(value string) resource.Quantity {
	quantity, err := resource.ParseQuantity(value)
	if err != nil {
		return resource.MustParse("0")
	}
	return quantity
}

func podReady(pod corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

func podRestartCount(pod corev1.Pod) int32 {
	var restarts int32
	for _, status := range pod.Status.ContainerStatuses {
		restarts += status.RestartCount
	}
	return restarts
}

func podMessage(pod corev1.Pod) string {
	if pod.Status.Message != "" {
		return pod.Status.Message
	}
	for _, status := range pod.Status.ContainerStatuses {
		if status.State.Waiting != nil && status.State.Waiting.Message != "" {
			return status.State.Waiting.Message
		}
		if status.State.Terminated != nil && status.State.Terminated.Message != "" {
			return status.State.Terminated.Message
		}
	}
	return ""
}

func podStartTime(pod corev1.Pod) time.Time {
	if pod.Status.StartTime == nil {
		return time.Time{}
	}
	return pod.Status.StartTime.Time
}

func containerStatusesFromPod(pod corev1.Pod) []provider_daemon.ContainerStatus {
	statuses := make([]provider_daemon.ContainerStatus, 0, len(pod.Status.ContainerStatuses))
	for _, status := range pod.Status.ContainerStatuses {
		state := "unknown"
		message := ""
		switch {
		case status.State.Running != nil:
			state = "running"
		case status.State.Waiting != nil:
			state = "waiting"
			message = status.State.Waiting.Message
		case status.State.Terminated != nil:
			state = "terminated"
			message = status.State.Terminated.Message
		}
		statuses = append(statuses, provider_daemon.ContainerStatus{
			Name:         status.Name,
			Ready:        status.Ready,
			RestartCount: status.RestartCount,
			State:        state,
			Message:      message,
		})
	}
	return statuses
}
