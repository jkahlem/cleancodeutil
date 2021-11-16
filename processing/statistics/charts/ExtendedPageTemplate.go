package charts

var ExtendedPageTpl = `
{{- define "extendedPage" }}
<!DOCTYPE html>
<html>
    {{- template "header" . }}
<body>
{{ if eq .Layout "none" }}
    {{- range .Charts }}
		{{ if eq .Type "table" }}
			<div class="container-md">{{ template "table" . }}</div>
		{{ else }}
			{{ template "base" . }}
		{{ end }}
	{{- end }}
{{ end }}
{{ if eq .Layout "center" }}
    {{- range .Charts }}
		{{ if eq .Type "table" }}
			<div class="container-md">{{ template "table" . }}</div>
		{{ else }}
			<style> .container {display: flex;justify-content: center;align-items: center;} .item {margin: auto;} </style>
			{{ template "base" . }}
		{{ end }}
	{{- end }}
{{ end }}
{{ if eq .Layout "flex" }}
    <div class="box">
		{{- range .Charts }}
			{{ if eq .Type "table" }}
				{{ template "table" . }}
			{{ else }}
				<style> .box { justify-content:center; display:flex; flex-wrap:wrap } </style>
				{{ template "base" . }}
			{{ end }}
		{{- end }}
	</div>
{{ end }}
</body>
</html>
{{ end }}
`
