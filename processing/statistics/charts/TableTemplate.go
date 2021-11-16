package charts

// A template for tables using datatables from https://datatables.net/
var TableTpl = `
{{- define "table" }}
<div style="margin-bottom: 3em">
	<h3 style="margin-bottom: 1em">{{ .Title }}</h3>
	<table id="{{ .ChartID }}" class="table table-striped">
		<thead>
			<tr>
				{{- range .Headings }}
				<th scope="col">{{.}}</th>
				{{- end }}
			</tr>
		</thead>
		<tbody>
			{{- range .Records}}
			<tr>
				{{- range . }}
				<td>{{.}}</td>
				{{- end }}
			</tr>
			{{- end }}
		</tbody>
	</table>
	<script type="text/javascript">
		"use strict";
		$(document).ready(function() {
			$("#{{ .ChartID }}").DataTable();
		});
	</script>
</div>
{{ end }}
`
