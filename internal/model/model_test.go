package model_test

import (
	"testing"

	"github.com/integralist/integralist.co.uk/internal/model"
)

func TestPost_ReadingTime(t *testing.T) {
	testCases := []struct {
		name  string
		words int
		want  int
	}{
		{"short post rounds up to 1 min", 50, 1},
		{"exactly 200 words", 200, 1},
		{"201 words rounds up to 2 min", 201, 2},
		{"long post", 1000, 5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body := make([]byte, 0, tc.words*6)
			for i := range tc.words {
				if i > 0 {
					body = append(body, ' ')
				}
				body = append(body, []byte("word")...)
			}
			p := model.Post{SourceMD: body}
			if got := p.ReadingTime(); got != tc.want {
				t.Errorf("ReadingTime() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello World", "hello-world"},
		{"Go 1.26", "go-1-26"},
		{"  spaces  ", "spaces"},
		{"UPPERCASE", "uppercase"},
		{"special!@#chars", "special-chars"},
		{"multiple---dashes", "multiple-dashes"},
		{"trailing-", "trailing"},
		{"café", "caf"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := model.Slugify(tt.input)
			if got != tt.want {
				t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
