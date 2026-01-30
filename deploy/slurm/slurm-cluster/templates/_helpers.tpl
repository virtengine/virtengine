{{/*
Expand the name of the chart.
*/}}
{{- define "slurm-cluster.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "slurm-cluster.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
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
SLURM version tag
*/}}
{{- define "slurm-cluster.slurmVersion" -}}
{{- default .Chart.AppVersion .Values.global.slurmVersion }}
{{- end }}

{{/*
Image registry prefix
*/}}
{{- define "slurm-cluster.imageRegistry" -}}
{{- .Values.global.imageRegistry | default "ghcr.io/virtengine" }}
{{- end }}

{{/*
Controller image
*/}}
{{- define "slurm-cluster.controller.image" -}}
{{- $registry := include "slurm-cluster.imageRegistry" . }}
{{- $tag := default (include "slurm-cluster.slurmVersion" .) .Values.controller.image.tag }}
{{- printf "%s/%s:%s" $registry .Values.controller.image.repository $tag }}
{{- end }}

{{/*
Database image
*/}}
{{- define "slurm-cluster.database.image" -}}
{{- $registry := include "slurm-cluster.imageRegistry" . }}
{{- $tag := default (include "slurm-cluster.slurmVersion" .) .Values.database.image.tag }}
{{- printf "%s/%s:%s" $registry .Values.database.image.repository $tag }}
{{- end }}

{{/*
Compute image
*/}}
{{- define "slurm-cluster.compute.image" -}}
{{- $registry := include "slurm-cluster.imageRegistry" . }}
{{- $tag := default (include "slurm-cluster.slurmVersion" .) .Values.compute.image.tag }}
{{- printf "%s/%s:%s" $registry .Values.compute.image.repository $tag }}
{{- end }}

{{/*
Munge image
*/}}
{{- define "slurm-cluster.munge.image" -}}
{{- $registry := include "slurm-cluster.imageRegistry" . }}
{{- $tag := default (include "slurm-cluster.slurmVersion" .) .Values.munge.image.tag }}
{{- printf "%s/%s:%s" $registry .Values.munge.image.repository $tag }}
{{- end }}

{{/*
Node agent image
*/}}
{{- define "slurm-cluster.nodeAgent.image" -}}
{{- $registry := include "slurm-cluster.imageRegistry" . }}
{{- printf "%s/%s:%s" $registry .Values.nodeAgent.image.repository .Values.nodeAgent.image.tag }}
{{- end }}

{{/*
Munge secret name
*/}}
{{- define "slurm-cluster.munge.secretName" -}}
{{- if .Values.munge.existingSecret }}
{{- .Values.munge.existingSecret }}
{{- else }}
{{- include "slurm-cluster.fullname" . }}-munge
{{- end }}
{{- end }}

{{/*
Database secret name
*/}}
{{- define "slurm-cluster.database.secretName" -}}
{{- if .Values.database.config.existingSecret }}
{{- .Values.database.config.existingSecret }}
{{- else }}
{{- include "slurm-cluster.fullname" . }}-db
{{- end }}
{{- end }}

{{/*
MariaDB secret name
*/}}
{{- define "slurm-cluster.mariadb.secretName" -}}
{{- if .Values.mariadb.existingSecret }}
{{- .Values.mariadb.existingSecret }}
{{- else }}
{{- include "slurm-cluster.fullname" . }}-mariadb
{{- end }}
{{- end }}

{{/*
Controller service name
*/}}
{{- define "slurm-cluster.controller.serviceName" -}}
{{- include "slurm-cluster.fullname" . }}-controller
{{- end }}

{{/*
Database service name
*/}}
{{- define "slurm-cluster.database.serviceName" -}}
{{- include "slurm-cluster.fullname" . }}-slurmdbd
{{- end }}

{{/*
MariaDB service name
*/}}
{{- define "slurm-cluster.mariadb.serviceName" -}}
{{- include "slurm-cluster.fullname" . }}-mariadb
{{- end }}

{{/*
Compute headless service name
*/}}
{{- define "slurm-cluster.compute.serviceName" -}}
{{- include "slurm-cluster.fullname" . }}-compute
{{- end }}

{{/*
Generate node list for SLURM configuration
*/}}
{{- define "slurm-cluster.nodeList" -}}
{{- $fullname := include "slurm-cluster.fullname" . }}
{{- $replicas := int .Values.compute.replicas }}
{{- if eq $replicas 1 }}
{{- printf "%s-compute-0" $fullname }}
{{- else }}
{{- printf "%s-compute-[0-%d]" $fullname (sub $replicas 1) }}
{{- end }}
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
