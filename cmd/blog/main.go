package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	skipDirs         = []string{".git", "assets", "cmd", "docs"}
	initOnce         sync.Once
	contentSubPage   []byte
	needleMainInsert = []byte("{INSERT_MAIN}")
	needleNavInsert  = []byte("{INSERT_NAV}")
	errInit          error
)

func main() {
	pages := []string{}

	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if d.IsDir() && slices.Contains(skipDirs, d.Name()) {
			fmt.Printf("skipping: %+v\n", d.Name())
			return filepath.SkipDir
		}
		if d.Name() == "index.html.md" {
			pages = append(pages, path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	sideNavContent := renderSideNav(pages, "../../")

	for _, page := range pages {
		renderPosts(page, sideNavContent)
	}

	renderHome(pages)
	renderLLMsTxt(pages)
}

// initTemplates initializes the template content and ensures it runs only once.
func initTemplates() {
	initOnce.Do(func() {
		f, err := os.Open("assets/templates/page.tpl")
		if err != nil {
			errInit = fmt.Errorf("failed to open page template: %w", err)
			return
		}
		defer f.Close()

		contentSubPage, err = io.ReadAll(f)
		if err != nil {
			errInit = fmt.Errorf("failed to read page template: %w", err)
		}
	})
}

func renderPosts(page, sideNavContent string) {
	// Initialize templates if not already done
	initTemplates()
	if errInit != nil {
		panic(errInit)
	}

	f, err := os.Open(page)
	if err != nil {
		fmt.Printf("failed to open page '%s': %s\n", page, err)
		return
	}

	rawMD, err := io.ReadAll(f)
	if err != nil {
		fmt.Printf("failed to read file '%s': %s\n", page, err)
		_ = f.Close()
		return
	}
	_ = f.Close()

	md := stripFrontMatter(rawMD)

	h := mdToHTML(md)
	content := bytes.Replace(contentSubPage, needleMainInsert, h, 1)

	segs := strings.Split(page, "/")
	if len(segs) < 3 {
		fmt.Printf("Warning: Skipping path with unexpected format in renderPosts: %s\n", page)
		return
	}
	dir := segs[0] + "/" + segs[1]

	content = bytes.Replace(content, needleNavInsert, []byte(sideNavContent), 1)

	dst := filepath.Join(dir, "index.html")
	err = writeFile(dst, content)
	if err != nil {
		fmt.Printf("failed to write file '%s': %s\n", dst, err)
		return
	}

	fmt.Printf("rendered: %s -> %s\n", page, dst)
}

func renderLLMsTxt(pages []string) {
	var (
		pageLinks []*Page
		postLinks []*Page
	)

	for _, path := range pages {
		p, err := parsePage(path)
		if err != nil {
			continue
		}
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

func renderSideNav(pages []string, root string) string {
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
	pageLinks := make([]*Page, 0)
	postLinks := make([]*Page, 0)

	// --- 1. Populate the separate slices ---
	for _, path := range pages {
		p, err := parsePage(path)
		if err != nil {
			fmt.Printf("Warning: Skipping path: %v\n", err)
			continue
		}

		if p.Year == "index" {
			pageLinks = append(pageLinks, p)
		} else {
			postLinks = append(postLinks, p)
		}
	}

	// --- 2. Sort the pageLinks slice alphabetically by title ---
	sort.Slice(pageLinks, func(i, j int) bool {
		return pageLinks[i].Title < pageLinks[j].Title
	})

	// --- 3. Sort the postLinks slice by date descending ---
	sort.Slice(postLinks, func(i, j int) bool {
		date1, err1 := time.Parse("2006-01-02", postLinks[i].Date)
		date2, err2 := time.Parse("2006-01-02", postLinks[j].Date)
		if err1 != nil {
			return false
		}
		if err2 != nil {
			return true
		}
		return date2.Before(date1) // Descending order
	})

	// --- 4. Build the final navigation HTML string ---
	var finalNav strings.Builder

	// Helper to generate HTML for a list of pages
	generateLinks := func(pages []*Page) string {
		var sb strings.Builder
		for _, p := range pages {
			link := filepath.Join(root, p.Dir, "index.html")
			contentNav := strings.Replace(tplNavSingleLink, "{TITLE}", p.Title, 1)
			contentNav = strings.Replace(contentNav, "{LINK}", link, 1)
			sb.WriteString(contentNav)
		}
		return sb.String()
	}

	// Add "Pages" section (if any pages exist)
	if len(pageLinks) > 0 {
		pagesContainer := strings.ReplaceAll(tplNavGroupLinks, "{YEAR}", "Pages")
		pagesContainer = strings.ReplaceAll(pagesContainer, "{YEAR_LINKS}", generateLinks(pageLinks))
		finalNav.WriteString(pagesContainer)
	}

	// Group sorted posts by year
	postsByYear := make(map[string][]*Page)
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

	// Add post sections by year (already in descending date order)
	for _, year := range yearOrder {
		yearContainer := strings.ReplaceAll(tplNavGroupLinks, "{YEAR}", year)
		yearContainer = strings.ReplaceAll(yearContainer, "{YEAR_LINKS}", generateLinks(postsByYear[year]))
		finalNav.WriteString(yearContainer)
	}

	return finalNav.String()
}

func renderHome(pages []string) { // nolint:revive // function-length
	sideNavContent := renderSideNav(pages, "")

	tplIndex := "assets/templates/index.tpl"

	fi, err := os.Open(tplIndex)
	if err != nil {
		err = fmt.Errorf("failed to open index template: %w", err)
		panic(err)
	}

	contentIndex, err := io.ReadAll(fi)
	if err != nil {
		err = fmt.Errorf("failed to read index template: %w", err)
		panic(err)
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

	type post struct {
		date    string // expects ISO 8601 format, e.g., "2024-12-15"
		content string
	}

	var ( // nolint:prealloc
		bufMain bytes.Buffer
		posts   []post
	)

	for _, path := range pages {
		p, err := parsePage(path)
		if err != nil {
			fmt.Printf("Warning: Skipping path in renderHome: %v\n", err)
			continue
		}

		if p.Year != "index" {
			link := filepath.Join(p.Dir, dst)
			contentMain := strings.Replace(tplMain, "{TITLE}", p.Title, 1)
			contentMain = strings.Replace(contentMain, "{LINK}", link, 1)
			contentMain = strings.Replace(contentMain, "{DATE}", p.Date, 1)

			posts = append(posts, post{date: p.Date, content: contentMain})
		}
	}

	sort.Slice(posts, func(i, j int) bool {
		// Parse dates for comparison
		date1, _ := time.Parse("2006-01-02", posts[i].date)
		date2, _ := time.Parse("2006-01-02", posts[j].date)
		return date2.Before(date1) // Descending order
	})

	for _, post := range posts {
		_, _ = bufMain.WriteString(post.content)
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
	// Create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// Create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank | html.TOC | html.LazyLoadImages
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

type Page struct {
	Title string
	Date  string
	Year  string
	Dir   string
	Path  string
}

func parsePage(path string) (*Page, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer f.Close()

	var (
		title, date string
		inFrontMatter bool
	)

	scanner := bufio.NewScanner(f)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
		line := scanner.Text()
		
		if line == "---" {
			if lineCount == 1 {
				inFrontMatter = true
				continue
			}
			if inFrontMatter {
				break // End of front matter
			}
		}

		if inFrontMatter {
			if strings.HasPrefix(line, "date:") {
				date = strings.TrimSpace(strings.TrimPrefix(line, "date:"))
			}
			if strings.HasPrefix(line, "title:") {
				title = strings.TrimSpace(strings.TrimPrefix(line, "title:"))
			}
		}
	}

	if date == "" {
		date = "index"
	}

	segs := strings.Split(path, "/")
	if len(segs) < 3 {
		return nil, fmt.Errorf("unexpected path format: %s", path)
	}
	dir := segs[0] + "/" + segs[1]

	year := date
	if strings.Contains(date, "-") {
		year = strings.Split(date, "-")[0]
	}

	// Fallback to extracting title from directory if not in Front Matter
	if title == "" {
		caser := cases.Title(language.BritishEnglish)
		title = strings.ReplaceAll(caser.String(segs[1]), "-", " ")
	}

	return &Page{
		Title: title,
		Date:  date,
		Year:  year,
		Dir:   dir,
		Path:  path,
	}, nil
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
