package builder_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lostsgnl.com/internal/builder"
)

func setupTestProject(t *testing.T) (contentDir, assetsDir, outputDir string) {
	t.Helper()
	base := t.TempDir()

	contentDir = filepath.Join(base, "content")
	assetsDir = filepath.Join(base, "assets")
	outputDir = filepath.Join(base, "public")

	// Create content
	os.MkdirAll(filepath.Join(contentDir, "posts"), 0o755)
	os.MkdirAll(filepath.Join(contentDir, "pages"), 0o755)

	os.WriteFile(filepath.Join(contentDir, "posts", "hello-world.md"), []byte(`---
title: "Hello World"
date: 2026-04-12
description: "My first post"
tags: [go, ssg]
author: "Mark"
---
# Hello

This is my first post.
`), 0o644)

	os.WriteFile(filepath.Join(contentDir, "pages", "about.md"), []byte(`---
title: "About"
nav_order: 1
---
# About

This is about me.
`), 0o644)

	// Copy real templates
	realTemplates := "../../assets/templates"
	templateDir := filepath.Join(assetsDir, "templates")
	os.MkdirAll(templateDir, 0o755)
	entries, _ := os.ReadDir(realTemplates)
	for _, e := range entries {
		data, _ := os.ReadFile(filepath.Join(realTemplates, e.Name()))
		os.WriteFile(filepath.Join(templateDir, e.Name()), data, 0o644)
	}

	// Create a CSS file
	cssDir := filepath.Join(assetsDir, "css")
	os.MkdirAll(cssDir, 0o755)
	os.WriteFile(filepath.Join(cssDir, "style.css"), []byte("body { margin: 0; }"), 0o644)

	return contentDir, assetsDir, outputDir
}

func TestBuild_CreatesOutputDir(t *testing.T) {
	contentDir, assetsDir, outputDir := setupTestProject(t)
	b := builder.New(contentDir, assetsDir, outputDir, "https://lostsgnl.com")

	if err := b.Build(); err != nil {
		t.Fatalf("Build error: %v", err)
	}

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("output directory not created")
	}
}

func TestBuild_CopiesAssets(t *testing.T) {
	contentDir, assetsDir, outputDir := setupTestProject(t)
	b := builder.New(contentDir, assetsDir, outputDir, "https://lostsgnl.com")

	if err := b.Build(); err != nil {
		t.Fatalf("Build error: %v", err)
	}

	cssPath := filepath.Join(outputDir, "assets", "css", "style.css")
	if _, err := os.Stat(cssPath); os.IsNotExist(err) {
		t.Error("CSS file not copied to output")
	}

	// Templates should NOT be copied
	templatesPath := filepath.Join(outputDir, "assets", "templates")
	if _, err := os.Stat(templatesPath); !os.IsNotExist(err) {
		t.Error("templates directory should not be copied to output")
	}
}

func TestBuild_GeneratesPostHTML(t *testing.T) {
	contentDir, assetsDir, outputDir := setupTestProject(t)
	b := builder.New(contentDir, assetsDir, outputDir, "https://lostsgnl.com")

	if err := b.Build(); err != nil {
		t.Fatalf("Build error: %v", err)
	}

	htmlPath := filepath.Join(outputDir, "posts", "hello-world", "index.html")
	data, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("post HTML not generated: %v", err)
	}
	html := string(data)
	if !strings.Contains(html, "Hello World") {
		t.Error("post HTML missing title")
	}
}

func TestBuild_GeneratesPostMarkdownSource(t *testing.T) {
	contentDir, assetsDir, outputDir := setupTestProject(t)
	b := builder.New(contentDir, assetsDir, outputDir, "https://lostsgnl.com")

	if err := b.Build(); err != nil {
		t.Fatalf("Build error: %v", err)
	}

	mdPath := filepath.Join(outputDir, "posts", "hello-world", "index.md")
	data, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("post markdown source not generated: %v", err)
	}
	if !strings.Contains(string(data), "# Hello") {
		t.Error("markdown source missing content")
	}
}

func TestBuild_GeneratesPageHTML(t *testing.T) {
	contentDir, assetsDir, outputDir := setupTestProject(t)
	b := builder.New(contentDir, assetsDir, outputDir, "https://lostsgnl.com")

	if err := b.Build(); err != nil {
		t.Fatalf("Build error: %v", err)
	}

	htmlPath := filepath.Join(outputDir, "about", "index.html")
	data, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("page HTML not generated: %v", err)
	}
	if !strings.Contains(string(data), "About") {
		t.Error("page HTML missing title")
	}
}

func TestBuild_GeneratesTagPages(t *testing.T) {
	contentDir, assetsDir, outputDir := setupTestProject(t)
	b := builder.New(contentDir, assetsDir, outputDir, "https://lostsgnl.com")

	if err := b.Build(); err != nil {
		t.Fatalf("Build error: %v", err)
	}

	goTagPath := filepath.Join(outputDir, "tags", "go", "index.html")
	data, err := os.ReadFile(goTagPath)
	if err != nil {
		t.Fatalf("tag page not generated: %v", err)
	}
	if !strings.Contains(string(data), "Hello World") {
		t.Error("tag page missing post")
	}
}

func TestBuild_GeneratesTagsIndex(t *testing.T) {
	contentDir, assetsDir, outputDir := setupTestProject(t)
	b := builder.New(contentDir, assetsDir, outputDir, "https://lostsgnl.com")

	if err := b.Build(); err != nil {
		t.Fatalf("Build error: %v", err)
	}

	tagsPath := filepath.Join(outputDir, "tags", "index.html")
	data, err := os.ReadFile(tagsPath)
	if err != nil {
		t.Fatalf("tags index not generated: %v", err)
	}
	html := string(data)
	if !strings.Contains(html, "/tags/go/") {
		t.Error("tags index missing go tag")
	}
}

func TestBuild_GeneratesHomepage(t *testing.T) {
	contentDir, assetsDir, outputDir := setupTestProject(t)
	b := builder.New(contentDir, assetsDir, outputDir, "https://lostsgnl.com")

	if err := b.Build(); err != nil {
		t.Fatalf("Build error: %v", err)
	}

	indexPath := filepath.Join(outputDir, "index.html")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("homepage not generated: %v", err)
	}
	html := string(data)
	if !strings.Contains(html, "Hello World") {
		t.Error("homepage missing post title")
	}
	if !strings.Contains(html, "/posts/hello-world/") {
		t.Error("homepage missing post link")
	}
}

// Verifies that robots.txt is generated with sitemap directive and AI bot rules.
func TestBuild_GeneratesRobotsTxt(t *testing.T) {
	contentDir, assetsDir, outputDir := setupTestProject(t)
	b := builder.New(contentDir, assetsDir, outputDir, "https://lostsgnl.com")

	if err := b.Build(); err != nil {
		t.Fatalf("Build error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outputDir, "robots.txt"))
	if err != nil {
		t.Fatalf("robots.txt not generated: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "Sitemap:") {
		t.Error("robots.txt missing Sitemap directive")
	}
	if !strings.Contains(content, "sitemap.xml") {
		t.Error("robots.txt missing sitemap.xml reference")
	}
	if !strings.Contains(content, "User-agent: *") {
		t.Error("robots.txt missing wildcard user-agent")
	}
}

// Verifies that sitemap.xml is generated with URLs for posts, pages, tags, and homepage.
func TestBuild_GeneratesSitemapXml(t *testing.T) {
	contentDir, assetsDir, outputDir := setupTestProject(t)
	b := builder.New(contentDir, assetsDir, outputDir, "https://lostsgnl.com")

	if err := b.Build(); err != nil {
		t.Fatalf("Build error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outputDir, "sitemap.xml"))
	if err != nil {
		t.Fatalf("sitemap.xml not generated: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "https://lostsgnl.com/") {
		t.Error("sitemap missing homepage URL")
	}
	if !strings.Contains(content, "https://lostsgnl.com/posts/hello-world/") {
		t.Error("sitemap missing post URL")
	}
	if !strings.Contains(content, "https://lostsgnl.com/about/") {
		t.Error("sitemap missing page URL")
	}
	if !strings.Contains(content, "https://lostsgnl.com/tags/go/") {
		t.Error("sitemap missing tag URL")
	}
	if !strings.Contains(content, "https://lostsgnl.com/tags/") {
		t.Error("sitemap missing tags index URL")
	}
	if !strings.Contains(content, "<urlset") {
		t.Error("sitemap missing XML urlset element")
	}
}

// Verifies that llms.txt is generated with site description and content listing.
func TestBuild_GeneratesLlmsTxt(t *testing.T) {
	contentDir, assetsDir, outputDir := setupTestProject(t)
	b := builder.New(contentDir, assetsDir, outputDir, "https://lostsgnl.com")

	if err := b.Build(); err != nil {
		t.Fatalf("Build error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outputDir, "llms.txt"))
	if err != nil {
		t.Fatalf("llms.txt not generated: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "lostsgnl") {
		t.Error("llms.txt missing site name")
	}
	if !strings.Contains(content, "/posts/hello-world/index.md") {
		t.Error("llms.txt missing post markdown URL")
	}
	if !strings.Contains(content, "/about/index.md") {
		t.Error("llms.txt missing page markdown URL")
	}
}

// Verifies that rss.xml is generated with correct RSS 2.0 structure and post content.
func TestBuild_GeneratesRSSFeed(t *testing.T) {
	contentDir, assetsDir, outputDir := setupTestProject(t)
	b := builder.New(contentDir, assetsDir, outputDir, "https://lostsgnl.com")

	if err := b.Build(); err != nil {
		t.Fatalf("Build error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outputDir, "rss.xml"))
	if err != nil {
		t.Fatalf("rss.xml not generated: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "<rss") {
		t.Error("rss.xml missing rss root element")
	}
	if !strings.Contains(content, `version="2.0"`) {
		t.Error("rss.xml missing version 2.0")
	}
	if !strings.Contains(content, "<channel>") {
		t.Error("rss.xml missing channel element")
	}
	if !strings.Contains(content, "<title>lostsgnl</title>") {
		t.Error("rss.xml missing channel title")
	}
	if !strings.Contains(content, "https://lostsgnl.com/") {
		t.Error("rss.xml missing site link")
	}
	if !strings.Contains(content, "<item>") {
		t.Error("rss.xml missing item element")
	}
	if !strings.Contains(content, "<title>Hello World</title>") {
		t.Error("rss.xml missing post title")
	}
	if !strings.Contains(content, "https://lostsgnl.com/posts/hello-world/") {
		t.Error("rss.xml missing post link")
	}
	if !strings.Contains(content, "<description>") {
		t.Error("rss.xml missing description element")
	}
	if !strings.Contains(content, "This is my first post") {
		t.Error("rss.xml missing full post content in description")
	}
}

// Verifies that the RSS link tag is present in generated HTML.
func TestBuild_RSSLinkInHTML(t *testing.T) {
	contentDir, assetsDir, outputDir := setupTestProject(t)
	b := builder.New(contentDir, assetsDir, outputDir, "https://lostsgnl.com")

	if err := b.Build(); err != nil {
		t.Fatalf("Build error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outputDir, "index.html"))
	if err != nil {
		t.Fatalf("homepage not generated: %v", err)
	}
	html := string(data)
	if !strings.Contains(html, `application/rss+xml`) {
		t.Error("HTML missing RSS link type")
	}
	if !strings.Contains(html, `/rss.xml`) {
		t.Error("HTML missing RSS link href")
	}
}

func TestBuild_MarkdownAlternateLinkInHTML(t *testing.T) {
	contentDir, assetsDir, outputDir := setupTestProject(t)
	b := builder.New(contentDir, assetsDir, outputDir, "https://lostsgnl.com")

	if err := b.Build(); err != nil {
		t.Fatalf("Build error: %v", err)
	}

	htmlPath := filepath.Join(outputDir, "posts", "hello-world", "index.html")
	data, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("post HTML not generated: %v", err)
	}
	html := string(data)
	if !strings.Contains(html, `rel="alternate"`) {
		t.Error("post HTML missing alternate link")
	}
	if !strings.Contains(html, `type="text/markdown"`) {
		t.Error("post HTML missing markdown type")
	}
}
