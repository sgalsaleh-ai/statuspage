{{- define "statuspage.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "statuspage.fullname" -}}
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

{{- define "statuspage.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
app.kubernetes.io/name: {{ include "statuspage.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "statuspage.selectorLabels" -}}
app.kubernetes.io/name: {{ include "statuspage.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "statuspage.dbHost" -}}
{{- if .Values.postgresql.enabled }}
{{- printf "%s-postgresql" .Release.Name }}
{{- else }}
{{- .Values.externalPostgresql.host }}
{{- end }}
{{- end }}

{{- define "statuspage.dbPort" -}}
{{- if .Values.postgresql.enabled }}
{{- "5432" }}
{{- else }}
{{- .Values.externalPostgresql.port | default "5432" }}
{{- end }}
{{- end }}

{{- define "statuspage.dbUser" -}}
{{- if .Values.postgresql.enabled }}
{{- .Values.postgresql.auth.username }}
{{- else }}
{{- .Values.externalPostgresql.user }}
{{- end }}
{{- end }}

{{- define "statuspage.dbPassword" -}}
{{- if .Values.postgresql.enabled }}
{{- .Values.postgresql.auth.password }}
{{- else }}
{{- .Values.externalPostgresql.password }}
{{- end }}
{{- end }}

{{- define "statuspage.dbName" -}}
{{- if .Values.postgresql.enabled }}
{{- .Values.postgresql.auth.database }}
{{- else }}
{{- .Values.externalPostgresql.database }}
{{- end }}
{{- end }}

{{- define "statuspage.centrifugoURL" -}}
{{- if .Values.centrifugo.enabled }}
{{- printf "http://%s-centrifugo:%s" .Release.Name "9000" }}
{{- else }}
{{- .Values.externalCentrifugo.url }}
{{- end }}
{{- end }}

{{- define "statuspage.centrifugoPublicURL" -}}
{{- if .Values.centrifugo.enabled }}
{{- printf "http://%s-centrifugo:%s" .Release.Name "8000" }}
{{- else }}
{{- .Values.externalCentrifugo.url }}
{{- end }}
{{- end }}

{{- define "statuspage.centrifugoAPIKey" -}}
{{- if .Values.centrifugo.enabled }}
{{- .Values.centrifugo.config.api_key | default "statuspage-api-key" }}
{{- else }}
{{- .Values.externalCentrifugo.apiKey }}
{{- end }}
{{- end }}

{{- define "statuspage.tlsSecretName" -}}
{{- if eq .Values.tls.mode "manual" }}
{{- .Values.tls.secretName }}
{{- else if eq .Values.tls.mode "selfSigned" }}
{{- printf "%s-tls" (include "statuspage.fullname" .) }}
{{- else if eq .Values.tls.mode "auto" }}
{{- printf "%s-tls" (include "statuspage.fullname" .) }}
{{- else }}
{{- "" }}
{{- end }}
{{- end }}
