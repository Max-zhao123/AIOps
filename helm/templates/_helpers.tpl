{{/*
Expand the name of the chart.
*/}}
{{- define "aiops.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "aiops.fullname" -}}
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

{{- define "aiops.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "aiops.labels" -}}
helm.sh/chart: {{ include "aiops.chart" . }}
{{ include "aiops.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "aiops.selectorLabels" -}}
app.kubernetes.io/name: {{ include "aiops.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "aiops.mysqlLabels" -}}
{{ include "aiops.selectorLabels" . }}
app.kubernetes.io/component: mysql
{{- end }}

{{- define "aiops.mysql.fullname" -}}
{{- printf "%s-mysql" (include "aiops.fullname" .) }}
{{- end }}

{{- define "aiops.mysql.host" -}}
{{- if .Values.mysql.enabled }}
{{- include "aiops.mysql.fullname" . }}
{{- else }}
{{- required "externalMysql.host is required when mysql.enabled=false" .Values.externalMysql.host }}
{{- end }}
{{- end }}

{{- define "aiops.mysql.password" -}}
{{- if .Values.mysql.enabled }}
{{- .Values.mysql.auth.rootPassword }}
{{- else }}
{{- required "externalMysql.password is required when mysql.enabled=false" .Values.externalMysql.password }}
{{- end }}
{{- end }}

{{- define "aiops.mysql.username" -}}
{{- if .Values.mysql.enabled }}
{{- .Values.mysql.auth.username }}
{{- else }}
{{- .Values.externalMysql.username }}
{{- end }}
{{- end }}

{{- define "aiops.mysql.database" -}}
{{- if .Values.mysql.enabled }}
{{- .Values.mysql.auth.database }}
{{- else }}
{{- .Values.externalMysql.database }}
{{- end }}
{{- end }}

{{- define "aiops.mysql.port" -}}
{{- if .Values.mysql.enabled }}
{{- .Values.mysql.service.port | quote }}
{{- else }}
{{- .Values.externalMysql.port | quote }}
{{- end }}
{{- end }}

{{- define "aiops.mysql.pvcName" -}}
{{- if .Values.mysql.persistence.existingClaim }}
{{- .Values.mysql.persistence.existingClaim }}
{{- else }}
{{- include "aiops.mysql.fullname" . }}
{{- end }}
{{- end }}

{{- define "aiops.imagePullSecrets" -}}
{{- with .Values.imagePullSecrets }}
imagePullSecrets:
  {{- toYaml . | nindent 2 }}
{{- end }}
{{- end }}

{{- define "aiops.image" -}}
{{- $registry := .Values.imageRegistry -}}
{{- $name := .svcName -}}
{{- printf "%s%s:%s" $registry $name $.Values.imageTag -}}
{{- end }}

{{- define "aiops.msLabels" -}}
app.kubernetes.io/name: {{ .svc.name }}
app.kubernetes.io/instance: {{ .root.Release.Name }}
app.kubernetes.io/component: {{ .key }}
{{- end }}
