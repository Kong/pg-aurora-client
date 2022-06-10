{{/*
Expand the name of the chart.
*/}}
{{- define "pg-aurora-client.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "pg-aurora-client.fullname" -}}
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
Base labels
*/}}
{{- define "pg-aurora-client.base.labels" -}}
helm.sh/chart: {{ include "pg-aurora-client.chart" . }}
{{ include "pg-aurora-client.base.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Base selector labels
*/}}
{{- define "pg-aurora-client.base.selectorLabels" -}}
app.kubernetes.io/name: {{ include "pg-aurora-client.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
pg-aurora-client canary labels
*/}}
{{- define "pg-aurora-client.labels" -}}
{{ include "pg-aurora-client.base.labels" . }}
app.kubernetes.io/component: canary
{{- end }}

{{/*
pg-aurora-client canary selector labels
*/}}
{{- define "pg-aurora-client.selectorLabels" -}}
{{ include "pg-aurora-client.base.selectorLabels" . }}
app.kubernetes.io/component: canary
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "pg-aurora-client.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create database connection env vars
*/}}
{{- define "pg-aurora-client.dbEnv" -}}
- name: "PG_DATABASE"
  value: {{ .Values.database.db_name }}
- name: "PG_HOST"
  value: {{ .Values.database.hosts.rw }}
{{- if .Values.database.hosts.ro }}
- name: "PG_RO_HOST"
  value: {{ .Values.database.hosts.ro -}}
{{ end }}
- name: "PG_PORT"
  value: {{ quote .Values.database.port }}
- name: "PG_USER"
  value: {{ .Values.database.username }}
- name: "PG_PASSWORD"
  valueFrom:
    secretKeyRef:
      key: password
      name: {{ .Values.database.secret_name }}
{{- end }}