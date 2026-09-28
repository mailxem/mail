{{- define "xem.name" -}}
{{- printf "%s-xem" .Release.Name | trunc 45 | trimSuffix "-" -}}
{{- end -}}
{{- define "xem.labels" -}}
app.kubernetes.io/name: xem
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | quote }}
{{- end -}}
{{- define "xem.selector" -}}
app.kubernetes.io/name: xem
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
