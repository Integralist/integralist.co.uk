package parser

import (
	"regexp"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	mdparser "github.com/gomarkdown/markdown/parser"
)

var (
	imgTag   = regexp.MustCompile(`<img\s+src="([^"]+)"([^/]*)/>`)
	alertTag = regexp.MustCompile(`(?s)<blockquote>\s*<p>\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\]\s*\n?(.*?)</p>\s*</blockquote>`)
)

var alertIcons = map[string]string{
	"NOTE":      "ℹ️",
	"TIP":       "💡",
	"IMPORTANT": "❗",
	"WARNING":   "⚠️",
	"CAUTION":   "🔥",
}

// MarkdownToHTML converts markdown bytes to HTML bytes.
func MarkdownToHTML(md []byte) []byte {
	extensions := mdparser.CommonExtensions | mdparser.AutoHeadingIDs
	p := mdparser.NewWithExtensions(extensions)

	opts := html.RendererOptions{Flags: html.CommonFlags | html.HrefTargetBlank}
	renderer := html.NewRenderer(opts)

	out := markdown.ToHTML(md, p, renderer)
	out = wrapImagesInLinks(out)
	out = transformAlerts(out)
	return out
}

func wrapImagesInLinks(html []byte) []byte {
	return imgTag.ReplaceAllFunc(html, func(match []byte) []byte {
		groups := imgTag.FindSubmatch(match)
		src := groups[1]
		return []byte(`<a href="` + string(src) + `" target="_blank">` + string(match) + `</a>`)
	})
}

func transformAlerts(src []byte) []byte {
	return alertTag.ReplaceAllFunc(src, func(match []byte) []byte {
		groups := alertTag.FindSubmatch(match)
		alertType := string(groups[1])
		body := strings.TrimSpace(string(groups[2]))
		icon := alertIcons[alertType]
		lower := strings.ToLower(alertType)
		return []byte(`<div class="alert alert-` + lower + `"><p class="alert-title">` + icon + " " + alertType + `</p><p>` + body + `</p></div>`)
	})
}
