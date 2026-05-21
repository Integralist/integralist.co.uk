package parser_test

import (
	"strings"
	"testing"

	"github.com/integralist/integralist.co.uk/internal/parser"
)

func TestMarkdownToHTML_Paragraph(t *testing.T) {
	got := string(parser.MarkdownToHTML([]byte("Hello world.")))
	if !strings.Contains(got, "<p>Hello world.</p>") {
		t.Errorf("got %q, want paragraph tag", got)
	}
}

func TestMarkdownToHTML_Heading(t *testing.T) {
	got := string(parser.MarkdownToHTML([]byte("# Title")))
	if !strings.Contains(got, "<h1") || !strings.Contains(got, "Title</h1>") {
		t.Errorf("got %q, want h1 tag", got)
	}
}

func TestMarkdownToHTML_CodeBlock(t *testing.T) {
	input := "```go\nfmt.Println(\"hi\")\n```"
	got := string(parser.MarkdownToHTML([]byte(input)))
	if !strings.Contains(got, "<code") {
		t.Errorf("got %q, want code tag", got)
	}
}

// Verifies that external links get href and target="_blank".
func TestMarkdownToHTML_Link(t *testing.T) {
	got := string(parser.MarkdownToHTML([]byte("[click](https://example.com)")))
	if !strings.Contains(got, `href="https://example.com"`) {
		t.Errorf("got %q, want link href", got)
	}
	if !strings.Contains(got, `target="_blank"`) {
		t.Errorf("external link missing target=_blank: %q", got)
	}
}

// Verifies that relative links stay in the same tab.
func TestMarkdownToHTML_RelativeLinkNoTargetBlank(t *testing.T) {
	got := string(parser.MarkdownToHTML([]byte("[other post](../something/)")))
	if strings.Contains(got, `target="_blank"`) {
		t.Errorf("relative link should not have target=_blank: %q", got)
	}
}

func TestMarkdownToHTML_ImageWrappedInLink(t *testing.T) {
	got := string(parser.MarkdownToHTML([]byte("![alt text](/assets/img/photo.jpg)")))
	if !strings.Contains(got, `<a href="/assets/img/photo.jpg"`) {
		t.Errorf("got %q, want img wrapped in link", got)
	}
	if !strings.Contains(got, `target="_blank"`) {
		t.Errorf("got %q, want target=_blank on image link", got)
	}
}

func TestMarkdownToHTML_AlertBlockquotes(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		wantClass string
		wantTitle string
	}{
		{"warning", "> [!WARNING]\n> Be careful here.", "alert-warning", "WARNING"},
		{"note", "> [!NOTE]\n> This is informational.", "alert-note", "NOTE"},
		{"tip", "> [!TIP]\n> A useful tip.", "alert-tip", "TIP"},
		{"important", "> [!IMPORTANT]\n> Don't miss this.", "alert-important", "IMPORTANT"},
		{"caution", "> [!CAUTION]\n> Danger zone.", "alert-caution", "CAUTION"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := string(parser.MarkdownToHTML([]byte(tc.input)))
			if !strings.Contains(got, tc.wantClass) {
				t.Errorf("got %q, want class %q", got, tc.wantClass)
			}
			if !strings.Contains(got, tc.wantTitle) {
				t.Errorf("got %q, want title %q", got, tc.wantTitle)
			}
			if strings.Contains(got, "<blockquote>") {
				t.Errorf("got %q, blockquote should be transformed", got)
			}
		})
	}
}
