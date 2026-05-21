package content_test

import (
	"os"
	"path/filepath"
	"testing"

	"lostsgnl.com/internal/content"
)

func writeFile(t *testing.T, dir, name, data string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadSite_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "posts"), 0o755)
	os.MkdirAll(filepath.Join(dir, "pages"), 0o755)

	site, err := content.LoadSite(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(site.Posts) != 0 {
		t.Errorf("got %d posts, want 0", len(site.Posts))
	}
	if len(site.Pages) != 0 {
		t.Errorf("got %d pages, want 0", len(site.Pages))
	}
}

func TestLoadSite_SinglePost(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "posts"), 0o755)
	os.MkdirAll(filepath.Join(dir, "pages"), 0o755)

	writeFile(t, dir, "posts/hello-world.md", `---
title: "Hello World"
date: 2026-04-12
description: "First post"
tags: [go, ssg]
author: "Mark"
---
# Hello

This is my first post.`)

	site, err := content.LoadSite(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(site.Posts) != 1 {
		t.Fatalf("got %d posts, want 1", len(site.Posts))
	}
	p := site.Posts[0]
	if p.Title != "Hello World" {
		t.Errorf("title = %q, want %q", p.Title, "Hello World")
	}
	if p.Slug != "hello-world" {
		t.Errorf("slug = %q, want %q", p.Slug, "hello-world")
	}
	if p.URL != "/posts/hello-world/" {
		t.Errorf("url = %q, want %q", p.URL, "/posts/hello-world/")
	}
	if p.MarkdownURL != "/posts/hello-world/index.md" {
		t.Errorf("markdown url = %q, want %q", p.MarkdownURL, "/posts/hello-world/index.md")
	}
	if len(p.Tags) != 2 {
		t.Errorf("tags = %v, want [go ssg]", p.Tags)
	}
	if len(p.SourceMD) == 0 {
		t.Error("SourceMD is empty")
	}
	if len(p.Content) == 0 {
		t.Error("Content is empty")
	}
}

func TestLoadSite_PostsSortedByDate(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "posts"), 0o755)
	os.MkdirAll(filepath.Join(dir, "pages"), 0o755)

	writeFile(t, dir, "posts/old.md", `---
title: "Old"
date: 2026-01-01
---
Old post.`)

	writeFile(t, dir, "posts/new.md", `---
title: "New"
date: 2026-04-12
---
New post.`)

	site, err := content.LoadSite(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(site.Posts) != 2 {
		t.Fatalf("got %d posts, want 2", len(site.Posts))
	}
	if site.Posts[0].Title != "New" {
		t.Errorf("first post = %q, want %q (sorted desc by date)", site.Posts[0].Title, "New")
	}
}

func TestLoadSite_DraftPostExcluded(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "posts"), 0o755)
	os.MkdirAll(filepath.Join(dir, "pages"), 0o755)

	writeFile(t, dir, "posts/published.md", `---
title: "Published"
date: 2026-01-01
---
Visible.`)

	writeFile(t, dir, "posts/wip.md", `---
title: "WIP"
date: 2026-02-01
draft: true
---
Not visible.`)

	site, err := content.LoadSite(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(site.Posts) != 1 {
		t.Fatalf("got %d posts, want 1 (draft excluded)", len(site.Posts))
	}
	if site.Posts[0].Title != "Published" {
		t.Errorf("post = %q, want %q", site.Posts[0].Title, "Published")
	}
}

func TestLoadSite_DraftPageExcluded(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "posts"), 0o755)
	os.MkdirAll(filepath.Join(dir, "pages"), 0o755)

	writeFile(t, dir, "pages/visible.md", `---
title: "Visible"
nav_order: 1
---
Visible.`)

	writeFile(t, dir, "pages/hidden.md", `---
title: "Hidden"
nav_order: 2
draft: true
---
Not visible.`)

	site, err := content.LoadSite(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(site.Pages) != 1 {
		t.Fatalf("got %d pages, want 1 (draft excluded)", len(site.Pages))
	}
	if site.Pages[0].Title != "Visible" {
		t.Errorf("page = %q, want %q", site.Pages[0].Title, "Visible")
	}
}

func TestLoadSite_TagsCollected(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "posts"), 0o755)
	os.MkdirAll(filepath.Join(dir, "pages"), 0o755)

	writeFile(t, dir, "posts/a.md", `---
title: "A"
date: 2026-01-01
tags: [go, css]
---
A.`)

	writeFile(t, dir, "posts/b.md", `---
title: "B"
date: 2026-02-01
tags: [go, html]
---
B.`)

	site, err := content.LoadSite(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(site.Tags) != 3 {
		t.Fatalf("got %d tags, want 3 (css, go, html)", len(site.Tags))
	}
	// Tags should be sorted alphabetically
	if site.Tags[0].Name != "css" {
		t.Errorf("first tag = %q, want %q", site.Tags[0].Name, "css")
	}
	// "go" tag should have 2 posts
	for _, tag := range site.Tags {
		if tag.Name == "go" && len(tag.Posts) != 2 {
			t.Errorf("go tag has %d posts, want 2", len(tag.Posts))
		}
	}
	// Tags should have colours assigned
	if site.Tags[0].Color == "" {
		t.Error("tag color is empty")
	}
}

func TestLoadSite_Pages(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "posts"), 0o755)
	os.MkdirAll(filepath.Join(dir, "pages"), 0o755)

	writeFile(t, dir, "pages/about.md", `---
title: "About"
nav_order: 1
---
About me.`)

	writeFile(t, dir, "pages/contact.md", `---
title: "Contact"
nav_order: 2
---
Contact me.`)

	site, err := content.LoadSite(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(site.Pages) != 2 {
		t.Fatalf("got %d pages, want 2", len(site.Pages))
	}
	if site.Pages[0].Title != "About" {
		t.Errorf("first page = %q, want %q (sorted by nav_order)", site.Pages[0].Title, "About")
	}
	if site.Pages[0].URL != "/about/" {
		t.Errorf("url = %q, want %q", site.Pages[0].URL, "/about/")
	}
}

// Verifies that the image frontmatter field is loaded into the Post.
func TestLoadSite_PostImageField(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "posts"), 0o755)
	os.MkdirAll(filepath.Join(dir, "pages"), 0o755)

	writeFile(t, dir, "posts/with-image.md", `---
title: "With Image"
date: 2026-04-12
description: "A post with an image"
tags: [test]
image: /assets/img/hero.jpg
---
Content.`)

	site, err := content.LoadSite(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(site.Posts) != 1 {
		t.Fatalf("got %d posts, want 1", len(site.Posts))
	}
	if site.Posts[0].Image != "/assets/img/hero.jpg" {
		t.Errorf("image = %q, want %q", site.Posts[0].Image, "/assets/img/hero.jpg")
	}
}

func TestLoadSite_IgnoresNonMarkdown(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "posts"), 0o755)
	os.MkdirAll(filepath.Join(dir, "pages"), 0o755)

	writeFile(t, dir, "posts/notes.txt", "not markdown")
	writeFile(t, dir, "posts/image.png", "not markdown")

	site, err := content.LoadSite(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(site.Posts) != 0 {
		t.Errorf("got %d posts, want 0 (non-markdown ignored)", len(site.Posts))
	}
}
