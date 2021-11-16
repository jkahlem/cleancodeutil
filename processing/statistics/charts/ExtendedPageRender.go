package charts

import (
	"bytes"
	"io"
	"regexp"

	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/render"
	"github.com/go-echarts/go-echarts/v2/templates"
)

const (
	Jquery        = "https://code.jquery.com/jquery-3.3.1.min.js"
	DatatablesJS  = "https://cdn.datatables.net/v/bs5/jq-3.3.1/dt-1.10.25/datatables.min.js"
	DatatablesCSS = "https://cdn.datatables.net/v/bs5/jq-3.3.1/dt-1.10.25/datatables.min.css"
	BootstrapCSS  = "https://cdn.jsdelivr.net/npm/bootstrap@5.0.2/dist/css/bootstrap.min.css"
)

// Creates a go-echarts page with the extended renderer
func NewPage() *components.Page {
	page := components.NewPage()
	page.Renderer = NewPageRender(page, page.Validate)
	page.AddCustomizedJSAssets(Jquery, DatatablesJS)
	page.AddCustomizedCSSAssets(DatatablesCSS, BootstrapCSS)
	return page
}

// This page render extends the pageRender of go-echarts by supporting table charts
type pageRender struct {
	c      interface{}
	before []func()
}

// NewPageRender returns an extended render implementation for Page.
func NewPageRender(c interface{}, before ...func()) render.Renderer {
	return &pageRender{c: c, before: before}
}

// Render
func (r *pageRender) Render(w io.Writer) error {
	for _, fn := range r.before {
		fn()
	}

	// add extended page template
	contents := []string{templates.HeaderTpl, templates.BaseTpl, TableTpl, ExtendedPageTpl}
	tpl := render.MustTemplate("extendedPage", contents)

	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, "extendedPage", r.c); err != nil {
		return err
	}

	pat := regexp.MustCompile(`(__f__")|("__f__)|(__f__)`)
	content := pat.ReplaceAll(buf.Bytes(), []byte(""))

	_, err := w.Write(content)
	return err
}
