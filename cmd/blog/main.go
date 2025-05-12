package main

import (
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
	skipDirs         = []string{".git", "assets", "cmd"}
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
		if filepath.Ext(d.Name()) == ".md" && d.Name() != "README.md" {
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

	md, err := io.ReadAll(f)
	if err != nil {
		fmt.Printf("failed to read file '%s': %s\n", page, err)
		_ = f.Close()
		return
	}
	_ = f.Close()

	h := mdToHTML(md)
	content := bytes.Replace(contentSubPage, needleMainInsert, h, 1)

	segs := strings.Split(page, "/")
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

// Helper function to extract link text from the content string
// Example input: "\n\t<li><a href=\"...\">Link Text</a></li>\n\t"
// Example output: "Link Text"
func extractLinkText(content string) string {
	// Find the position right after the opening '>' of the <a> tag
	startIdx := strings.Index(content, ">")
	if startIdx == -1 {
		return content // Return original content as fallback if '>' not found
	}
	startIdx++ // Move past the '>'

	// Find the position of the closing '</a>' tag
	endIdx := strings.Index(content, "</a>")
	if endIdx == -1 || endIdx <= startIdx {
		return content // Return original content as fallback if '</a>' not found or is before '>'
	}

	// Extract the substring and trim leading/trailing whitespace
	return strings.TrimSpace(content[startIdx:endIdx])
}

func renderSideNav(pages []string, root string) string {
	caser := cases.Title(language.BritishEnglish)

	tplNavGroupLinks := `
	<li>
	  <span class="opener">{YEAR}</span>
	  <ul>
		{YEAR_LINKS}
	  </ul>
	</li>
	`
	tplNavSingleLink := `<li><a href="{LINK}">{TITLE}</a></li>`

	type post struct {
		date    string // expects ISO 8601 format, e.g., "2024-12-15" or "index"
		year    string // e.g., "2024" or "index"
		content string // HTML <li> content
	}

	// Create separate slices for pages and posts
	pageLinks := make([]post, 0)
	postLinks := make([]post, 0)

	// --- 1. Populate the separate slices ---
	for _, path := range pages {
		segs := strings.Split(path, "/")
		if len(segs) < 3 {
			fmt.Printf("Warning: Skipping path with unexpected format: %s\n", path)
			continue // Skip malformed paths
		}
		dir := segs[0] + "/" + segs[1]
		date := strings.Split(segs[2], ".")[0] // e.g., "index" or "2025-05-12"
		year := date                           // Default year to date
		if strings.Contains(date, "-") {
			year = strings.Split(date, "-")[0] // e.g., "2025" if date has '-'
		}
		// Handle cases like "pages/resume/index.md" where year should be "index"
		if date == "index" {
			year = "index"
		}

		title := strings.ReplaceAll(caser.String(segs[1]), "-", " ")
		link := filepath.Join(root, dir, "index.html")
		contentNav := strings.Replace(tplNavSingleLink, "{TITLE}", title, 1)
		contentNav = strings.Replace(contentNav, "{LINK}", link, 1)

		p := post{date: date, year: year, content: contentNav}

		if p.year == "index" {
			pageLinks = append(pageLinks, p)
		} else {
			postLinks = append(postLinks, p)
		}
	}

	// --- 2. Sort the pageLinks slice alphabetically by title ---
	sort.Slice(pageLinks, func(i, j int) bool {
		linkTextI := extractLinkText(pageLinks[i].content)
		linkTextJ := extractLinkText(pageLinks[j].content)
		return linkTextI < linkTextJ
	})

	// --- 3. Sort the postLinks slice by date descending ---
	sort.Slice(postLinks, func(i, j int) bool {
		date1, err1 := time.Parse("2006-01-02", postLinks[i].date)
		date2, err2 := time.Parse("2006-01-02", postLinks[j].date)
		// Basic error handling: treat unparsable dates as "equal" or log them
		if err1 != nil {
			fmt.Printf("Warning: Could not parse date '%s' for sorting post '%s'\n", postLinks[i].date, extractLinkText(postLinks[i].content))
			return false // Keep order stable relative to date2 if date1 fails
		}
		if err2 != nil {
			fmt.Printf("Warning: Could not parse date '%s' for sorting post '%s'\n", postLinks[j].date, extractLinkText(postLinks[j].content))
			return true // Keep order stable relative to date1 if date2 fails
		}
		return date2.Before(date1) // Descending order
	})

	// --- 4. Build the final navigation HTML string ---
	var finalNav strings.Builder // Use strings.Builder for efficiency

	// Add "Pages" section (if any pages exist)
	if len(pageLinks) > 0 {
		var pageContentLinks strings.Builder
		for _, pLink := range pageLinks {
			pageContentLinks.WriteString(pLink.content) // Append the pre-formatted <li>...</li>
		}
		// Use strings.ReplaceAll for potentially multiple replacements if templates change
		pagesContainer := strings.ReplaceAll(tplNavGroupLinks, "{YEAR}", "Pages")
		pagesContainer = strings.ReplaceAll(pagesContainer, "{YEAR_LINKS}", pageContentLinks.String())
		finalNav.WriteString(pagesContainer)
	}

	// Group sorted posts by year
	postsByYear := make(map[string][]string)
	yearOrder := make([]string, 0) // Keep track of year order as encountered in sorted posts
	yearSeen := make(map[string]bool)

	for _, pLink := range postLinks { // Iterate sorted postLinks
		year := pLink.year
		postsByYear[year] = append(postsByYear[year], pLink.content)
		if !yearSeen[year] {
			yearOrder = append(yearOrder, year)
			yearSeen[year] = true
		}
	}

	// Add post sections by year (already in descending date order from postLinks sort)
	for _, year := range yearOrder {
		var yearContentLinks strings.Builder
		// Links within postsByYear[year] are already sorted by date
		for _, content := range postsByYear[year] {
			yearContentLinks.WriteString(content)
		}
		yearContainer := strings.ReplaceAll(tplNavGroupLinks, "{YEAR}", year)
		yearContainer = strings.ReplaceAll(yearContainer, "{YEAR_LINKS}", yearContentLinks.String())
		finalNav.WriteString(yearContainer)
	}

	return finalNav.String()
}

func renderHome(pages []string) { // nolint:revive // function-length
	sideNavContent := renderSideNav(pages, "")
	caser := cases.Title(language.BritishEnglish)

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
		year    string
		content string
	}

	var ( // nolint:prealloc
		bufMain bytes.Buffer
		posts   []post
	)

	for _, path := range pages {
		segs := strings.Split(path, "/")
		dir := segs[0] + "/" + segs[1]
		date := strings.Split(segs[2], ".")[0]
		year := strings.Split(date, "-")[0]
		title := strings.ReplaceAll(caser.String(segs[1]), "-", " ")
		link := filepath.Join(dir, dst)
		contentMain := strings.Replace(tplMain, "{TITLE}", title, 1)
		contentMain = strings.Replace(contentMain, "{LINK}", link, 1)
		contentMain = strings.Replace(contentMain, "{DATE}", date, 1)
		if year != "index" {
			// Avoid adding generic pages to the home page list of pages in the
			// main section (generic pages are fine to add to side nav `links`)
			posts = append(posts, post{date: date, year: year, content: contentMain})
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
