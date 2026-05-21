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

// Verifies that alerts separated by other content don't swallow that content.
func TestMarkdownToHTML_AlertsWithContentBetween(t *testing.T) {
	input := "> [!IMPORTANT]\n> First.\n\n```go\nfmt.Println(\"hi\")\n```\n\nhttps://example.com\n\n> [!NOTE]\n> Second.\n"
	got := string(parser.MarkdownToHTML([]byte(input)))

	if !strings.Contains(got, "<pre>") {
		t.Errorf("code block missing from output:\n%s", got)
	}
	if !strings.Contains(got, "alert-important") {
		t.Errorf("missing alert-important:\n%s", got)
	}
	if !strings.Contains(got, "alert-note") {
		t.Errorf("missing alert-note:\n%s", got)
	}
}

// Verifies that consecutive GFM alerts separated by blank lines render as separate alert divs.
func TestMarkdownToHTML_ConsecutiveAlerts(t *testing.T) {
	input := "> [!IMPORTANT]\n> First.\n\n> [!NOTE]\n> Second.\n\n> [!TIP]\n> Third.\n"
	got := string(parser.MarkdownToHTML([]byte(input)))

	if strings.Count(got, `class="alert `) != 3 {
		t.Errorf("expected 3 alert divs, got HTML:\n%s", got)
	}
	if strings.Contains(got, "<blockquote>") {
		t.Errorf("blockquote should be fully transformed, got:\n%s", got)
	}
	for _, want := range []string{"alert-important", "alert-note", "alert-tip"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %s in:\n%s", want, got)
		}
	}
}

// Verifies consecutive multi-line alerts with inline markup render correctly.
func TestMarkdownToHTML_ConsecutiveAlertsMultiLine(t *testing.T) {
	input := "> [!IMPORTANT]\n> You can `range` over a channel, but the loop will never stop unless the\n> channel is closed.\\\n> So when ranging over a channel, think how the program can proceed.\n\n> [!NOTE]\n> You can create [buffered](https://go.dev/tour/concurrency/3) channels.\\\n> Sends to a buffered channel block only when the buffer is full.\n"
	got := string(parser.MarkdownToHTML([]byte(input)))

	if strings.Count(got, `class="alert `) != 2 {
		t.Errorf("expected 2 alert divs, got HTML:\n%s", got)
	}
	if strings.Contains(got, "<blockquote>") {
		t.Errorf("blockquote should be fully transformed, got:\n%s", got)
	}
	if !strings.Contains(got, "alert-important") {
		t.Errorf("missing alert-important in:\n%s", got)
	}
	if !strings.Contains(got, "alert-note") {
		t.Errorf("missing alert-note in:\n%s", got)
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
