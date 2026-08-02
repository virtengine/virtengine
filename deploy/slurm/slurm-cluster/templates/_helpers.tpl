{{/*
Expand the name of the chart.
*/}}
{{- define "slurm-cluster.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create the unbounded fully qualified app name used as input to stable DNS names.
*/}}
{{- define "slurm-cluster.rawFullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Bound a DNS label, preserving collision resistance when truncation is required.
*/}}
{{- define "slurm-cluster.dnsName" -}}
{{- $raw := index . 0 -}}
{{- $limit := int (index . 1) -}}
{{- if lt $limit 10 -}}{{ fail (printf "DNS name budget %d is too small for stable truncation" $limit) }}{{- end -}}
{{- if le (len $raw) $limit -}}
{{- $raw | trimSuffix "-" -}}
{{- else -}}
{{- $hash := sha256sum $raw | trunc 8 -}}
{{- $prefix := $raw | trunc (sub $limit 9) | trimSuffix "-" -}}
{{- printf "%s-%s" $prefix $hash -}}
{{- end -}}
{{- end }}

{{/*
Create a DNS label for a resource and reserve room for a StatefulSet pod ordinal.
*/}}
{{- define "slurm-cluster.resourceName" -}}
{{- $root := index . 0 -}}
{{- $suffix := index . 1 -}}
{{- $replicas := int (index . 2) -}}
{{- $ordinal := 0 -}}
{{- if gt $replicas 1 -}}{{- $ordinal = sub $replicas 1 -}}{{- end -}}
{{- $ordinalBudget := 0 -}}
{{- if gt $replicas 0 -}}{{- $ordinalBudget = add 1 (len (printf "%d" $ordinal)) -}}{{- end -}}
{{- $raw := include "slurm-cluster.rawFullname" $root -}}
{{- if $suffix -}}{{- $raw = printf "%s-%s" $raw $suffix -}}{{- end -}}
{{- include "slurm-cluster.dnsName" (list $raw (sub 63 $ordinalBudget)) -}}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "slurm-cluster.fullname" -}}
{{- include "slurm-cluster.resourceName" (list . "" 0) -}}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "slurm-cluster.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "slurm-cluster.labels" -}}
helm.sh/chart: {{ include "slurm-cluster.chart" . }}
{{ include "slurm-cluster.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
virtengine.com/component: hpc
virtengine.com/module: slurm
{{- if .Values.cluster.id }}
virtengine.com/cluster-id: {{ .Values.cluster.id | quote }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "slurm-cluster.selectorLabels" -}}
app.kubernetes.io/name: {{ include "slurm-cluster.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "slurm-cluster.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "slurm-cluster.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Require one complete, lowercase, digest-pinned image reference.
*/}}
{{- define "slurm-cluster.immutableImage" -}}
{{- $name := index . 0 -}}
{{- $reference := required (printf "%s.image.reference must be an explicit repository@sha256:<lowercase64> reference" $name) (index . 1) -}}
{{- if not (regexMatch "^[a-z0-9]+([._-][a-z0-9]+)*(:[0-9]+)?(/[a-z0-9]+([._-][a-z0-9]+)*)*@sha256:[a-f0-9]{64}$" $reference) -}}
{{- fail (printf "%s.image.reference must be an explicit repository@sha256:<lowercase64> reference" $name) -}}
{{- end -}}
{{- $reference -}}
{{- end }}

{{- define "slurm-cluster.controller.image" -}}
{{- include "slurm-cluster.immutableImage" (list "controller" .Values.controller.image.reference) -}}
{{- end }}

{{- define "slurm-cluster.database.image" -}}
{{- include "slurm-cluster.immutableImage" (list "database" .Values.database.image.reference) -}}
{{- end }}

{{- define "slurm-cluster.compute.image" -}}
{{- include "slurm-cluster.immutableImage" (list "compute" .Values.compute.image.reference) -}}
{{- end }}

{{- define "slurm-cluster.munge.image" -}}
{{- include "slurm-cluster.immutableImage" (list "munge" .Values.munge.image.reference) -}}
{{- end }}

{{- define "slurm-cluster.mariadb.image" -}}
{{- include "slurm-cluster.immutableImage" (list "mariadb" .Values.mariadb.image.reference) -}}
{{- end }}

{{- define "slurm-cluster.nodeAgent.image" -}}
{{- include "slurm-cluster.immutableImage" (list "nodeAgent" .Values.nodeAgent.image.reference) -}}
{{- end }}

{{- define "slurm-cluster.utility.image" -}}
{{- include "slurm-cluster.immutableImage" (list "utility" .Values.utilityImage.reference) -}}
{{- end }}

{{/*
Fixed pod and container controls with explicit image identity contracts.
*/}}
{{- define "slurm-cluster.podSecurityContext" -}}
{{- $root := index . 0 -}}
fsGroup: {{ $root.Values.securityIdentities.munge.gid }}
fsGroupChangePolicy: OnRootMismatch
runAsNonRoot: true
seccompProfile:
	type: RuntimeDefault
{{- end }}

{{- define "slurm-cluster.containerSecurityContext" -}}
{{- $root := index . 0 -}}
{{- $component := index . 1 -}}
{{- $identityKey := $component -}}
{{- if has $component (list "controller" "database" "compute") -}}
{{- $identityKey = "slurm" -}}
{{- end -}}
{{- $identity := required (printf "securityIdentities.%s is required" $identityKey) (index $root.Values.securityIdentities $identityKey) -}}
allowPrivilegeEscalation: false
capabilities:
	drop:
		- ALL
privileged: false
readOnlyRootFilesystem: true
runAsGroup: {{ required (printf "securityIdentities.%s.gid is required" $component) $identity.gid }}
runAsNonRoot: true
runAsUser: {{ required (printf "securityIdentities.%s.uid is required" $component) $identity.uid }}
seccompProfile:
	type: RuntimeDefault
{{- end }}

{{/* Shared daemon username for slurmctld, slurmdbd, and slurmd. */}}
{{- define "slurm-cluster.slurmUser" -}}
{{- required "securityIdentities.slurm.username is required" .Values.securityIdentities.slurm.username -}}
{{- end }}

{{/*
Munge secret name
*/}}
{{- define "slurm-cluster.munge.secretName" -}}
{{- required "munge.existingSecret is required when munge.enabled=true; provision the Secret before installing" .Values.munge.existingSecret }}
{{- end }}

{{/*
Database secret name
*/}}
{{- define "slurm-cluster.database.secretName" -}}
{{- required "database.config.existingSecret is required when database.enabled=true; provision the Secret before installing" .Values.database.config.existingSecret }}
{{- end }}

{{/*
MariaDB secret name
*/}}
{{- define "slurm-cluster.mariadb.secretName" -}}
{{- required "mariadb.existingSecret is required when mariadb.enabled=true; provision the Secret before installing" .Values.mariadb.existingSecret }}
{{- end }}

{{/*
Controller service name
*/}}
{{- define "slurm-cluster.controller.serviceName" -}}
{{- include "slurm-cluster.resourceName" (list . "controller" .Values.controller.replicas) -}}
{{- end }}

{{/*
Database service name
*/}}
{{- define "slurm-cluster.database.serviceName" -}}
{{- include "slurm-cluster.resourceName" (list . "slurmdbd" .Values.database.replicas) -}}
{{- end }}

{{/*
MariaDB service name
*/}}
{{- define "slurm-cluster.mariadb.serviceName" -}}
{{- include "slurm-cluster.resourceName" (list . "mariadb" 1) -}}
{{- end }}

{{/*
Compute headless service name
*/}}
{{- define "slurm-cluster.compute.serviceName" -}}
{{- include "slurm-cluster.resourceName" (list . "compute" .Values.compute.replicas) -}}
{{- end }}

{{/*
Node pool headless service name
*/}}
{{- define "slurm-cluster.nodePool.serviceName" -}}
{{- $root := index . 0 -}}
{{- $pool := index . 1 -}}
{{- include "slurm-cluster.resourceName" (list $root $pool.name $pool.replicas) -}}
{{- end }}

{{/*
Whether a node pool is enabled. Pools are enabled by default for compatibility.
*/}}
{{- define "slurm-cluster.nodePool.enabled" -}}
{{- $pool := index . 1 -}}
{{- if or (not (hasKey $pool "enabled")) $pool.enabled -}}true{{- else -}}false{{- end -}}
{{- end }}

{{/*
Build the authoritative enabled compute capacity and SLURM hostlist.
*/}}
{{- define "slurm-cluster.compute.capacity" -}}
{{- $root := . -}}
{{- $total := 0 -}}
{{- $nodes := list -}}
{{- $names := dict -}}
{{- $serviceNames := dict -}}
{{- $pools := dict -}}
{{- $reserved := list "controller" "slurmdbd" "database" "mariadb" "mariadb-init" "compute" "munge" "node-agent" "config" "default" -}}
{{- if .Values.compute.enabled -}}
{{- $replicas := int .Values.compute.replicas -}}
{{- if lt $replicas 1 -}}{{ fail "compute.replicas must be at least 1 when compute.enabled is true" }}{{- end -}}
{{- $_ := set $names "compute" true -}}
{{- $_ := set $serviceNames (include "slurm-cluster.compute.serviceName" .) true -}}
{{- $total = add $total $replicas -}}
{{- $defaultNodes := printf "%s-[0-%d]" (include "slurm-cluster.compute.serviceName" .) (sub $replicas 1) -}}
{{- $nodes = append $nodes $defaultNodes -}}
{{- $_ := set $pools "default" (dict "enabled" true "replicas" $replicas "nodes" $defaultNodes) -}}
{{- else -}}
{{- $_ := set $pools "default" (dict "enabled" false "replicas" 0 "nodes" "") -}}
{{- end -}}
{{- range $pool := .Values.nodePools -}}
{{- $name := required "nodePools[].name is required" $pool.name -}}
{{- if not (regexMatch "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$" $name) -}}{{ fail (printf "node pool name %q must be a DNS label" $name) }}{{- end -}}
{{- if has $name $reserved -}}{{ fail (printf "node pool name %q is reserved by an existing chart resource or component" $name) }}{{- end -}}
{{- if hasKey $names $name -}}{{ fail (printf "node pool name %q is duplicated or conflicts with the default compute pool" $name) }}{{- end -}}
{{- $_ := set $names $name true -}}
{{- $serviceName := include "slurm-cluster.nodePool.serviceName" (list $root $pool) -}}
{{- if hasKey $serviceNames $serviceName -}}{{ fail (printf "node pool name %q collides at rendered StatefulSet name %q" $name $serviceName) }}{{- end -}}
{{- $_ := set $serviceNames $serviceName true -}}
{{- $enabled := eq (include "slurm-cluster.nodePool.enabled" (list $root $pool)) "true" -}}
{{- if $enabled -}}
{{- $replicas := int $pool.replicas -}}
{{- if lt $replicas 1 -}}{{ fail (printf "node pool %q replicas must be at least 1 when enabled" $name) }}{{- end -}}
{{- $total = add $total $replicas -}}
{{- $poolNodes := printf "%s-[0-%d]" $serviceName (sub $replicas 1) -}}
{{- $nodes = append $nodes $poolNodes -}}
{{- $_ := set $pools $name (dict "enabled" true "replicas" $replicas "nodes" $poolNodes) -}}
{{- else -}}
{{- $_ := set $pools $name (dict "enabled" false "replicas" 0 "nodes" "") -}}
{{- end -}}
{{- end -}}
{{- if lt $total 1 -}}{{ fail "at least one compute replica must be enabled" }}{{- end -}}
{{- dict "replicas" $total "nodes" (join "," $nodes) "pools" $pools | toJson -}}
{{- end }}

{{/*
Resolve and validate the exact enabled node pools selected by one partition.
The default partition selects all capacity when nodePools is omitted.
*/}}
{{- define "slurm-cluster.partition.capacity" -}}
{{- $partition := index . 1 -}}
{{- $capacity := index . 2 -}}
{{- $selectors := $partition.nodePools | default (list) -}}
{{- if and (eq (len $selectors) 0) $partition.default -}}
{{- dict "replicas" $capacity.replicas "nodes" $capacity.nodes | toJson -}}
{{- else -}}
{{- if eq (len $selectors) 0 -}}{{ fail (printf "partition %q must select at least one node pool" $partition.name) }}{{- end -}}
{{- $seen := dict -}}
{{- $nodes := list -}}
{{- $replicas := 0 -}}
{{- range $selector := $selectors -}}
{{- if hasKey $seen $selector -}}{{ fail (printf "partition %q has duplicate node pool selector %q" $partition.name $selector) }}{{- end -}}
{{- $_ := set $seen $selector true -}}
{{- if not (hasKey $capacity.pools $selector) -}}{{ fail (printf "partition %q selects unknown node pool %q" $partition.name $selector) }}{{- end -}}
{{- $pool := index $capacity.pools $selector -}}
{{- if not $pool.enabled -}}{{ fail (printf "partition %q selects disabled node pool %q" $partition.name $selector) }}{{- end -}}
{{- $nodes = append $nodes $pool.nodes -}}
{{- $replicas = add $replicas (int $pool.replicas) -}}
{{- end -}}
{{- if lt $replicas 1 -}}{{ fail (printf "partition %q selects zero compute nodes" $partition.name) }}{{- end -}}
{{- dict "replicas" $replicas "nodes" (join "," $nodes) | toJson -}}
{{- end -}}
{{- end }}

{{/*
Storage class
*/}}
{{- define "slurm-cluster.storageClass" -}}
{{- if .storageClass }}
{{- .storageClass }}
{{- else if $.Values.global.storageClass }}
{{- $.Values.global.storageClass }}
{{- else }}
{{- "" }}
{{- end }}
{{- end }}
