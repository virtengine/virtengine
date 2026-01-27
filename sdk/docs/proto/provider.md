<!-- This file is auto-generated. Please do not modify it yourself. -->
 # Protobuf Documentation
 <a name="top"></a>

 ## Table of Contents
 
 - [virtengine/inventory/v1/resourcepair.proto](#virtengine/inventory/v1/resourcepair.proto)
     - [ResourcePair](#virtengine.inventory.v1.ResourcePair)
   
 - [virtengine/inventory/v1/cpu.proto](#virtengine/inventory/v1/cpu.proto)
     - [CPU](#virtengine.inventory.v1.CPU)
     - [CPUInfo](#virtengine.inventory.v1.CPUInfo)
   
 - [virtengine/inventory/v1/gpu.proto](#virtengine/inventory/v1/gpu.proto)
     - [GPU](#virtengine.inventory.v1.GPU)
     - [GPUInfo](#virtengine.inventory.v1.GPUInfo)
   
 - [virtengine/inventory/v1/memory.proto](#virtengine/inventory/v1/memory.proto)
     - [Memory](#virtengine.inventory.v1.Memory)
     - [MemoryInfo](#virtengine.inventory.v1.MemoryInfo)
   
 - [virtengine/inventory/v1/resources.proto](#virtengine/inventory/v1/resources.proto)
     - [NodeResources](#virtengine.inventory.v1.NodeResources)
   
 - [virtengine/inventory/v1/node.proto](#virtengine/inventory/v1/node.proto)
     - [Node](#virtengine.inventory.v1.Node)
     - [NodeCapabilities](#virtengine.inventory.v1.NodeCapabilities)
   
 - [virtengine/inventory/v1/storage.proto](#virtengine/inventory/v1/storage.proto)
     - [Storage](#virtengine.inventory.v1.Storage)
     - [StorageInfo](#virtengine.inventory.v1.StorageInfo)
   
 - [virtengine/inventory/v1/cluster.proto](#virtengine/inventory/v1/cluster.proto)
     - [Cluster](#virtengine.inventory.v1.Cluster)
   
 - [virtengine/inventory/v1/service.proto](#virtengine/inventory/v1/service.proto)
     - [ClusterRPC](#virtengine.inventory.v1.ClusterRPC)
     - [NodeRPC](#virtengine.inventory.v1.NodeRPC)
   
 - [virtengine/manifest/v2beta3/httpoptions.proto](#virtengine/manifest/v2beta3/httpoptions.proto)
     - [ServiceExposeHTTPOptions](#virtengine.manifest.v2beta3.ServiceExposeHTTPOptions)
   
 - [virtengine/manifest/v2beta3/serviceexpose.proto](#virtengine/manifest/v2beta3/serviceexpose.proto)
     - [ServiceExpose](#virtengine.manifest.v2beta3.ServiceExpose)
   
 - [virtengine/manifest/v2beta3/service.proto](#virtengine/manifest/v2beta3/service.proto)
     - [ImageCredentials](#virtengine.manifest.v2beta3.ImageCredentials)
     - [Service](#virtengine.manifest.v2beta3.Service)
     - [ServiceParams](#virtengine.manifest.v2beta3.ServiceParams)
     - [StorageParams](#virtengine.manifest.v2beta3.StorageParams)
   
 - [virtengine/manifest/v2beta3/group.proto](#virtengine/manifest/v2beta3/group.proto)
     - [Group](#virtengine.manifest.v2beta3.Group)
   
 - [virtengine/provider/lease/v1/service.proto](#virtengine/provider/lease/v1/service.proto)
     - [ForwarderPortStatus](#virtengine.provider.lease.v1.ForwarderPortStatus)
     - [LeaseIPStatus](#virtengine.provider.lease.v1.LeaseIPStatus)
     - [LeaseServiceStatus](#virtengine.provider.lease.v1.LeaseServiceStatus)
     - [SendManifestRequest](#virtengine.provider.lease.v1.SendManifestRequest)
     - [SendManifestResponse](#virtengine.provider.lease.v1.SendManifestResponse)
     - [ServiceLogs](#virtengine.provider.lease.v1.ServiceLogs)
     - [ServiceLogsRequest](#virtengine.provider.lease.v1.ServiceLogsRequest)
     - [ServiceLogsResponse](#virtengine.provider.lease.v1.ServiceLogsResponse)
     - [ServiceStatus](#virtengine.provider.lease.v1.ServiceStatus)
     - [ServiceStatusRequest](#virtengine.provider.lease.v1.ServiceStatusRequest)
     - [ServiceStatusResponse](#virtengine.provider.lease.v1.ServiceStatusResponse)
     - [ShellRequest](#virtengine.provider.lease.v1.ShellRequest)
   
     - [LeaseRPC](#virtengine.provider.lease.v1.LeaseRPC)
   
 - [virtengine/provider/v1/status.proto](#virtengine/provider/v1/status.proto)
     - [BidEngineStatus](#virtengine.provider.v1.BidEngineStatus)
     - [ClusterStatus](#virtengine.provider.v1.ClusterStatus)
     - [Inventory](#virtengine.provider.v1.Inventory)
     - [Leases](#virtengine.provider.v1.Leases)
     - [ManifestStatus](#virtengine.provider.v1.ManifestStatus)
     - [Reservations](#virtengine.provider.v1.Reservations)
     - [ReservationsMetric](#virtengine.provider.v1.ReservationsMetric)
     - [ResourcesMetric](#virtengine.provider.v1.ResourcesMetric)
     - [ResourcesMetric.StorageEntry](#virtengine.provider.v1.ResourcesMetric.StorageEntry)
     - [Status](#virtengine.provider.v1.Status)
   
 - [virtengine/provider/v1/service.proto](#virtengine/provider/v1/service.proto)
     - [ProviderRPC](#virtengine.provider.v1.ProviderRPC)
   
 - [Scalar Value Types](#scalar-value-types)

 
 
 <a name="virtengine/inventory/v1/resourcepair.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/inventory/v1/resourcepair.proto
 

 
 <a name="virtengine.inventory.v1.ResourcePair"></a>

 ### ResourcePair
 ResourcePair to extents resource.Quantity to provide total and available units of the resource

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `allocatable` | [k8s.io.apimachinery.pkg.api.resource.Quantity](#k8s.io.apimachinery.pkg.api.resource.Quantity) |  |  |
 | `allocated` | [k8s.io.apimachinery.pkg.api.resource.Quantity](#k8s.io.apimachinery.pkg.api.resource.Quantity) |  |  |
 | `attributes` | [virtengine.base.attributes.v1.Attribute](#virtengine.base.attributes.v1.Attribute) | repeated |  |
 | `capacity` | [k8s.io.apimachinery.pkg.api.resource.Quantity](#k8s.io.apimachinery.pkg.api.resource.Quantity) |  |  |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/inventory/v1/cpu.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/inventory/v1/cpu.proto
 

 
 <a name="virtengine.inventory.v1.CPU"></a>

 ### CPU
 CPU reports CPU inventory details

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `quantity` | [ResourcePair](#virtengine.inventory.v1.ResourcePair) |  |  |
 | `info` | [CPUInfo](#virtengine.inventory.v1.CPUInfo) | repeated |  |
 
 

 

 
 <a name="virtengine.inventory.v1.CPUInfo"></a>

 ### CPUInfo
 CPUInfo reports CPU details

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [string](#string) |  |  |
 | `vendor` | [string](#string) |  |  |
 | `model` | [string](#string) |  |  |
 | `vcores` | [uint32](#uint32) |  |  |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/inventory/v1/gpu.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/inventory/v1/gpu.proto
 

 
 <a name="virtengine.inventory.v1.GPU"></a>

 ### GPU
 GPUInfo reports GPU inventory details

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `quantity` | [ResourcePair](#virtengine.inventory.v1.ResourcePair) |  |  |
 | `info` | [GPUInfo](#virtengine.inventory.v1.GPUInfo) | repeated |  |
 
 

 

 
 <a name="virtengine.inventory.v1.GPUInfo"></a>

 ### GPUInfo
 GPUInfo reports GPU details

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `vendor` | [string](#string) |  |  |
 | `vendor_id` | [string](#string) |  |  |
 | `name` | [string](#string) |  |  |
 | `modelid` | [string](#string) |  |  |
 | `interface` | [string](#string) |  |  |
 | `memory_size` | [string](#string) |  |  |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/inventory/v1/memory.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/inventory/v1/memory.proto
 

 
 <a name="virtengine.inventory.v1.Memory"></a>

 ### Memory
 Memory reports Memory inventory details

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `quantity` | [ResourcePair](#virtengine.inventory.v1.ResourcePair) |  |  |
 | `info` | [MemoryInfo](#virtengine.inventory.v1.MemoryInfo) | repeated |  |
 
 

 

 
 <a name="virtengine.inventory.v1.MemoryInfo"></a>

 ### MemoryInfo
 MemoryInfo reports Memory details

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `vendor` | [string](#string) |  |  |
 | `type` | [string](#string) |  |  |
 | `total_size` | [string](#string) |  |  |
 | `speed` | [string](#string) |  |  |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/inventory/v1/resources.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/inventory/v1/resources.proto
 

 
 <a name="virtengine.inventory.v1.NodeResources"></a>

 ### NodeResources
 NodeResources reports node inventory details

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `cpu` | [CPU](#virtengine.inventory.v1.CPU) |  |  |
 | `memory` | [Memory](#virtengine.inventory.v1.Memory) |  |  |
 | `gpu` | [GPU](#virtengine.inventory.v1.GPU) |  |  |
 | `ephemeral_storage` | [ResourcePair](#virtengine.inventory.v1.ResourcePair) |  |  |
 | `volumes_attached` | [ResourcePair](#virtengine.inventory.v1.ResourcePair) |  |  |
 | `volumes_mounted` | [ResourcePair](#virtengine.inventory.v1.ResourcePair) |  |  |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/inventory/v1/node.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/inventory/v1/node.proto
 

 
 <a name="virtengine.inventory.v1.Node"></a>

 ### Node
 Node reports node inventory details

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `name` | [string](#string) |  |  |
 | `resources` | [NodeResources](#virtengine.inventory.v1.NodeResources) |  |  |
 | `capabilities` | [NodeCapabilities](#virtengine.inventory.v1.NodeCapabilities) |  |  |
 
 

 

 
 <a name="virtengine.inventory.v1.NodeCapabilities"></a>

 ### NodeCapabilities
 NodeCapabilities extended list of node capabilities

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `storage_classes` | [string](#string) | repeated |  |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/inventory/v1/storage.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/inventory/v1/storage.proto
 

 
 <a name="virtengine.inventory.v1.Storage"></a>

 ### Storage
 Storage reports Storage inventory details

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `quantity` | [ResourcePair](#virtengine.inventory.v1.ResourcePair) |  |  |
 | `info` | [StorageInfo](#virtengine.inventory.v1.StorageInfo) |  |  |
 
 

 

 
 <a name="virtengine.inventory.v1.StorageInfo"></a>

 ### StorageInfo
 StorageInfo reports Storage details

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `class` | [string](#string) |  |  |
 | `iops` | [string](#string) |  |  |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/inventory/v1/cluster.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/inventory/v1/cluster.proto
 

 
 <a name="virtengine.inventory.v1.Cluster"></a>

 ### Cluster
 Cluster reports inventory across entire cluster.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `nodes` | [Node](#virtengine.inventory.v1.Node) | repeated |  |
 | `storage` | [Storage](#virtengine.inventory.v1.Storage) | repeated |  |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/inventory/v1/service.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/inventory/v1/service.proto
 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.inventory.v1.ClusterRPC"></a>

 ### ClusterRPC
 ClusterRPC defines the RPC server of cluster

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `QueryCluster` | [.google.protobuf.Empty](#google.protobuf.Empty) | [Cluster](#virtengine.inventory.v1.Cluster) | QueryCluster defines a method to query hardware state of the cluster buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/v1/inventory|
 | `StreamCluster` | [.google.protobuf.Empty](#google.protobuf.Empty) | [Cluster](#virtengine.inventory.v1.Cluster) stream | StreamCluster defines a method to stream hardware state of the cluster buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | |
 
 
 <a name="virtengine.inventory.v1.NodeRPC"></a>

 ### NodeRPC
 NodeRPC defines the RPC server of node

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `QueryNode` | [.google.protobuf.Empty](#google.protobuf.Empty) | [Node](#virtengine.inventory.v1.Node) | QueryNode defines a method to query hardware state of the node buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/v1/node|
 | `StreamNode` | [.google.protobuf.Empty](#google.protobuf.Empty) | [Node](#virtengine.inventory.v1.Node) stream | StreamNode defines a method to stream hardware state of the node buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | |
 
  <!-- end services -->

 
 
 <a name="virtengine/manifest/v2beta3/httpoptions.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/manifest/v2beta3/httpoptions.proto
 

 
 <a name="virtengine.manifest.v2beta3.ServiceExposeHTTPOptions"></a>

 ### ServiceExposeHTTPOptions
 ServiceExposeHTTPOptions

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `max_body_size` | [uint32](#uint32) |  |  |
 | `read_timeout` | [uint32](#uint32) |  |  |
 | `send_timeout` | [uint32](#uint32) |  |  |
 | `next_tries` | [uint32](#uint32) |  |  |
 | `next_timeout` | [uint32](#uint32) |  |  |
 | `next_cases` | [string](#string) | repeated |  |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/manifest/v2beta3/serviceexpose.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/manifest/v2beta3/serviceexpose.proto
 

 
 <a name="virtengine.manifest.v2beta3.ServiceExpose"></a>

 ### ServiceExpose
 ServiceExpose stores exposed ports and hosts details

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `port` | [uint32](#uint32) |  | port on the container |
 | `external_port` | [uint32](#uint32) |  | port on the service definition |
 | `proto` | [string](#string) |  |  |
 | `service` | [string](#string) |  |  |
 | `global` | [bool](#bool) |  |  |
 | `hosts` | [string](#string) | repeated |  |
 | `http_options` | [ServiceExposeHTTPOptions](#virtengine.manifest.v2beta3.ServiceExposeHTTPOptions) |  |  |
 | `ip` | [string](#string) |  | The name of the IP address associated with this, if any |
 | `endpoint_sequence_number` | [uint32](#uint32) |  | The sequence number of the associated endpoint in the on-chain data |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/manifest/v2beta3/service.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/manifest/v2beta3/service.proto
 

 
 <a name="virtengine.manifest.v2beta3.ImageCredentials"></a>

 ### ImageCredentials
 Credentials to fetch image from registry

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `host` | [string](#string) |  |  |
 | `email` | [string](#string) |  |  |
 | `username` | [string](#string) |  |  |
 | `password` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.manifest.v2beta3.Service"></a>

 ### Service
 Service stores name, image, args, env, unit, count and expose list of service

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `name` | [string](#string) |  |  |
 | `image` | [string](#string) |  |  |
 | `command` | [string](#string) | repeated |  |
 | `args` | [string](#string) | repeated |  |
 | `env` | [string](#string) | repeated |  |
 | `resources` | [virtengine.base.resources.v1beta4.Resources](#virtengine.base.resources.v1beta4.Resources) |  |  |
 | `count` | [uint32](#uint32) |  |  |
 | `expose` | [ServiceExpose](#virtengine.manifest.v2beta3.ServiceExpose) | repeated |  |
 | `params` | [ServiceParams](#virtengine.manifest.v2beta3.ServiceParams) |  |  |
 | `credentials` | [ImageCredentials](#virtengine.manifest.v2beta3.ImageCredentials) |  |  |
 
 

 

 
 <a name="virtengine.manifest.v2beta3.ServiceParams"></a>

 ### ServiceParams
 ServiceParams

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `storage` | [StorageParams](#virtengine.manifest.v2beta3.StorageParams) | repeated |  |
 | `credentials` | [ImageCredentials](#virtengine.manifest.v2beta3.ImageCredentials) |  |  |
 
 

 

 
 <a name="virtengine.manifest.v2beta3.StorageParams"></a>

 ### StorageParams
 StorageParams

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `name` | [string](#string) |  |  |
 | `mount` | [string](#string) |  |  |
 | `read_only` | [bool](#bool) |  |  |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/manifest/v2beta3/group.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/manifest/v2beta3/group.proto
 

 
 <a name="virtengine.manifest.v2beta3.Group"></a>

 ### Group
 Group store name and list of services

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `name` | [string](#string) |  |  |
 | `services` | [Service](#virtengine.manifest.v2beta3.Service) | repeated |  |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/provider/lease/v1/service.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/provider/lease/v1/service.proto
 

 
 <a name="virtengine.provider.lease.v1.ForwarderPortStatus"></a>

 ### ForwarderPortStatus
 ForwarderPortStatus

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `host` | [string](#string) |  |  |
 | `port` | [uint32](#uint32) |  |  |
 | `external_port` | [uint32](#uint32) |  |  |
 | `proto` | [string](#string) |  |  |
 | `name` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.provider.lease.v1.LeaseIPStatus"></a>

 ### LeaseIPStatus
 LeaseIPStatus

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `port` | [uint32](#uint32) |  |  |
 | `external_port` | [uint32](#uint32) |  |  |
 | `protocol` | [string](#string) |  |  |
 | `ip` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.provider.lease.v1.LeaseServiceStatus"></a>

 ### LeaseServiceStatus
 LeaseServiceStatus

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `available` | [int32](#int32) |  |  |
 | `total` | [int32](#int32) |  |  |
 | `uris` | [string](#string) | repeated |  |
 | `observed_generation` | [int64](#int64) |  |  |
 | `replicas` | [int32](#int32) |  |  |
 | `updated_replicas` | [int32](#int32) |  |  |
 | `ready_replicas` | [int32](#int32) |  |  |
 | `available_replicas` | [int32](#int32) |  |  |
 
 

 

 
 <a name="virtengine.provider.lease.v1.SendManifestRequest"></a>

 ### SendManifestRequest
 SendManifestRequest is request type for the SendManifest Providers RPC method

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `lease_id` | [virtengine.market.v1.LeaseID](#virtengine.market.v1.LeaseID) |  |  |
 | `manifest` | [virtengine.manifest.v2beta3.Group](#virtengine.manifest.v2beta3.Group) | repeated |  |
 
 

 

 
 <a name="virtengine.provider.lease.v1.SendManifestResponse"></a>

 ### SendManifestResponse
 SendManifestResponse is response type for the SendManifest Providers RPC method

 

 

 
 <a name="virtengine.provider.lease.v1.ServiceLogs"></a>

 ### ServiceLogs
 ServiceLogs

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `name` | [string](#string) |  |  |
 | `logs` | [bytes](#bytes) |  |  |
 
 

 

 
 <a name="virtengine.provider.lease.v1.ServiceLogsRequest"></a>

 ### ServiceLogsRequest
 ServiceLogsRequest

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `lease_id` | [virtengine.market.v1.LeaseID](#virtengine.market.v1.LeaseID) |  |  |
 | `services` | [string](#string) | repeated |  |
 
 

 

 
 <a name="virtengine.provider.lease.v1.ServiceLogsResponse"></a>

 ### ServiceLogsResponse
 ServiceLogsResponse

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `services` | [ServiceLogs](#virtengine.provider.lease.v1.ServiceLogs) | repeated |  |
 
 

 

 
 <a name="virtengine.provider.lease.v1.ServiceStatus"></a>

 ### ServiceStatus
 ServiceStatus

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `name` | [string](#string) |  |  |
 | `status` | [LeaseServiceStatus](#virtengine.provider.lease.v1.LeaseServiceStatus) |  |  |
 | `ports` | [ForwarderPortStatus](#virtengine.provider.lease.v1.ForwarderPortStatus) | repeated |  |
 | `ips` | [LeaseIPStatus](#virtengine.provider.lease.v1.LeaseIPStatus) | repeated |  |
 
 

 

 
 <a name="virtengine.provider.lease.v1.ServiceStatusRequest"></a>

 ### ServiceStatusRequest
 ServiceStatusRequest

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `lease_id` | [virtengine.market.v1.LeaseID](#virtengine.market.v1.LeaseID) |  |  |
 | `services` | [string](#string) | repeated |  |
 
 

 

 
 <a name="virtengine.provider.lease.v1.ServiceStatusResponse"></a>

 ### ServiceStatusResponse
 ServiceStatusResponse

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `services` | [ServiceStatus](#virtengine.provider.lease.v1.ServiceStatus) | repeated |  |
 
 

 

 
 <a name="virtengine.provider.lease.v1.ShellRequest"></a>

 ### ShellRequest
 ShellRequest

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `lease_id` | [virtengine.market.v1.LeaseID](#virtengine.market.v1.LeaseID) |  |  |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.provider.lease.v1.LeaseRPC"></a>

 ### LeaseRPC
 LeaseRPC defines the RPC server for lease control

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `SendManifest` | [SendManifestRequest](#virtengine.provider.lease.v1.SendManifestRequest) | [SendManifestResponse](#virtengine.provider.lease.v1.SendManifestResponse) | SendManifest sends manifest to the provider | |
 | `ServiceStatus` | [ServiceStatusRequest](#virtengine.provider.lease.v1.ServiceStatusRequest) | [ServiceStatusResponse](#virtengine.provider.lease.v1.ServiceStatusResponse) | ServiceStatus buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | |
 | `StreamServiceStatus` | [ServiceStatusRequest](#virtengine.provider.lease.v1.ServiceStatusRequest) | [ServiceStatusResponse](#virtengine.provider.lease.v1.ServiceStatusResponse) stream | StreamServiceStatus buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | |
 | `ServiceLogs` | [ServiceLogsRequest](#virtengine.provider.lease.v1.ServiceLogsRequest) | [ServiceLogsResponse](#virtengine.provider.lease.v1.ServiceLogsResponse) | ServiceLogs buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | |
 | `StreamServiceLogs` | [ServiceLogsRequest](#virtengine.provider.lease.v1.ServiceLogsRequest) | [ServiceLogsResponse](#virtengine.provider.lease.v1.ServiceLogsResponse) stream | StreamServiceLogs buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | |
 
  <!-- end services -->

 
 
 <a name="virtengine/provider/v1/status.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/provider/v1/status.proto
 

 
 <a name="virtengine.provider.v1.BidEngineStatus"></a>

 ### BidEngineStatus
 BidEngineStatus

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `orders` | [uint32](#uint32) |  |  |
 
 

 

 
 <a name="virtengine.provider.v1.ClusterStatus"></a>

 ### ClusterStatus
 ClusterStatus

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `leases` | [Leases](#virtengine.provider.v1.Leases) |  |  |
 | `inventory` | [Inventory](#virtengine.provider.v1.Inventory) |  |  |
 
 

 

 
 <a name="virtengine.provider.v1.Inventory"></a>

 ### Inventory
 Inventory

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `cluster` | [virtengine.inventory.v1.Cluster](#virtengine.inventory.v1.Cluster) |  |  |
 | `reservations` | [Reservations](#virtengine.provider.v1.Reservations) |  |  |
 
 

 

 
 <a name="virtengine.provider.v1.Leases"></a>

 ### Leases
 Leases

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `active` | [uint32](#uint32) |  |  |
 
 

 

 
 <a name="virtengine.provider.v1.ManifestStatus"></a>

 ### ManifestStatus
 ManifestStatus

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `deployments` | [uint32](#uint32) |  |  |
 
 

 

 
 <a name="virtengine.provider.v1.Reservations"></a>

 ### Reservations
 Reservations

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `pending` | [ReservationsMetric](#virtengine.provider.v1.ReservationsMetric) |  |  |
 | `active` | [ReservationsMetric](#virtengine.provider.v1.ReservationsMetric) |  |  |
 
 

 

 
 <a name="virtengine.provider.v1.ReservationsMetric"></a>

 ### ReservationsMetric
 ReservationsMetric

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `count` | [uint32](#uint32) |  |  |
 | `resources` | [ResourcesMetric](#virtengine.provider.v1.ResourcesMetric) |  |  |
 
 

 

 
 <a name="virtengine.provider.v1.ResourcesMetric"></a>

 ### ResourcesMetric
 ResourceMetrics

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `cpu` | [k8s.io.apimachinery.pkg.api.resource.Quantity](#k8s.io.apimachinery.pkg.api.resource.Quantity) |  |  |
 | `memory` | [k8s.io.apimachinery.pkg.api.resource.Quantity](#k8s.io.apimachinery.pkg.api.resource.Quantity) |  |  |
 | `gpu` | [k8s.io.apimachinery.pkg.api.resource.Quantity](#k8s.io.apimachinery.pkg.api.resource.Quantity) |  |  |
 | `ephemeral_storage` | [k8s.io.apimachinery.pkg.api.resource.Quantity](#k8s.io.apimachinery.pkg.api.resource.Quantity) |  |  |
 | `storage` | [ResourcesMetric.StorageEntry](#virtengine.provider.v1.ResourcesMetric.StorageEntry) | repeated |  |
 
 

 

 
 <a name="virtengine.provider.v1.ResourcesMetric.StorageEntry"></a>

 ### ResourcesMetric.StorageEntry
 

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `key` | [string](#string) |  |  |
 | `value` | [k8s.io.apimachinery.pkg.api.resource.Quantity](#k8s.io.apimachinery.pkg.api.resource.Quantity) |  |  |
 
 

 

 
 <a name="virtengine.provider.v1.Status"></a>

 ### Status
 Status

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `errors` | [string](#string) | repeated |  |
 | `cluster` | [ClusterStatus](#virtengine.provider.v1.ClusterStatus) |  |  |
 | `bid_engine` | [BidEngineStatus](#virtengine.provider.v1.BidEngineStatus) |  |  |
 | `manifest` | [ManifestStatus](#virtengine.provider.v1.ManifestStatus) |  |  |
 | `public_hostnames` | [string](#string) | repeated |  |
 | `timestamp` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/provider/v1/service.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/provider/v1/service.proto
 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.provider.v1.ProviderRPC"></a>

 ### ProviderRPC
 ProviderRPC defines the RPC server for provider

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `GetStatus` | [.google.protobuf.Empty](#google.protobuf.Empty) | [Status](#virtengine.provider.v1.Status) | GetStatus defines a method to query provider state buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/v1/status|
 | `StreamStatus` | [.google.protobuf.Empty](#google.protobuf.Empty) | [Status](#virtengine.provider.v1.Status) stream | Status defines a method to stream provider state buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | |
 
  <!-- end services -->

 

 ## Scalar Value Types

 | .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
 | ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
 | <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
 | <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
 | <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
 | <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
 | <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
 | <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
 | <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
 | <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
 | <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
 | <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
 | <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
 | <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
 | <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
 | <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
 | <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |
 
