{{/*
VE-502: SLURM Cluster Helm Chart Templates

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
virtengine.com/cluster-name: {{ .Values.cluster.name | quote }}
virtengine.com/region: {{ .Values.cluster.region | quote }}
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
Controller selector labels
*/}}
{{- define "slurm-cluster.controller.selectorLabels" -}}
{{ include "slurm-cluster.selectorLabels" . }}
app.kubernetes.io/component: controller
{{- end }}

{{/*
Compute selector labels
*/}}
{{- define "slurm-cluster.compute.selectorLabels" -}}
{{ include "slurm-cluster.selectorLabels" . }}
app.kubernetes.io/component: compute
{{- end }}

{{/*
Database selector labels
*/}}
{{- define "slurm-cluster.database.selectorLabels" -}}
{{ include "slurm-cluster.selectorLabels" . }}
app.kubernetes.io/component: database
{{- end }}
