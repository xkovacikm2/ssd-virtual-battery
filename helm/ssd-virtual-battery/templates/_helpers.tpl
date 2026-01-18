{{/*
Expand the name of the chart.
*/}}
{{- define "ssd-virtual-battery.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "ssd-virtual-battery.fullname" -}}
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
{{- define "ssd-virtual-battery.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "ssd-virtual-battery.labels" -}}
helm.sh/chart: {{ include "ssd-virtual-battery.chart" . }}
{{ include "ssd-virtual-battery.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "ssd-virtual-battery.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ssd-virtual-battery.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
PostgreSQL host
*/}}
{{- define "ssd-virtual-battery.postgresqlHost" -}}
{{- printf "%s-postgresql" (include "ssd-virtual-battery.fullname" .) }}
{{- end }}

{{/*
Database URL
*/}}
{{- define "ssd-virtual-battery.databaseUrl" -}}
{{- printf "postgresql://%s:$(POSTGRES_PASSWORD)@%s:5432/%s" .Values.postgresql.auth.username (include "ssd-virtual-battery.postgresqlHost" .) .Values.postgresql.auth.database }}
{{- end }}
