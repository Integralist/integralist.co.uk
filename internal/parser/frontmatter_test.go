package parser_test

import (
	"testing"

	"lostsgnl.com/internal/parser"
)

func TestParseFrontMatter_ValidYAML(t *testing.T) {
	input := []byte(`---
title: "Hello World"
date: 2026-04-12
tags: [go, ssg]
---
Body content here.`)

	meta, body, err := parser.ParseFrontMatter(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta["title"] != "Hello World" {
		t.Errorf("title = %q, want %q", meta["title"], "Hello World")
	}
	tags, ok := meta["tags"].([]any)
	if !ok || len(tags) != 2 {
		t.Errorf("tags = %v, want [go ssg]", meta["tags"])
	}
	if string(body) != "Body content here." {
		t.Errorf("body = %q, want %q", string(body), "Body content here.")
	}
}

func TestParseFrontMatter_NoFrontMatter(t *testing.T) {
	input := []byte("Just plain markdown content.")

	meta, body, err := parser.ParseFrontMatter(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(meta) != 0 {
		t.Errorf("meta = %v, want empty", meta)
	}
	if string(body) != "Just plain markdown content." {
		t.Errorf("body = %q, want %q", string(body), "Just plain markdown content.")
	}
}

func TestParseFrontMatter_InvalidYAML(t *testing.T) {
	input := []byte(`---
title: [invalid
---
Body.`)

	_, _, err := parser.ParseFrontMatter(input)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestParseFrontMatter_EmptyFile(t *testing.T) {
	meta, body, err := parser.ParseFrontMatter([]byte{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(meta) != 0 {
		t.Errorf("meta = %v, want empty", meta)
	}
	if len(body) != 0 {
		t.Errorf("body = %q, want empty", string(body))
	}
}
