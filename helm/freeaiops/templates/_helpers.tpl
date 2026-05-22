{{/*
Expand the name of the chart.
*/}}
{{- define "freeaiops.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "freeaiops.fullname" -}}
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

{{- define "freeaiops.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "freeaiops.labels" -}}
helm.sh/chart: {{ include "freeaiops.chart" . }}
{{ include "freeaiops.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "freeaiops.selectorLabels" -}}
app.kubernetes.io/name: {{ include "freeaiops.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "freeaiops.appLabels" -}}
{{ include "freeaiops.selectorLabels" . }}
app.kubernetes.io/component: app
{{- end }}

{{- define "freeaiops.mysqlLabels" -}}
{{ include "freeaiops.selectorLabels" . }}
app.kubernetes.io/component: mysql
{{- end }}

{{- define "freeaiops.mysql.fullname" -}}
{{- printf "%s-mysql" (include "freeaiops.fullname" .) }}
{{- end }}

{{- define "freeaiops.mysql.host" -}}
{{- if .Values.mysql.enabled }}
{{- include "freeaiops.mysql.fullname" . }}
{{- else }}
{{- required "externalMysql.host is required when mysql.enabled=false" .Values.externalMysql.host }}
{{- end }}
{{- end }}

{{- define "freeaiops.mysql.password" -}}
{{- if .Values.mysql.enabled }}
{{- .Values.mysql.auth.rootPassword }}
{{- else }}
{{- required "externalMysql.password is required when mysql.enabled=false" .Values.externalMysql.password }}
{{- end }}
{{- end }}

{{- define "freeaiops.mysql.username" -}}
{{- if .Values.mysql.enabled }}
{{- .Values.mysql.auth.username }}
{{- else }}
{{- .Values.externalMysql.username }}
{{- end }}
{{- end }}

{{- define "freeaiops.mysql.database" -}}
{{- if .Values.mysql.enabled }}
{{- .Values.mysql.auth.database }}
{{- else }}
{{- .Values.externalMysql.database }}
{{- end }}
{{- end }}

{{- define "freeaiops.mysql.port" -}}
{{- if .Values.mysql.enabled }}
{{- .Values.mysql.service.port | quote }}
{{- else }}
{{- .Values.externalMysql.port | quote }}
{{- end }}
{{- end }}
