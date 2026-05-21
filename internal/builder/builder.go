package builder

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"lostsgnl.com/internal/content"
	"lostsgnl.com/internal/model"
	"lostsgnl.com/internal/renderer"
)

// Builder orchestrates the static site build.
type Builder struct {
	baseURL    string
	contentDir string
	assetsDir  string
	outputDir  string
}

// New creates a Builder.
func New(contentDir, assetsDir, outputDir, baseURL string) *Builder {
	return &Builder{
		baseURL:    baseURL,
		contentDir: contentDir,
		assetsDir:  assetsDir,
		outputDir:  outputDir,
	}
}

// Build generates the static site.
func (b *Builder) Build() error {
	if err := b.clean(); err != nil {
		return fmt.Errorf("clean: %w", err)
	}

	if err := b.copyAssets(); err != nil {
		return fmt.Errorf("copy assets: %w", err)
	}

	site, err := content.LoadSite(b.contentDir)
	if err != nil {
		return fmt.Errorf("load content: %w", err)
	}
	site.BaseURL = b.baseURL

	templateDir := filepath.Join(b.assetsDir, "templates")
	r, err := renderer.New(templateDir)
	if err != nil {
		return fmt.Errorf("init renderer: %w", err)
	}

	if err := b.renderSite(r, site); err != nil {
		return fmt.Errorf("render: %w", err)
	}

	if err := b.generateDiscoveryFiles(site); err != nil {
		return fmt.Errorf("discovery files: %w", err)
	}

	fmt.Printf("Built %d posts, %d pages, %d tags\n", len(site.Posts), len(site.Pages), len(site.Tags))
	return nil
}

func (b *Builder) clean() error {
	if err := os.RemoveAll(b.outputDir); err != nil {
		return err
	}
	return os.MkdirAll(b.outputDir, 0o755)
}

func (b *Builder) copyAssets() error {
	return filepath.WalkDir(b.assetsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(b.assetsDir, path)
		if err != nil {
			return err
		}

		// Skip templates directory
		if strings.HasPrefix(rel, "templates") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		dest := filepath.Join(b.outputDir, "assets", rel)

		if d.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}

		return copyFile(path, dest)
	})
}

func (b *Builder) renderSite(r *renderer.Renderer, site *model.Site) error {
	// Homepage
	html, err := r.RenderHome(site)
	if err != nil {
		return fmt.Errorf("render home: %w", err)
	}
	if err := writeFile(filepath.Join(b.outputDir, "index.html"), html); err != nil {
		return err
	}

	// Posts
	for _, post := range site.Posts {
		dir := filepath.Join(b.outputDir, "posts", post.Slug)
		html, err := r.RenderPost(post, site)
		if err != nil {
			return fmt.Errorf("render post %s: %w", post.Slug, err)
		}
		if err := writeFile(filepath.Join(dir, "index.html"), html); err != nil {
			return err
		}
		if err := writeFile(filepath.Join(dir, "index.md"), post.SourceMD); err != nil {
			return err
		}
	}

	// Pages
	for _, page := range site.Pages {
		dir := filepath.Join(b.outputDir, page.Slug)
		html, err := r.RenderPage(page, site)
		if err != nil {
			return fmt.Errorf("render page %s: %w", page.Slug, err)
		}
		if err := writeFile(filepath.Join(dir, "index.html"), html); err != nil {
			return err
		}
		if err := writeFile(filepath.Join(dir, "index.md"), page.SourceMD); err != nil {
			return err
		}
	}

	// Tags index
	html, err = r.RenderTagsIndex(site)
	if err != nil {
		return fmt.Errorf("render tags index: %w", err)
	}
	if err := writeFile(filepath.Join(b.outputDir, "tags", "index.html"), html); err != nil {
		return err
	}

	// Individual tag pages
	for _, tag := range site.Tags {
		html, err := r.RenderTagPage(tag, site)
		if err != nil {
			return fmt.Errorf("render tag %s: %w", tag.Slug, err)
		}
		if err := writeFile(filepath.Join(b.outputDir, "tags", tag.Slug, "index.html"), html); err != nil {
			return err
		}
	}

	return nil
}

func (b *Builder) generateDiscoveryFiles(site *model.Site) error {
	if err := b.generateRobotsTxt(site); err != nil {
		return fmt.Errorf("robots.txt: %w", err)
	}
	if err := b.generateSitemap(site); err != nil {
		return fmt.Errorf("sitemap.xml: %w", err)
	}
	if err := b.generateLlmsTxt(site); err != nil {
		return fmt.Errorf("llms.txt: %w", err)
	}
	if err := b.generateRSS(site); err != nil {
		return fmt.Errorf("rss.xml: %w", err)
	}
	return nil
}

func (b *Builder) generateRobotsTxt(site *model.Site) error {
	var buf strings.Builder
	buf.WriteString("User-agent: *\n")
	buf.WriteString("Allow: /\n\n")
	buf.WriteString("Sitemap: " + site.BaseURL + "/sitemap.xml\n")
	return writeFile(filepath.Join(b.outputDir, "robots.txt"), []byte(buf.String()))
}

type sitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

func (b *Builder) generateSitemap(site *model.Site) error {
	var urls []sitemapURL

	urls = append(urls, sitemapURL{Loc: site.BaseURL + "/"})

	for _, post := range site.Posts {
		urls = append(urls, sitemapURL{
			Loc:     site.BaseURL + post.URL,
			LastMod: post.Date.Format("2006-01-02"),
		})
	}

	for _, page := range site.Pages {
		urls = append(urls, sitemapURL{Loc: site.BaseURL + page.URL})
	}

	urls = append(urls, sitemapURL{Loc: site.BaseURL + "/tags/"})
	for _, tag := range site.Tags {
		urls = append(urls, sitemapURL{Loc: site.BaseURL + tag.URL})
	}

	urlset := sitemapURLSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}

	data, err := xml.MarshalIndent(urlset, "", "  ")
	if err != nil {
		return err
	}

	out := []byte(xml.Header)
	out = append(out, data...)
	out = append(out, '\n')
	return writeFile(filepath.Join(b.outputDir, "sitemap.xml"), out)
}

func (b *Builder) generateLlmsTxt(site *model.Site) error {
	var buf strings.Builder
	buf.WriteString("# lostsgnl.com\n\n")
	buf.WriteString("> A personal blog about emotions and the human experience.\n\n")
	buf.WriteString("Every page has a companion Markdown file at the same path with an index.md suffix.\n\n")

	buf.WriteString("## Posts\n\n")
	for _, post := range site.Posts {
		buf.WriteString("- [" + post.Title + "](" + site.BaseURL + post.URL + "index.md)\n")
	}

	if len(site.Pages) > 0 {
		buf.WriteString("\n## Pages\n\n")
		for _, page := range site.Pages {
			buf.WriteString("- [" + page.Title + "](" + site.BaseURL + page.URL + "index.md)\n")
		}
	}

	return writeFile(filepath.Join(b.outputDir, "llms.txt"), []byte(buf.String()))
}

type rssChannel struct {
	Description string    `xml:"description"`
	Items       []rssItem `xml:"item"`
	Link        string    `xml:"link"`
	Title       string    `xml:"title"`
}

type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssItem struct {
	Description string `xml:"description"`
	GUID        string `xml:"guid"`
	Link        string `xml:"link"`
	PubDate     string `xml:"pubDate"`
	Title       string `xml:"title"`
}

func (b *Builder) generateRSS(site *model.Site) error {
	items := make([]rssItem, 0, len(site.Posts))
	for _, post := range site.Posts {
		link := site.BaseURL + post.URL
		items = append(items, rssItem{
			Description: string(post.Content),
			GUID:        link,
			Link:        link,
			PubDate:     post.Date.Format("Mon, 02 Jan 2006 15:04:05 -0700"),
			Title:       post.Title,
		})
	}

	feed := rssFeed{
		Version: "2.0",
		Channel: rssChannel{
			Description: "A personal blog about emotions and the human experience.",
			Items:       items,
			Link:        site.BaseURL + "/",
			Title:       "lostsgnl",
		},
	}

	data, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return err
	}

	out := []byte(xml.Header)
	out = append(out, data...)
	out = append(out, '\n')
	return writeFile(filepath.Join(b.outputDir, "rss.xml"), out)
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
