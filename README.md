goldmark-alert
=========================

[![GoDev][godev-image]][godev-url]

[godev-image]: https://pkg.go.dev/badge/github.com/yuin/goldmark-alert
[godev-url]: https://pkg.go.dev/github.com/yuin/goldmark-alert

goldmark-alert is a goldmark extension that adds support for Github-style alert blocks in Markdown.

## Compatiblity
`github.com/yuin/goldmark-alert` is compatible with `goldmark/v2`.

## Installation

```sh
go get github.com/yuin/goldmark-alert
```

## Usage

```go
import (
    "bytes"
    "fmt"

    "github.com/yuin/goldmark-alert"
    "github.com/yuin/goldmark/v2/parser"
    "github.com/yuin/goldmark/v2/renderer/html"
)

func main() {
    p := parser.New(parser.WithExtensions(alert.Parser)
    // p := parser.New(parser.WithExtensions(alert.NewParser(alert.WithAlertKinds("INFO", "TIP"), alert.WithAllowsNested(true))))
    r := html.New(html.WithExtensions(alert.HTMLRenderer))
    // r := html.New(html.WithExtensions(alert.NewHTMLRenderer(alert.WithIcons(map[string]string{...})))

    source := []byte(`# Title

> [!NOTE] Custom title
>
> This is a note.

`)
    node := p.Parse((source)

    var buf bytes.Buffer
    _ := r.Render(&buf, source, node)
}
```

goldmark-alert supports [GitLab alerts custom titles](https://docs.gitlab.com/user/markdown/#alerts)

HTML output for the above example:

```html
<div class="markdown-alert markdown-alert-note">
<p class="markdown-alert-title">
<svg class="markdown-icon markdown-icon-note mr-2" ... ></svg>
Custom title
</p>
<p>This is a note</p>
</div>
```


**Parser options**

| Option | Description | Default |
| --------|-------------| ---------|
| `alert.WithAlertKinds(kinds ... string)` | Enable info alert blocks | `[]string{"NOTE", "TIP", "IMPORTANT", "WARNING", "CAUTION"}` |
| `alert.WithAllowsNested(...bool)` | Github does not allow nested alert blocks, but you can enable it | `false` |

**HTML Renderer options**

| Option | Description | Default |
| --------|-------------| ---------|
| `alert.WithIcons(icons map[string]string)` | Set icons html for each alert kind | default icons |


## License
MIT

## Author

Yusuke Inuzuka
