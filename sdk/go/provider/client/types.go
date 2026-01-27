package rest

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	inventoryV1 "github.com/virtengine/virtengine/sdk/go/inventory/v1"
	manifest "github.com/virtengine/virtengine/sdk/go/manifest/v2beta3"
)

// ClusterStatus represents the current state of the provider's cluster, including
// the number of active leases and inventory metrics.
type ClusterStatus struct {
	Leases    uint32                       `json:"leases"`
	Inventory inventoryV1.InventoryMetrics `json:"inventory"`
}

// BidEngineStatus represents the current state of the bid engine service,
// tracking the number of active orders.
type BidEngineStatus struct {
	Orders uint32 `json:"orders"`
}

// ManifestStatus represents the current state of the manifest service,
// tracking the number of active deployments.
type ManifestStatus struct {
	Deployments uint32 `json:"deployments"`
}

// ProviderStatus represents the overall status of a provider, including
// cluster, bid engine, and manifest service states, along with the provider's
// public hostname.
type ProviderStatus struct {
	Cluster               *ClusterStatus   `json:"cluster"`
	Bidengine             *BidEngineStatus `json:"bidengine"`
	Manifest              *ManifestStatus  `json:"manifest"`
	ClusterPublicHostname string           `json:"cluster_public_hostname,omitempty"`
}

// MetricTotal represents the total resource metrics across the cluster,
// including CPU, GPU, memory, ephemeral storage, and persistent storage.
type MetricTotal struct {
	CPU              uint64           `json:"cpu"`
	GPU              uint64           `json:"gpu"`
	Memory           uint64           `json:"memory"`
	StorageEphemeral uint64           `json:"storage_ephemeral"`
	Storage          map[string]int64 `json:"storage,omitempty"`
}

// ServiceStatus represents the current state of a service within a lease,
// including availability, total instances, URIs, and various replica counts.
type ServiceStatus struct {
	Name      string   `json:"name"`
	Available int32    `json:"available"`
	Total     int32    `json:"total"`
	URIs      []string `json:"uris"`

	ObservedGeneration int64 `json:"observed_generation"`
	Replicas           int32 `json:"replicas"`
	UpdatedReplicas    int32 `json:"updated_replicas"`
	ReadyReplicas      int32 `json:"ready_replicas"`
	AvailableReplicas  int32 `json:"available_replicas"`
}

// ForwardedPortStatus represents the status of a forwarded port for a service,
// including host, port numbers, protocol, and service name.
type ForwardedPortStatus struct {
	Host         string                   `json:"host,omitempty"`
	Port         uint16                   `json:"port"`
	ExternalPort uint16                   `json:"externalPort"`
	Proto        manifest.ServiceProtocol `json:"proto"`
	Name         string                   `json:"name"`
}

// LeasedIPStatus represents the status of an IP address leased to a service,
// including port numbers, protocol, and the IP address itself.
type LeasedIPStatus struct {
	Port         uint32
	ExternalPort uint32
	Protocol     string
	IP           string
}

// LeaseStatus represents the current state of a lease, including any error messages,
// service statuses, forwarded ports, and IP addresses.
type LeaseStatus struct {
	Messages       []string                         `json:"errors,omitempty"`
	Services       map[string]*ServiceStatus        `json:"services"`
	ForwardedPorts map[string][]ForwardedPortStatus `json:"forwarded_ports"` // Container services that are externally accessible
	IPs            map[string][]LeasedIPStatus      `json:"ips"`
}

// LeaseEventObject represents the object associated with a lease event,
// including its kind, namespace, and name.
type LeaseEventObject struct {
	Kind      string `json:"kind" yaml:"kind"`
	Namespace string `json:"namespace" yaml:"namespace"`
	Name      string `json:"name" yaml:"name"`
}

// LeaseEvent represents an event that occurred in relation to a lease,
// including event type, reporting information, reason, and associated object.
type LeaseEvent struct {
	Type                string           `json:"type" yaml:"type"`
	ReportingController string           `json:"reportingController,omitempty" yaml:"reportingController"`
	ReportingInstance   string           `json:"reportingInstance,omitempty" yaml:"reportingInstance"`
	Reason              string           `json:"reason" yaml:"reason"`
	Note                string           `json:"note" yaml:"note"`
	Object              LeaseEventObject `json:"object" yaml:"object"`
}

// ValidateGroupSpecResult represents the result of validating a group specification,
// including the minimum bid price required.
type ValidateGroupSpecResult struct {
	MinBidPrice sdk.DecCoin `json:"min_bid_price"`
}

// MigrateRequestBody represents a request to migrate hostnames to a new deployment,
// including the list of hostnames and destination deployment/group sequence numbers.
type MigrateRequestBody struct {
	HostnamesToMigrate []string `json:"hostnames_to_migrate"`
	DestinationDSeq    uint64   `json:"destination_dseq"`
	DestinationGSeq    uint32   `json:"destination_gseq"`
}

// EndpointMigrateRequestBody represents a request to migrate endpoints to a new deployment,
// including the list of endpoints and destination deployment/group sequence numbers.
type EndpointMigrateRequestBody struct {
	EndpointsToMigrate []string `json:"endpoints_to_migrate"`
	DestinationDSeq    uint64   `json:"destination_dseq"`
	DestinationGSeq    uint32   `json:"destination_gseq"`
}

// LeaseShellResponse represents the response from a lease shell operation,
// including the exit code and any message.
type LeaseShellResponse struct {
	ExitCode int    `json:"exit_code"`
	Message  string `json:"message,omitempty"`
}

// LeaseKubeEvent represents a Kubernetes event related to a lease,
// including the action taken and associated message.
type LeaseKubeEvent struct {
	Action  string `json:"action"`
	Message string `json:"message"`
}

// ServiceLogMessage represents a log message from a service,
// including the service name and the log message content.
type ServiceLogMessage struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

// LeaseKubeEvents represents a stream of Kubernetes events for a lease,
// including the event stream and a channel for close notifications.
type LeaseKubeEvents struct {
	Stream  <-chan LeaseEvent
	OnClose <-chan string
}

// ServiceLogs represents a stream of log messages from services,
// including the log message stream and a channel for close notifications.
type ServiceLogs struct {
	Stream  <-chan ServiceLogMessage
	OnClose <-chan string
}
