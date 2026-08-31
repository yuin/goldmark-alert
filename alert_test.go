package alert_test

import (
	"alert"
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
	"github.com/yuin/goldmark/v2/testutil"
)

func renderHTML(t *testing.T, source string, parserExt parser.Extension, rendererExt html.Extension) string {
	t.Helper()

	p := parser.New(parser.WithExtensions(parserExt))
	r := html.New(html.WithExtensions(rendererExt))
	n := p.ParseStringSource(source)

	var buf bytes.Buffer
	_ = r.RenderStringSource(&buf, source, n)
	return strings.TrimSpace(buf.String())
}

func assertRenderedHTML(t *testing.T, source string, parserExt parser.Extension, rendererExt html.Extension, expected string) {
	t.Helper()

	result := renderHTML(t, source, parserExt, rendererExt)
	expected = strings.TrimSpace(expected)
	if result != expected {
		t.Errorf("unexpected result:\n%s", testutil.DiffPretty([]byte(expected), []byte(result)))
	}
}

func TestAlert(t *testing.T) {
	source := `## Title

> [!TIP]
>
> > [!NOTE]
> > This is a note inside a tip.
>
> hoge
`
	expected := strings.TrimSpace(`
<h2>Title</h2>
<div class="markdown-alert markdown-alert-tip">
<p class="markdown-alert-title">
<svg class="markdown-icon markdown-icon-tip mr-2" viewBox="0 0 16 16" width="16" height="16" aria-hidden="true">
  <path d="M8 1.5c-2.363 0-4 1.69-4 3.75 0 .984.424 1.625.984 2.304l.214.253c.223.264.47.556.673.848.284.411.537.896.621 1.49a.75.75 0 0 1-1.484.211c-.04-.282-.163-.547-.37-.847a8.456 8.456 0 0 0-.542-.68c-.084-.1-.173-.205-.268-.32C3.201 7.75 2.5 6.766 2.5 5.25 2.5 2.31 4.863 0 8 0s5.5 2.31 5.5 5.25c0 1.516-.701 2.5-1.328 3.259-.095.115-.184.22-.268.319-.207.245-.383.453-.541.681-.208.3-.33.565-.37.847a.751.751 0 0 1-1.485-.212c.084-.593.337-1.078.621-1.489.203-.292.45-.584.673-.848.075-.088.147-.173.213-.253.561-.679.985-1.32.985-2.304 0-2.06-1.637-3.75-4-3.75ZM5.75 12h4.5a.75.75 0 0 1 0 1.5h-4.5a.75.75 0 0 1 0-1.5ZM6 15.25a.75.75 0 0 1 .75-.75h2.5a.75.75 0 0 1 0 1.5h-2.5a.75.75 0 0 1-.75-.75Z"></path>
</svg>
Tip
</p>
<div class="markdown-alert markdown-alert-note">
<p class="markdown-alert-title">
<svg
  class="markdown-icon markdown-icon-info mr-2"
  viewBox="0 0 16 16"
  width="16"
  height="16"
  aria-hidden="true"
>
  <path d="M0 8a8 8 0 1 1 16 0A8 8 0 0 1 0 8Zm8-6.5a6.5 6.5 0 1 0 0 13 6.5 6.5 0 0 0 0-13ZM6.5 7.75A.75.75 0 0 1 7.25 7h1a.75.75 0 0 1 .75.75v2.75h.25a.75.75 0 0 1 0 1.5h-2a.75.75 0 0 1 0-1.5h.25v-2h-.25a.75.75 0 0 1-.75-.75ZM8 6a1 1 0 1 1 0-2 1 1 0 0 1 0 2Z"></path>
</svg>
Note
</p>
<p>This is a note inside a tip.</p>
</div><p>hoge</p>
</div>
`)
	assertRenderedHTML(t, source, alert.NewParser(alert.WithAllowsNested(true)), alert.HTMLRenderer, expected)
}

func TestAlert_DefaultTitle(t *testing.T) {
	source := `> [!note]
> body
`
	expected := `
<div class="markdown-alert markdown-alert-note">
<p class="markdown-alert-title">
<svg
  class="markdown-icon markdown-icon-info mr-2"
  viewBox="0 0 16 16"
  width="16"
  height="16"
  aria-hidden="true"
>
  <path d="M0 8a8 8 0 1 1 16 0A8 8 0 0 1 0 8Zm8-6.5a6.5 6.5 0 1 0 0 13 6.5 6.5 0 0 0 0-13ZM6.5 7.75A.75.75 0 0 1 7.25 7h1a.75.75 0 0 1 .75.75v2.75h.25a.75.75 0 0 1 0 1.5h-2a.75.75 0 0 1 0-1.5h.25v-2h-.25a.75.75 0 0 1-.75-.75ZM8 6a1 1 0 1 1 0-2 1 1 0 0 1 0 2Z"></path>
</svg>
Note
</p>
<p>body</p>
</div>
`
	assertRenderedHTML(t, source, alert.Parser, alert.HTMLRenderer, expected)
}

func TestAlert_CustomTitle(t *testing.T) {
	source := `> [!WARNING] Heads up
> body
`
	expected := `
<div class="markdown-alert markdown-alert-warning">
<p class="markdown-alert-title">
<svg class="markdown-icon markdown-icon-warning mr-2" viewBox="0 0 16 16" width="16" height="16" aria-hidden="true">
  <path d="M6.457 1.047c.659-1.234 2.427-1.234 3.086 0l6.082 11.378A1.75 1.75 0 0 1 14.082 15H1.918a1.75 1.75 0 0 1-1.543-2.575Zm1.763.707a.25.25 0 0 0-.44 0L1.698 13.132a.25.25 0 0 0 .22.368h12.164a.25.25 0 0 0 .22-.368Zm.53 3.996v2.5a.75.75 0 0 1-1.5 0v-2.5a.75.75 0 0 1 1.5 0ZM9 11a1 1 0 1 1-2 0 1 1 0 0 1 2 0Z"></path>
</svg>
Heads up
</p>
<p>body</p>
</div>
`
	assertRenderedHTML(t, source, alert.Parser, alert.HTMLRenderer, expected)
}

func TestAlert_NestedDisabledByDefault(t *testing.T) {
	source := `> [!TIP]
>
> > [!NOTE]
> > body
`
	expected := `
<div class="markdown-alert markdown-alert-tip">
<p class="markdown-alert-title">
<svg class="markdown-icon markdown-icon-tip mr-2" viewBox="0 0 16 16" width="16" height="16" aria-hidden="true">
  <path d="M8 1.5c-2.363 0-4 1.69-4 3.75 0 .984.424 1.625.984 2.304l.214.253c.223.264.47.556.673.848.284.411.537.896.621 1.49a.75.75 0 0 1-1.484.211c-.04-.282-.163-.547-.37-.847a8.456 8.456 0 0 0-.542-.68c-.084-.1-.173-.205-.268-.32C3.201 7.75 2.5 6.766 2.5 5.25 2.5 2.31 4.863 0 8 0s5.5 2.31 5.5 5.25c0 1.516-.701 2.5-1.328 3.259-.095.115-.184.22-.268.319-.207.245-.383.453-.541.681-.208.3-.33.565-.37.847a.751.751 0 0 1-1.485-.212c.084-.593.337-1.078.621-1.489.203-.292.45-.584.673-.848.075-.088.147-.173.213-.253.561-.679.985-1.32.985-2.304 0-2.06-1.637-3.75-4-3.75ZM5.75 12h4.5a.75.75 0 0 1 0 1.5h-4.5a.75.75 0 0 1 0-1.5ZM6 15.25a.75.75 0 0 1 .75-.75h2.5a.75.75 0 0 1 0 1.5h-2.5a.75.75 0 0 1-.75-.75Z"></path>
</svg>
Tip
</p>
<blockquote>
<p>[!NOTE]
body</p>
</blockquote>
</div>
`
	assertRenderedHTML(t, source, alert.Parser, alert.HTMLRenderer, expected)
}

func TestAlert_CustomKinds(t *testing.T) {
	source := `> [!todo] Work
> body
`
	expected := `
<div class="markdown-alert markdown-alert-todo">
<p class="markdown-alert-title">
Work
</p>
<p>body</p>
</div>
`
	assertRenderedHTML(t, source, alert.NewParser(alert.WithAlertKinds("TODO")), alert.HTMLRenderer, expected)
}

func TestAlert_CustomIcons(t *testing.T) {
	source := `> [!NOTE]
> body
`
	expected := `
<div class="markdown-alert markdown-alert-note">
<p class="markdown-alert-title">
<span>ICON</span>
Note
</p>
<p>body</p>
</div>
`
	assertRenderedHTML(
		t,
		source,
		alert.Parser,
		alert.NewHTMLRenderer(alert.WithAlertIcons(map[string]string{"NOTE": "<span>ICON</span>"})),
		expected,
	)
}

func TestAlert_InvalidKindsRemainBlockquotes(t *testing.T) {
	source := `> [!todo]
> body
`
	expected := `
<blockquote>
<p>[!todo]
body</p>
</blockquote>
`
	assertRenderedHTML(t, source, alert.Parser, alert.HTMLRenderer, expected)
}
