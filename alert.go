// Package alert provides a goldmark extension for parsing and rendering alert blocks in Markdown.
package alert

import (
	"bytes"
	"io"
	"slices"
	"strings"
	"unicode"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer"
	"github.com/yuin/goldmark/v2/renderer/html"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

// KindAlert is a custom AST node kind representing an alert in Markdown.
var KindAlert = ast.NewNodeKind("Alert")

var _ ast.Node = (*Alert)(nil)

// Option is a function for configuring an Alert node.
type Option func(*Alert)

// WithTitle sets the title of the alert.
func WithTitle(title text.SingleLineValue) Option {
	return func(a *Alert) {
		a.Title = title
	}
}

// Alert is a custom AST node representing an alert in Markdown.
type Alert struct {
	ast.BaseBlock

	// AlertKind is the type of alert (e.g., "info", "warning", "error").
	//
	// Note that this value is not 'normalized'; it is the raw value as provided in the Markdown source.
	AlertKind text.SingleLineValue

	// Title is the title of the alert, if provided.
	//
	// If no title is provided, capitization of the Kind is used as the title.
	Title text.SingleLineValue
}

// NewAlert creates a new Alert node with the specified kind and options.
func NewAlert(kind text.SingleLineValue, opts ...Option) *Alert {
	n := &Alert{
		AlertKind: kind,
	}
	for _, opt := range opts {
		opt(n)
	}
	n.Init(n)
	return n
}

// Dump implements [ast.Node].
func (n *Alert) Dump(_ []byte) *ast.NodeDump {
	return ast.NewNodeDump(n, map[string]any{
		"AlertKind": n.AlertKind,
		"Title":     n.Title,
	})
}

// Kind implements [ast.Node].
func (n *Alert) Kind() ast.NodeKind {
	return KindAlert
}

// DisplayTitle returns the title of the alert for display purposes.
func (n *Alert) DisplayTitle(source []byte) string {
	t := n.Title.Value(source)
	if len(t) == 0 {
		t = capitalize(n.AlertKind.Value(source))
	}
	return t
}

type parserConfig struct {
	AllowsNested bool

	Kinds []string
}

// ParserOption is a function for configuring the parser.
type ParserOption func(*parserConfig)

// WithAllowsNested enables or disables the ability to nest alerts within other alerts.
//
// This defaults to false.
func WithAllowsNested(allowsNested ...bool) ParserOption {
	return func(c *parserConfig) {
		if len(allowsNested) > 0 {
			c.AllowsNested = allowsNested[0]
		} else {
			c.AllowsNested = true
		}
	}
}

// GithubAlertKinds is a list of the allowed kinds of alerts in GitHub Flavored Markdown (GFM).
var GithubAlertKinds = []string{"NOTE", "TIP", "IMPORTANT", "WARNING", "CAUTION"}

// WithAlertKinds sets the allowed kinds of alerts.
//
// This defaults to [GithubAlertKinds].
func WithAlertKinds(kinds ...string) ParserOption {
	return func(c *parserConfig) {
		c.Kinds = kinds
	}
}

type astTransformer struct {
	cfg parserConfig
}

func (a *astTransformer) Transform(node *ast.Document, reader text.Reader, _ parser.Context) {
	var blockquotes []*ast.Blockquote
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if bq, ok := n.(*ast.Blockquote); ok {
			if a.cfg.AllowsNested || bq.Parent() == node {
				blockquotes = append(blockquotes, bq)
			}
		}
		return ast.WalkContinue, nil
	})

	// TODO: Support index-based processing if possible

	for i := range len(blockquotes) {
		bq := blockquotes[i]
		fc := bq.FirstChild()
		if fc == nil || fc.Kind() != ast.KindParagraph {
			continue
		}
		p := fc.(*ast.Paragraph)
		if p.FirstChild() == nil || p.FirstChild().Kind() != ast.KindText {
			continue
		}

		var firstLine []byte
		var remove []ast.Node
		for c := p.FirstChild(); c != nil; c = c.NextSibling() {
			if c.Kind() != ast.KindText {
				continue
			}
			t := c.(*ast.Text)
			firstLine = append(firstLine, t.Value.Bytes(reader.Source())...)
			remove = append(remove, c)
			if t.SoftLineBreak() {
				break
			}
		}

		line := util.TrimLeftSpace(firstLine)
		l := len(line)
		i := 0
		if !bytes.HasPrefix(line[i:], []byte("[!")) {
			continue
		}
		i += 2
		start := i
		if n, ok := util.ReadWhile(line, [2]int{i, l}, func(b byte) bool {
			return b != ']'
		}); ok {
			i = n
		}
		if i < l && line[i] != ']' {
			continue
		}
		stop := i
		kind := string(line[start:stop])
		valid := false
		for _, k := range a.cfg.Kinds {
			if strings.EqualFold(kind, k) {
				valid = true
				break
			}
		}
		if !valid {
			continue
		}

		var title string
		if stop != l {
			title = string(util.TrimLeftSpace(line[stop+1:]))
		}

		var opts []Option
		if len(title) > 0 {
			opts = append(opts, WithTitle(
				text.NewSingleLineValueFromString(title, reader.Decoder())))
		}
		alert := NewAlert(text.NewSingleLineValueFromString(kind, reader.Decoder()), opts...)
		for _, c := range remove {
			p.RemoveChild(c)
		}
		if p.FirstChild() == nil {
			bq.RemoveChild(p)
		}
		children := slices.Collect(bq.Children())
		for _, c := range children {
			bq.RemoveChild(c)
			alert.AppendChild(c)
		}
		alert.SetPos(bq.Pos())
		alert.SetBlankPreviousLines(bq.HasBlankPreviousLines())
		bq.Parent().ReplaceChild(bq, alert)
	}
}

type parserExtension struct {
	cfg parserConfig
}

// NewParser creates a new parser extension for handling alert blocks in Markdown.
func NewParser(opts ...ParserOption) parser.Extension {
	cfg := parserConfig{
		AllowsNested: false,
		Kinds:        GithubAlertKinds,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &parserExtension{cfg: cfg}
}

// Parser is a default parser extension for handling alert blocks in Markdown.
var Parser = NewParser()

func (e *parserExtension) ParserOptions(_ *parser.Config) []parser.Option {
	var t parser.ASTTransformer = &astTransformer{e.cfg}

	return []parser.Option{
		parser.WithASTTransformers(
			util.Prioritized(t, 999),
		),
	}
}

// GithubAlertIcons is a map of the default icons to be used for each alert kind in GitHub Flavored Markdown (GFM).
var GithubAlertIcons = map[string]string{
	"NOTE": `<svg
  class="markdown-icon markdown-icon-info mr-2"
  viewBox="0 0 16 16"
  width="16"
  height="16"
  aria-hidden="true"
>
  <path d="M0 8a8 8 0 1 1 16 0A8 8 0 0 1 0 8Zm8-6.5a6.5 6.5 0 1 0 0 13 6.5 6.5 0 0 0 0-13ZM6.5 7.75A.75.75 0 0 1 7.25 7h1a.75.75 0 0 1 .75.75v2.75h.25a.75.75 0 0 1 0 1.5h-2a.75.75 0 0 1 0-1.5h.25v-2h-.25a.75.75 0 0 1-.75-.75ZM8 6a1 1 0 1 1 0-2 1 1 0 0 1 0 2Z"></path>
</svg>`,
	"TIP": `<svg class="markdown-icon markdown-icon-tip mr-2" viewBox="0 0 16 16" width="16" height="16" aria-hidden="true">
  <path d="M8 1.5c-2.363 0-4 1.69-4 3.75 0 .984.424 1.625.984 2.304l.214.253c.223.264.47.556.673.848.284.411.537.896.621 1.49a.75.75 0 0 1-1.484.211c-.04-.282-.163-.547-.37-.847a8.456 8.456 0 0 0-.542-.68c-.084-.1-.173-.205-.268-.32C3.201 7.75 2.5 6.766 2.5 5.25 2.5 2.31 4.863 0 8 0s5.5 2.31 5.5 5.25c0 1.516-.701 2.5-1.328 3.259-.095.115-.184.22-.268.319-.207.245-.383.453-.541.681-.208.3-.33.565-.37.847a.751.751 0 0 1-1.485-.212c.084-.593.337-1.078.621-1.489.203-.292.45-.584.673-.848.075-.088.147-.173.213-.253.561-.679.985-1.32.985-2.304 0-2.06-1.637-3.75-4-3.75ZM5.75 12h4.5a.75.75 0 0 1 0 1.5h-4.5a.75.75 0 0 1 0-1.5ZM6 15.25a.75.75 0 0 1 .75-.75h2.5a.75.75 0 0 1 0 1.5h-2.5a.75.75 0 0 1-.75-.75Z"></path>
</svg>`,
	"IMPORTANT": `<svg class="markdown-icon markdown-icon-important mr-2" viewBox="0 0 16 16" width="16" height="16" aria-hidden="true">
  <path d="M0 1.75C0 .784.784 0 1.75 0h12.5C15.216 0 16 .784 16 1.75v9.5A1.75 1.75 0 0 1 14.25 13H8.06l-2.573 2.573A1.458 1.458 0 0 1 3 14.543V13H1.75A1.75 1.75 0 0 1 0 11.25Zm1.75-.25a.25.25 0 0 0-.25.25v9.5c0 .138.112.25.25.25h2a.75.75 0 0 1 .75.75v2.19l2.72-2.72a.749.749 0 0 1 .53-.22h6.5a.25.25 0 0 0 .25-.25v-9.5Zm7 2.25v2.5a.75.75 0 0 1-1.5 0v-2.5a.75.75 0 0 1 1.5 0ZM9 9a1 1 0 1 1-2 0 1 1 0 0 1 2 0Z"></path>
</svg>`,
	"WARNING": `<svg class="markdown-icon markdown-icon-warning mr-2" viewBox="0 0 16 16" width="16" height="16" aria-hidden="true">
  <path d="M6.457 1.047c.659-1.234 2.427-1.234 3.086 0l6.082 11.378A1.75 1.75 0 0 1 14.082 15H1.918a1.75 1.75 0 0 1-1.543-2.575Zm1.763.707a.25.25 0 0 0-.44 0L1.698 13.132a.25.25 0 0 0 .22.368h12.164a.25.25 0 0 0 .22-.368Zm.53 3.996v2.5a.75.75 0 0 1-1.5 0v-2.5a.75.75 0 0 1 1.5 0ZM9 11a1 1 0 1 1-2 0 1 1 0 0 1 2 0Z"></path>
</svg>`,
	"CAUTION": `<svg class="markdown-icon markdown-icon-caution mr-2" viewBox="0 0 16 16" width="16" height="16" aria-hidden="true">
  <path d="M4.47.22A.749.749 0 0 1 5 0h6c.199 0 .389.079.53.22l4.25 4.25c.141.14.22.331.22.53v6a.749.749 0 0 1-.22.53l-4.25 4.25A.749.749 0 0 1 11 16H5a.749.749 0 0 1-.53-.22L.22 11.53A.749.749 0 0 1 0 11V5c0-.199.079-.389.22-.53Zm.84 1.28L1.5 5.31v5.38l3.81 3.81h5.38l3.81-3.81V5.31L10.69 1.5ZM8 4a.75.75 0 0 1 .75.75v3.5a.75.75 0 0 1-1.5 0v-3.5a.75.75 0 0 1 .75-.75Zm0 8a1 1 0 1 1 0-2 1 1 0 0 1 0 2Z"></path>
</svg>`,
}

type htmlRendererConfig struct {
	Icons map[string]string
}

// HTMLRendererOption is a function for configuring the HTML renderer.
type HTMLRendererOption func(*htmlRendererConfig)

// WithAlertIcons sets the icons to be used for each alert kind in the HTML renderer.
//
// This defaults to [GithubAlertIcons].
// Values are rendered as raw HTML, so they should be valid HTML fragments (e.g., SVG elements, image tags, etc.).
func WithAlertIcons(icons map[string]string) HTMLRendererOption {
	return func(c *htmlRendererConfig) {
		c.Icons = icons
	}
}

type htmlRendererExtension struct {
	cfg htmlRendererConfig
}

func (r *htmlRendererExtension) RendererOptions(_ *html.Config) []html.Option {
	return []html.Option{
		html.WithNodeRenderer(KindAlert, html.NodeRendererFunc(r.renderAlert)),
	}
}

func (r *htmlRendererExtension) renderAlert(
	writer io.Writer, source []byte, n ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	if entering {
		k := n.(*Alert).AlertKind.Str(source)
		tw := html.ContextTextWriter(rc)
		hw := html.ContextHTMLWriter(rc)

		_, _ = w.WriteString(`<div class="markdown-alert markdown-alert-` + strings.ToLower(k) + "\">\n")
		_, _ = w.WriteString("<p class=\"markdown-alert-title\">\n")
		var icon string
		for kind, v := range r.cfg.Icons {
			if strings.EqualFold(k, kind) {
				icon = v
				break
			}
		}
		if len(icon) > 0 {
			_, _ = hw.WriteString(icon)
			_ = w.WriteByte('\n')
		}

		_, _ = tw.WriteString(n.(*Alert).DisplayTitle(source))
		_ = w.WriteByte('\n')
		_, _ = w.WriteString("</p>\n")
	} else {
		_, _ = w.WriteString("</div>")
	}
	return ast.WalkContinue, nil
}

// NewHTMLRenderer creates a new HTML renderer extension for rendering alert blocks in Markdown.
func NewHTMLRenderer(opts ...HTMLRendererOption) html.Extension {
	cfg := htmlRendererConfig{
		Icons: GithubAlertIcons,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &htmlRendererExtension{cfg: cfg}
}

// HTMLRenderer is a default HTML renderer extension for rendering alert blocks in Markdown.
var HTMLRenderer = NewHTMLRenderer()

func capitalize(s string) string {
	if s == "" {
		return s
	}
	s = strings.ToLower(s)
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
