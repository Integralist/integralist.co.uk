package content

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/integralist/integralist.co.uk/internal/model"
	"github.com/integralist/integralist.co.uk/internal/parser"
)

var tagColors = []string{"#D4796A", "#D4A04A", "#6BA397", "#5B7FA5"}

// LoadSite reads all content from contentDir and returns a populated Site.
func LoadSite(contentDir string) (*model.Site, error) {
	posts, err := loadPosts(filepath.Join(contentDir, "posts"))
	if err != nil {
		return nil, fmt.Errorf("loading posts: %w", err)
	}

	pages, err := loadPages(filepath.Join(contentDir, "pages"))
	if err != nil {
		return nil, fmt.Errorf("loading pages: %w", err)
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})
	sort.Slice(pages, func(i, j int) bool {
		if pages[i].NavOrder != pages[j].NavOrder {
			return pages[i].NavOrder < pages[j].NavOrder
		}
		return pages[i].Title < pages[j].Title
	})

	tags := collectTags(posts)

	return &model.Site{Posts: posts, Pages: pages, Tags: tags}, nil
}

func loadPosts(dir string) ([]*model.Post, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var posts []*model.Post
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}

		meta, body, err := parser.ParseFrontMatter(data)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", e.Name(), err)
		}

		if getBool(meta, "draft") {
			continue
		}

		slug := strings.TrimSuffix(e.Name(), ".md")
		tags := getStringSlice(meta, "tags")
		keywords := getStringSlice(meta, "keywords")
		if len(keywords) == 0 {
			keywords = tags
		}
		post := &model.Post{
			Author:        getString(meta, "author"),
			Content:       template.HTML(parser.MarkdownToHTML(body)),
			Date:          getTime(meta, "date"),
			Description:   getString(meta, "description"),
			Image:         getString(meta, "image"),
			ImagePosition: getString(meta, "image_position"),
			Keywords:      keywords,
			MarkdownURL:   "/posts/" + slug + "/index.md",
			Slug:          slug,
			SourceMD:      data,
			Tags:          tags,
			Title:         getString(meta, "title"),
			URL:           "/posts/" + slug + "/",
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func loadPages(dir string) ([]*model.Page, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var pages []*model.Page
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}

		meta, body, err := parser.ParseFrontMatter(data)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", e.Name(), err)
		}

		if getBool(meta, "draft") {
			continue
		}

		slug := strings.TrimSuffix(e.Name(), ".md")
		page := &model.Page{
			Content:       template.HTML(parser.MarkdownToHTML(body)),
			Description:   getString(meta, "description"),
			Image:         getString(meta, "image"),
			ImagePosition: getString(meta, "image_position"),
			Keywords:      getStringSlice(meta, "keywords"),
			MarkdownURL:   "/" + slug + "/index.md",
			NavOrder:      getInt(meta, "nav_order"),
			Slug:          slug,
			SourceMD:      data,
			Title:         getString(meta, "title"),
			URL:           "/" + slug + "/",
		}
		pages = append(pages, page)
	}

	return pages, nil
}

func collectTags(posts []*model.Post) []*model.Tag {
	tagMap := make(map[string]*model.Tag)

	for _, p := range posts {
		for _, name := range p.Tags {
			slug := model.Slugify(name)
			tag, ok := tagMap[slug]
			if !ok {
				tag = &model.Tag{
					Name: name,
					Slug: slug,
					URL:  "/tags/" + slug + "/",
				}
				tagMap[slug] = tag
			}
			tag.Posts = append(tag.Posts, p)
		}
	}

	tags := make([]*model.Tag, 0, len(tagMap))
	for _, t := range tagMap {
		tags = append(tags, t)
	}
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Name < tags[j].Name
	})

	for i, t := range tags {
		t.Color = tagColors[i%len(tagColors)]
	}

	return tags
}

func getString(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	return s
}

func getInt(m map[string]any, key string) int {
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	default:
		return 0
	}
}

func getTime(m map[string]any, key string) time.Time {
	v, ok := m[key]
	if !ok {
		return time.Time{}
	}
	switch t := v.(type) {
	case time.Time:
		return t
	case string:
		parsed, err := time.Parse("2006-01-02", t)
		if err != nil {
			return time.Time{}
		}
		return parsed
	default:
		return time.Time{}
	}
}

func getBool(m map[string]any, key string) bool {
	v, ok := m[key]
	if !ok {
		return false
	}
	b, ok := v.(bool)
	return ok && b
}

func getStringSlice(m map[string]any, key string) []string {
	v, ok := m[key]
	if !ok {
		return nil
	}
	items, ok := v.([]any)
	if !ok {
		return nil
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}
