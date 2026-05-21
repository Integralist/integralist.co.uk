package renderer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
	"time"

	"lostsgnl.com/internal/model"
)

// TagWithColor pairs a tag name with its display color and URL for templates.
type TagWithColor struct {
	Name  string
	Color string
	URL   string
}

// Renderer parses and executes HTML templates.
type Renderer struct {
	home    *template.Template
	post    *template.Template
	page    *template.Template
	tag     *template.Template
	tagsIdx *template.Template
}

// New creates a Renderer by parsing templates from templateDir.
func New(templateDir string) (*Renderer, error) {
	shared := []string{
		filepath.Join(templateDir, "base.html"),
		filepath.Join(templateDir, "header.html"),
		filepath.Join(templateDir, "footer.html"),
	}

	parse := func(contentTemplate string) (*template.Template, error) {
		files := append(append([]string{}, shared...), filepath.Join(templateDir, contentTemplate))
		return template.ParseFiles(files...)
	}

	home, err := parse("home.html")
	if err != nil {
		return nil, fmt.Errorf("parsing home template: %w", err)
	}
	post, err := parse("post.html")
	if err != nil {
		return nil, fmt.Errorf("parsing post template: %w", err)
	}
	page, err := parse("page.html")
	if err != nil {
		return nil, fmt.Errorf("parsing page template: %w", err)
	}
	tag, err := parse("tag.html")
	if err != nil {
		return nil, fmt.Errorf("parsing tag template: %w", err)
	}
	tagsIdx, err := parse("tags.html")
	if err != nil {
		return nil, fmt.Errorf("parsing tags index template: %w", err)
	}

	return &Renderer{
		home:    home,
		post:    post,
		page:    page,
		tag:     tag,
		tagsIdx: tagsIdx,
	}, nil
}

type baseData struct {
	ArticleTags   []string
	Author        string
	BaseURL       string
	CanonicalURL  string
	Description   string
	Image         string
	JSONLD        template.HTML
	Keywords      string
	MarkdownURL   string
	NavPages      []*model.Page
	OGType        string
	PublishedTime string
	Title         string
	Year          int
}

func newBaseData(site *model.Site) baseData {
	return baseData{
		BaseURL:  site.BaseURL,
		NavPages: site.Pages,
		OGType:   "website",
		Year:     time.Now().Year(),
	}
}

func (r *Renderer) RenderHome(site *model.Site) ([]byte, error) {
	data := struct {
		baseData
		Posts     []*model.Post
		TagColors map[string]TagWithColor
	}{
		baseData:  newBaseData(site),
		Posts:     site.Posts,
		TagColors: buildTagColorMap(site.Tags),
	}
	data.Title = ""
	data.Description = "lostsgnl.com"
	data.CanonicalURL = site.BaseURL + "/"
	return execute(r.home, data)
}

func (r *Renderer) RenderPost(post *model.Post, site *model.Site) ([]byte, error) {
	tagsWithColors := resolveTagColors(post.Tags, site.Tags)

	data := struct {
		baseData
		Post           *model.Post
		TagsWithColors []TagWithColor
	}{
		baseData:       newBaseData(site),
		Post:           post,
		TagsWithColors: tagsWithColors,
	}
	data.ArticleTags = post.Tags
	data.Author = post.Author
	data.CanonicalURL = site.BaseURL + post.URL
	data.Description = post.Description
	if post.Image != "" {
		data.Image = site.BaseURL + post.Image
	}
	data.JSONLD = buildArticleJSONLD(post, site)
	data.Keywords = strings.Join(post.Keywords, ", ")
	data.MarkdownURL = "index.md"
	data.OGType = "article"
	data.PublishedTime = post.Date.Format(time.RFC3339)
	data.Title = post.Title
	return execute(r.post, data)
}

func (r *Renderer) RenderPage(page *model.Page, site *model.Site) ([]byte, error) {
	data := struct {
		baseData
		Page *model.Page
	}{
		baseData: newBaseData(site),
		Page:     page,
	}
	data.Title = page.Title
	data.Description = page.Description
	data.Keywords = strings.Join(page.Keywords, ", ")
	data.CanonicalURL = site.BaseURL + page.URL
	data.MarkdownURL = "index.md"
	return execute(r.page, data)
}

func (r *Renderer) RenderTagPage(tag *model.Tag, site *model.Site) ([]byte, error) {
	data := struct {
		baseData
		Tag *model.Tag
	}{
		baseData: newBaseData(site),
		Tag:      tag,
	}
	data.Title = "Posts tagged \"" + tag.Name + "\""
	data.CanonicalURL = site.BaseURL + tag.URL
	data.MarkdownURL = "index.md"
	return execute(r.tag, data)
}

func (r *Renderer) RenderTagsIndex(site *model.Site) ([]byte, error) {
	data := struct {
		baseData
		Tags []*model.Tag
	}{
		baseData: newBaseData(site),
		Tags:     site.Tags,
	}
	data.Title = "Tags"
	data.CanonicalURL = site.BaseURL + "/tags/"
	data.MarkdownURL = "index.md"
	return execute(r.tagsIdx, data)
}

func execute(tmpl *template.Template, data any) ([]byte, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func buildArticleJSONLD(post *model.Post, site *model.Site) template.HTML {
	ld := map[string]any{
		"@context":      "https://schema.org",
		"@type":         "Article",
		"headline":      post.Title,
		"datePublished": post.Date.Format(time.RFC3339),
		"description":   post.Description,
		"url":           site.BaseURL + post.URL,
	}
	if post.Author != "" {
		ld["author"] = map[string]string{
			"@type": "Person",
			"name":  post.Author,
		}
	}
	if post.Image != "" {
		ld["image"] = site.BaseURL + post.Image
	}

	data, err := json.Marshal(ld)
	if err != nil {
		return ""
	}
	return template.HTML(`<script type="application/ld+json">` + string(data) + `</script>`)
}

func buildTagColorMap(allTags []*model.Tag) map[string]TagWithColor {
	m := make(map[string]TagWithColor, len(allTags))
	for _, t := range allTags {
		m[t.Name] = TagWithColor{Name: t.Name, Color: t.Color, URL: t.URL}
	}
	return m
}

func resolveTagColors(tagNames []string, allTags []*model.Tag) []TagWithColor {
	tagMap := make(map[string]*model.Tag, len(allTags))
	for _, t := range allTags {
		tagMap[t.Name] = t
	}

	result := make([]TagWithColor, 0, len(tagNames))
	for _, name := range tagNames {
		if t, ok := tagMap[name]; ok {
			result = append(result, TagWithColor{Name: t.Name, Color: t.Color, URL: t.URL})
		}
	}
	return result
}
