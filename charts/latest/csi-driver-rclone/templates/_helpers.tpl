{{/* vim: set filetype=mustache: */}}

{{/* Expand the name of the chart.*/}}
{{- define "rclone.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/* labels for helm resources */}}
{{- define "rclone.labels" -}}
labels:
  app.kubernetes.io/instance: "{{ .Release.Name }}"
  app.kubernetes.io/managed-by: "{{ .Release.Service }}"
  app.kubernetes.io/name: "{{ template "rclone.name" . }}"
  app.kubernetes.io/version: "{{ .Chart.AppVersion }}"
  helm.sh/chart: "{{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}"
  {{- if .Values.customLabels }}
{{ toYaml .Values.customLabels | indent 2 -}}
  {{- end }}
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
