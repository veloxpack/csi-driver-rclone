{{/* vim: set filetype=mustache: */}}

{{/* Expand the name of the chart.*/}}
{{- define "rclone.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/* Create a default fully qualified app name. */}}
{{- define "rclone.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}


{{/* Common label key-value pairs (without the "labels:" key) */}}
{{- define "rclone.commonLabels" -}}
app.kubernetes.io/instance: "{{ .Release.Name }}"
app.kubernetes.io/managed-by: "{{ .Release.Service }}"
app.kubernetes.io/name: "{{ template "rclone.name" . }}"
app.kubernetes.io/version: "{{ .Chart.AppVersion }}"
helm.sh/chart: "{{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}"
{{- if .Values.customLabels }}
{{ toYaml .Values.customLabels }}
{{- end }}
{{- end -}}

{{/* labels for helm resources */}}
{{- define "rclone.labels" -}}
labels:
  {{- include "rclone.commonLabels" . | nindent 2 }}
{{- end -}}

{{/* Create the name of the service account to use */}}
{{- define "rclone.serviceAccountName.controller" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (printf "%s-controller-sa" (include "rclone.name" .)) .Values.serviceAccount.controller }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.controller }}
{{- end -}}
{{- end -}}

{{- define "rclone.serviceAccountName.node" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (printf "%s-node-sa" (include "rclone.name" .)) .Values.serviceAccount.node }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.node }}
{{- end -}}
{{- end -}}

{{/* Create the name of the rbac to use */}}
{{- define "rclone.rbacName" -}}
{{- if .Values.rbac.create -}}
    {{ default (printf "%s-%s" (include "rclone.name" .) .Values.rbac.name) .Values.rbac.name }}
{{- else -}}
    {{ default "default" .Values.rbac.name }}
{{- end -}}
{{- end -}}

{{/* Selector labels */}}
{{- define "rclone.selectorLabels" -}}
app.kubernetes.io/name: {{ include "rclone.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/* Resolve the tag to use for the rclone image */}}
{{- define "rclone.rcloneImageTag" -}}
{{- $tag := default "" .Values.image.rclone.tag -}}
{{- if $tag }}
{{- $tag -}}
{{- else if hasPrefix "v" .Chart.AppVersion }}
{{- .Chart.AppVersion -}}
{{- else -}}
{{- printf "v%s" .Chart.AppVersion -}}
{{- end -}}
{{- end -}}
