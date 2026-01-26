package blog

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

var (
	skipDirs         = []string{`.git`, `assets`, `cmd`, `docs`, `internal`}
	needleMainInsert = []byte("{INSERT_MAIN}")
	needleNavInsert  = []byte("{INSERT_NAV}")
)

type Renderer struct {
	contentSubPage []byte
}

func NewRenderer() (*Renderer, error) {
	r := &Renderer{}

	f, err := os.Open("assets/templates/page.tpl")
	if err != nil {
		return nil, fmt.Errorf("failed to open page template: %w", err)
	}
	defer f.Close()

	r.contentSubPage, err = io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read page template: %w", err)
	}

	return r, nil
}

func (r *Renderer) Generate(root string) error {
	var posts []*Post

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if d.IsDir() && slices.Contains(skipDirs, d.Name()) {
			fmt.Printf("skipping: %+v\n", d.Name())
			return filepath.SkipDir
		}
		if d.Name() == "index.html.md" {
			p, err := NewPost(path)
			if err == nil {
				posts = append(posts, p)
			} else {
				fmt.Printf("Warning: failed to parse post %s: %v\n", path, err)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	sideNavContent := r.renderSideNav(posts, "../../")

	for _, p := range posts {
		r.renderPost(p, sideNavContent)
	}

	r.renderHome(posts)
	r.renderLLMsTxt(posts)

	return nil
}

func (r *Renderer) renderPost(p *Post, sideNavContent string) {
	f, err := os.Open(p.Path)
	if err != nil {
		fmt.Printf("failed to open page '%s': %s\n", p.Path, err)
		return
	}
	defer f.Close()

	rawMD, err := io.ReadAll(f)
	if err != nil {
		fmt.Printf("failed to read file '%s': %s\n", p.Path, err)
		return
	}

	md := stripFrontMatter(rawMD)
	h := mdToHTML(md)
	content := bytes.Replace(r.contentSubPage, needleMainInsert, h, 1)

	content = bytes.Replace(content, needleNavInsert, []byte(sideNavContent), 1)

	dst := filepath.Join(p.Dir, "index.html")
	err = writeFile(dst, content)
	if err != nil {
		fmt.Printf("failed to write file '%s': %s\n", dst, err)
		return
	}

	fmt.Printf("rendered: %s -> %s\n", p.Path, dst)
}

func (r *Renderer) renderSideNav(posts []*Post, root string) string {
	tplNavGroupLinks := `
	<li>
	  <span class="opener">{YEAR}</span>
	  <ul>
		{YEAR_LINKS}
	  </ul>
	</li>
	`
	tplNavSingleLink := `<li><a href="{LINK}">{TITLE}</a></li>`

	// Create separate slices for pages and posts
	pageLinks := make([]*Post, 0)
	postLinks := make([]*Post, 0)

	for _, p := range posts {
		if p.Year == "index" {
			pageLinks = append(pageLinks, p)
		} else {
			postLinks = append(postLinks, p)
		}
	}

	// Sort pageLinks alphabetically by title
	sort.Slice(pageLinks, func(i, j int) bool {
		return pageLinks[i].Title < pageLinks[j].Title
	})

	// Sort postLinks by date descending
	sort.Slice(postLinks, func(i, j int) bool {
		date1, err1 := time.Parse("2006-01-02", postLinks[i].Date)
		date2, err2 := time.Parse("2006-01-02", postLinks[j].Date)
		if err1 != nil {
			return false
		}
		if err2 != nil {
			return true
		}
		return date2.Before(date1)
	})

	var finalNav strings.Builder

	// Helper to generate HTML for a list of pages
	generateLinks := func(pages []*Post) string {
		var sb strings.Builder
		for _, p := range pages {
			link := filepath.Join(root, p.Dir, "index.html")
			contentNav := strings.Replace(tplNavSingleLink, "{TITLE}", p.Title, 1)
			contentNav = strings.Replace(contentNav, "{LINK}", link, 1)
			sb.WriteString(contentNav)
		}
		return sb.String()
	}

	if len(pageLinks) > 0 {
		pagesContainer := strings.ReplaceAll(tplNavGroupLinks, "{YEAR}", "Pages")
		pagesContainer = strings.ReplaceAll(pagesContainer, "{YEAR_LINKS}", generateLinks(pageLinks))
		finalNav.WriteString(pagesContainer)
	}

	postsByYear := make(map[string][]*Post)
	yearOrder := make([]string, 0)
	yearSeen := make(map[string]bool)

	for _, p := range postLinks {
		year := p.Year
		postsByYear[year] = append(postsByYear[year], p)
		if !yearSeen[year] {
			yearOrder = append(yearOrder, year)
			yearSeen[year] = true
		}
	}

	for _, year := range yearOrder {
		yearContainer := strings.ReplaceAll(tplNavGroupLinks, "{YEAR}", year)
		yearContainer = strings.ReplaceAll(yearContainer, "{YEAR_LINKS}", generateLinks(postsByYear[year]))
		finalNav.WriteString(yearContainer)
	}

	return finalNav.String()
}

func (r *Renderer) renderHome(posts []*Post) {
	sideNavContent := r.renderSideNav(posts, "")

	tplIndex := "assets/templates/index.tpl"

	fi, err := os.Open(tplIndex)
	if err != nil {
		panic(fmt.Errorf("failed to open index template: %w", err))
	}
	defer fi.Close()

	contentIndex, err := io.ReadAll(fi)
	if err != nil {
		panic(fmt.Errorf("failed to read index template: %w", err))
	}

	dst := "index.html"
	tplMain := `
	<article>
	<h3>{TITLE}</h3>
	<p class="pubdate">{DATE}</p>
	<ul class="actions">
	<li><a href="{LINK}" class="button">Read</a></li>
	</ul>
	</article>
	`

	var bufMain bytes.Buffer
	
	// Filter and sort for home page
	homePosts := make([]*Post, 0)
	for _, p := range posts {
		if p.Year != "index" {
			homePosts = append(homePosts, p)
		}
	}

	sort.Slice(homePosts, func(i, j int) bool {
		date1, _ := time.Parse("2006-01-02", homePosts[i].Date)
		date2, _ := time.Parse("2006-01-02", homePosts[j].Date)
		return date2.Before(date1)
	})

	for _, p := range homePosts {
		link := filepath.Join(p.Dir, dst)
		contentMain := strings.Replace(tplMain, "{TITLE}", p.Title, 1)
		contentMain = strings.Replace(contentMain, "{LINK}", link, 1)
		contentMain = strings.Replace(contentMain, "{DATE}", p.Date, 1)
		bufMain.WriteString(contentMain)
	}

	contentIndex = bytes.Replace(contentIndex, needleMainInsert, bufMain.Bytes(), 1)
	contentIndex = bytes.Replace(contentIndex, needleNavInsert, []byte(sideNavContent), 1)

	err = writeFile("index.html", contentIndex)
	if err != nil {
		fmt.Printf("failed to write index file: %s\n", err)
		return
	}

	fmt.Printf("rendered: %s -> %s\n", tplIndex, dst)
}

func (r *Renderer) renderLLMsTxt(posts []*Post) {
	var (
		pageLinks []*Post
		postLinks []*Post
	)

	for _, p := range posts {
		if p.Year == "index" {
			pageLinks = append(pageLinks, p)
		} else {
			postLinks = append(postLinks, p)
		}
	}

	sort.Slice(pageLinks, func(i, j int) bool { return pageLinks[i].Title < pageLinks[j].Title })
	sort.Slice(postLinks, func(i, j int) bool {
		date1, _ := time.Parse("2006-01-02", postLinks[i].Date)
		date2, _ := time.Parse("2006-01-02", postLinks[j].Date)
		return date2.Before(date1)
	})

	var sb strings.Builder
	sb.WriteString("# Integralist\n\n")
	sb.WriteString("> Mark McDonnell's technical blog covering software engineering, architecture, and infrastructure.\n\n")

	if len(pageLinks) > 0 {
		sb.WriteString("## Pages\n\n")
		for _, p := range pageLinks {
			link := fmt.Sprintf("/%s/index.html.md", p.Dir)
			sb.WriteString(fmt.Sprintf("- [%s](%s)\n", p.Title, link))
		}
		sb.WriteString("\n")
	}

	if len(postLinks) > 0 {
		sb.WriteString("## Posts\n\n")
		for _, p := range postLinks {
			link := fmt.Sprintf("/%s/index.html.md", p.Dir)
			sb.WriteString(fmt.Sprintf("- [%s](%s): %s\n", p.Title, link, p.Date))
		}
	}

	err := writeFile("llms.txt", []byte(sb.String()))
	if err != nil {
		fmt.Printf("failed to write llms.txt: %s\n", err)
	} else {
		fmt.Println("rendered: llms.txt")
	}
}

func writeFile(filename string, content []byte) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		return fmt.Errorf("failed to write content to file: %w", err)
	}

	return nil
}

func mdToHTML(md []byte) []byte {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank | html.TOC | html.LazyLoadImages
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

func stripFrontMatter(content []byte) []byte {
	if !bytes.HasPrefix(content, []byte("---\n")) {
		return content
	}
	parts := bytes.SplitN(content, []byte("\n---\n"), 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return content
}
