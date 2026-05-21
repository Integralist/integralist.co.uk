package renderer_test

import (
	"html/template"
	"strings"
	"testing"
	"time"

	"github.com/integralist/integralist.co.uk/internal/model"
	"github.com/integralist/integralist.co.uk/internal/renderer"
)

const templateDir = "../../assets/templates"

func testSite() *model.Site {
	posts := []*model.Post{
		{
			Title:       "First Post",
			Slug:        "first-post",
			Date:        time.Date(2026, 4, 12, 0, 0, 0, 0, time.UTC),
			Description: "The first post",
			Tags:        []string{"go", "ssg"},
			Author:      "Mark",
			Content:     template.HTML("<p>Hello world.</p>"),
			URL:         "/posts/first-post/",
			MarkdownURL: "/posts/first-post/index.md",
		},
	}

	pages := []*model.Page{
		{
			Title:       "About",
			Slug:        "about",
			Content:     template.HTML("<p>About me.</p>"),
			URL:         "/about/",
			MarkdownURL: "/about/index.md",
			NavOrder:    1,
		},
	}

	tags := []*model.Tag{
		{Name: "go", Slug: "go", Posts: posts, URL: "/tags/go/", Color: "#D4796A"},
		{Name: "ssg", Slug: "ssg", Posts: posts, URL: "/tags/ssg/", Color: "#D4A04A"},
	}

	return &model.Site{BaseURL: "https://www.integralist.co.uk", Posts: posts, Pages: pages, Tags: tags}
}

func TestRenderHome_ContainsPostLinks(t *testing.T) {
	r, err := renderer.New(templateDir)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	site := testSite()
	out, err := r.RenderHome(site)
	if err != nil {
		t.Fatalf("RenderHome error: %v", err)
	}
	html := string(out)
	if !strings.Contains(html, "/posts/first-post/") {
		t.Error("home page missing post link")
	}
	if !strings.Contains(html, "First Post") {
		t.Error("home page missing post title")
	}
	if !strings.Contains(html, "<html") {
		t.Error("home page missing html wrapper from base template")
	}
}

func TestRenderPost_ContainsTitleAndMarkdownLink(t *testing.T) {
	r, err := renderer.New(templateDir)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	site := testSite()
	out, err := r.RenderPost(site.Posts[0], site)
	if err != nil {
		t.Fatalf("RenderPost error: %v", err)
	}
	html := string(out)
	if !strings.Contains(html, "First Post") {
		t.Error("post page missing title")
	}
	if !strings.Contains(html, `type="text/markdown"`) {
		t.Error("post page missing markdown alternate link")
	}
	if !strings.Contains(html, "index.md") {
		t.Error("post page missing markdown URL")
	}
}

func TestRenderPost_ContainsTagPills(t *testing.T) {
	r, err := renderer.New(templateDir)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	site := testSite()
	out, err := r.RenderPost(site.Posts[0], site)
	if err != nil {
		t.Fatalf("RenderPost error: %v", err)
	}
	html := string(out)
	if !strings.Contains(html, `class="tag"`) {
		t.Error("post page missing tag pills")
	}
	if !strings.Contains(html, "#D4796A") {
		t.Error("post page missing tag color")
	}
}

func TestRenderPage(t *testing.T) {
	r, err := renderer.New(templateDir)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	site := testSite()
	out, err := r.RenderPage(site.Pages[0], site)
	if err != nil {
		t.Fatalf("RenderPage error: %v", err)
	}
	html := string(out)
	if !strings.Contains(html, "About") {
		t.Error("page missing title")
	}
	if !strings.Contains(html, "About me.") {
		t.Error("page missing content")
	}
}

// Verifies that tag pages include a markdown alternate link in the HTML head.
func TestRenderTagPage_ContainsMarkdownAlternateLink(t *testing.T) {
	r, err := renderer.New(templateDir)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	site := testSite()
	out, err := r.RenderTagPage(site.Tags[0], site)
	if err != nil {
		t.Fatalf("RenderTagPage error: %v", err)
	}
	html := string(out)
	if !strings.Contains(html, `type="text/markdown"`) {
		t.Error("tag page missing markdown alternate link")
	}
}

// Verifies that the tags index page includes a markdown alternate link.
func TestRenderTagsIndex_ContainsMarkdownAlternateLink(t *testing.T) {
	r, err := renderer.New(templateDir)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	site := testSite()
	out, err := r.RenderTagsIndex(site)
	if err != nil {
		t.Fatalf("RenderTagsIndex error: %v", err)
	}
	html := string(out)
	if !strings.Contains(html, `type="text/markdown"`) {
		t.Error("tags index missing markdown alternate link")
	}
}

// Verifies that post pages include article:published_time and article:tag meta tags.
func TestRenderPost_ContainsArticleMetaTags(t *testing.T) {
	r, err := renderer.New(templateDir)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	site := testSite()
	out, err := r.RenderPost(site.Posts[0], site)
	if err != nil {
		t.Fatalf("RenderPost error: %v", err)
	}
	html := string(out)
	if !strings.Contains(html, `article:published_time`) {
		t.Error("post page missing article:published_time meta tag")
	}
	if !strings.Contains(html, `2026-04-12`) {
		t.Error("post page missing published date in article:published_time")
	}
	if !strings.Contains(html, `article:tag`) {
		t.Error("post page missing article:tag meta tags")
	}
}

// Verifies that post pages include twitter:url meta tag.
func TestRenderPost_ContainsTwitterURL(t *testing.T) {
	r, err := renderer.New(templateDir)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	site := testSite()
	out, err := r.RenderPost(site.Posts[0], site)
	if err != nil {
		t.Fatalf("RenderPost error: %v", err)
	}
	html := string(out)
	if !strings.Contains(html, `twitter:url`) {
		t.Error("post page missing twitter:url meta tag")
	}
	if !strings.Contains(html, `https://www.integralist.co.uk/posts/first-post/`) {
		t.Error("post page missing canonical URL in twitter:url")
	}
}

// Verifies that post pages include author/tags twitter labels.
func TestRenderPost_ContainsTwitterLabels(t *testing.T) {
	r, err := renderer.New(templateDir)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	site := testSite()
	out, err := r.RenderPost(site.Posts[0], site)
	if err != nil {
		t.Fatalf("RenderPost error: %v", err)
	}
	html := string(out)
	if !strings.Contains(html, `twitter:label1`) {
		t.Error("post page missing twitter:label1")
	}
	if !strings.Contains(html, `Written by`) {
		t.Error("post page missing 'Written by' label")
	}
	if !strings.Contains(html, `twitter:data1`) {
		t.Error("post page missing twitter:data1")
	}
	if !strings.Contains(html, `twitter:label2`) {
		t.Error("post page missing twitter:label2")
	}
	if !strings.Contains(html, `Filed under`) {
		t.Error("post page missing 'Filed under' label")
	}
}

// Verifies that post pages with an image get og:image and summary_large_image twitter card.
func TestRenderPost_WithImage(t *testing.T) {
	r, err := renderer.New(templateDir)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	site := testSite()
	site.Posts[0].Image = "/assets/img/hero.jpg"
	out, err := r.RenderPost(site.Posts[0], site)
	if err != nil {
		t.Fatalf("RenderPost error: %v", err)
	}
	html := string(out)
	if !strings.Contains(html, `og:image`) {
		t.Error("post page missing og:image meta tag")
	}
	if !strings.Contains(html, `https://www.integralist.co.uk/assets/img/hero.jpg`) {
		t.Error("post page missing full image URL in og:image")
	}
	if !strings.Contains(html, `twitter:image`) {
		t.Error("post page missing twitter:image meta tag")
	}
	if !strings.Contains(html, `summary_large_image`) {
		t.Error("post page should use summary_large_image when image is present")
	}
}

// Verifies that a post with an image renders a linked image inside post-content.
func TestRenderPost_WithImage_RendersLinkedImage(t *testing.T) {
	r, err := renderer.New(templateDir)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	site := testSite()
	site.Posts[0].Image = "/assets/img/hero.jpg"
	out, err := r.RenderPost(site.Posts[0], site)
	if err != nil {
		t.Fatalf("RenderPost error: %v", err)
	}
	html := string(out)
	if !strings.Contains(html, `class="post-hero"`) {
		t.Error("post hero image missing post-hero class")
	}
	if !strings.Contains(html, `href="/assets/img/hero.jpg"`) {
		t.Error("post hero image missing link wrapper")
	}
	if !strings.Contains(html, `target="_blank"`) {
		t.Error("post hero image link missing target=_blank")
	}
	if !strings.Contains(html, `src="/assets/img/hero.jpg"`) {
		t.Error("post hero image missing src")
	}
}

// Verifies that a post without an image does not render a hero image.
func TestRenderPost_WithoutImage_NoHeroImage(t *testing.T) {
	r, err := renderer.New(templateDir)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	site := testSite()
	site.Posts[0].Image = ""
	out, err := r.RenderPost(site.Posts[0], site)
	if err != nil {
		t.Fatalf("RenderPost error: %v", err)
	}
	html := string(out)
	if strings.Contains(html, `alt="First Post"`) {
		t.Error("post page without image should not render hero image")
	}
}

// Verifies that post pages without an image use summary twitter card.
func TestRenderPost_WithoutImage_UsesSummaryCard(t *testing.T) {
	r, err := renderer.New(templateDir)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	site := testSite()
	site.Posts[0].Image = ""
	out, err := r.RenderPost(site.Posts[0], site)
	if err != nil {
		t.Fatalf("RenderPost error: %v", err)
	}
	html := string(out)
	if strings.Contains(html, `summary_large_image`) {
		t.Error("post page without image should not use summary_large_image")
	}
	if !strings.Contains(html, `content="summary"`) {
		t.Error("post page without image should use summary twitter card")
	}
}

// Verifies that post pages include JSON-LD structured data with Article schema.
func TestRenderPost_ContainsJSONLD(t *testing.T) {
	r, err := renderer.New(templateDir)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	site := testSite()
	site.Posts[0].Image = "/assets/img/hero.jpg"
	out, err := r.RenderPost(site.Posts[0], site)
	if err != nil {
		t.Fatalf("RenderPost error: %v", err)
	}
	html := string(out)
	if !strings.Contains(html, `application/ld+json`) {
		t.Error("post page missing JSON-LD script tag")
	}
	if !strings.Contains(html, `"@type":"Article"`) {
		t.Error("post page JSON-LD missing Article type")
	}
	if !strings.Contains(html, `"headline":"First Post"`) {
		t.Error("post page JSON-LD missing headline")
	}
	if !strings.Contains(html, `"datePublished"`) {
		t.Error("post page JSON-LD missing datePublished")
	}
	if !strings.Contains(html, `"https://www.integralist.co.uk/posts/first-post/"`) {
		t.Error("post page JSON-LD missing url")
	}
	if !strings.Contains(html, `"name":"Mark"`) {
		t.Error("post page JSON-LD missing author name")
	}
	if !strings.Contains(html, `"https://www.integralist.co.uk/assets/img/hero.jpg"`) {
		t.Error("post page JSON-LD missing image")
	}
}

// Verifies that non-article pages do not include Article JSON-LD.
func TestRenderHome_NoArticleJSONLD(t *testing.T) {
	r, err := renderer.New(templateDir)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	site := testSite()
	out, err := r.RenderHome(site)
	if err != nil {
		t.Fatalf("RenderHome error: %v", err)
	}
	html := string(out)
	if strings.Contains(html, `"@type":"Article"`) {
		t.Error("home page should not have Article JSON-LD")
	}
}

// Verifies that all pages include the referrer meta tag.
func TestRenderPost_ContainsReferrerTag(t *testing.T) {
	r, err := renderer.New(templateDir)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	site := testSite()
	out, err := r.RenderPost(site.Posts[0], site)
	if err != nil {
		t.Fatalf("RenderPost error: %v", err)
	}
	html := string(out)
	if !strings.Contains(html, `no-referrer-when-downgrade`) {
		t.Error("page missing referrer meta tag")
	}
}

func TestRenderTagPage_ListsPosts(t *testing.T) {
	r, err := renderer.New(templateDir)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	site := testSite()
	out, err := r.RenderTagPage(site.Tags[0], site)
	if err != nil {
		t.Fatalf("RenderTagPage error: %v", err)
	}
	html := string(out)
	if !strings.Contains(html, "First Post") {
		t.Error("tag page missing post title")
	}
	if !strings.Contains(html, "/posts/first-post/") {
		t.Error("tag page missing post link")
	}
}

func TestRenderTagsIndex_ListsAllTags(t *testing.T) {
	r, err := renderer.New(templateDir)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	site := testSite()
	out, err := r.RenderTagsIndex(site)
	if err != nil {
		t.Fatalf("RenderTagsIndex error: %v", err)
	}
	html := string(out)
	if !strings.Contains(html, "/tags/go/") {
		t.Error("tags index missing go tag link")
	}
	if !strings.Contains(html, "/tags/ssg/") {
		t.Error("tags index missing ssg tag link")
	}
}
