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
{{- printf "http://%s-centrifugo:%s" .Release.Name "9000" }}
{{- end }}

{{- define "statuspage.centrifugoPublicURL" -}}
{{- if .Values.app.centrifugoPublicURL }}
{{- .Values.app.centrifugoPublicURL }}
{{- else }}
{{- printf "http://%s-centrifugo:%s" .Release.Name "8000" }}
{{- end }}
{{- end }}

{{- define "statuspage.centrifugoAPIKey" -}}
{{- .Values.centrifugo.config.api_key | default "statuspage-api-key" }}
{{- end }}

{{- define "statuspage.image" -}}
{{- if .Values.image.registry }}
{{- printf "%s/%s:%s" .Values.image.registry .Values.image.repository .Values.image.tag }}
{{- else }}
{{- printf "%s:%s" .Values.image.repository .Values.image.tag }}
{{- end }}
{{- end }}

{{- define "statuspage.imagePullSecrets" -}}
{{- $secrets := list }}
{{- range .Values.imagePullSecrets }}
{{- $secrets = append $secrets . }}
{{- end }}
{{- if .Values.global }}
{{- if .Values.global.replicated }}
{{- if .Values.global.replicated.dockerconfigjson }}
{{- $secrets = append $secrets (dict "name" (printf "%s-replicated-pull-secret" (include "statuspage.fullname" .))) }}
{{- end }}
{{- end }}
{{- end }}
{{- if $secrets }}
imagePullSecrets:
{{- toYaml $secrets | nindent 2 }}
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
