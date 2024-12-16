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

const (
	channelBufferSize = 10
	numOfProcessors   = 2
)

var skipDirs = []string{".git", "assets", "cmd"}

func main() {
	var wg sync.WaitGroup
	wg.Add(numOfProcessors)

	pages := make(chan string, channelBufferSize)
	links := make(chan string, channelBufferSize)

	go renderSubPages(pages, &wg)
	go renderHomepage(links, &wg)

	// Walk the directory and send file paths to the channel
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
			pages <- path
			links <- path
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	close(pages) // forces renderSubPages to complete
	close(links) // forces renderHomepage to complete
	wg.Wait()
}

// FIXME: Figure out how to render side-nav for sub pages.
func renderSubPages(pages <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	// caser := cases.Title(language.BritishEnglish)

	f, err := os.Open("assets/templates/page.tpl")
	if err != nil {
		err = fmt.Errorf("failed to open page template: %w", err)
		panic(err)
	}

	contentSubPage, err := io.ReadAll(f)
	if err != nil {
		err = fmt.Errorf("failed to read page template: %w", err)
		panic(err)
	}

	needleMainInsert := []byte("{INSERT_MAIN}")
	needleNavInsert := []byte("{INSERT_NAV}")
	// tplNav := `
	// <li>
	//   <span class="opener">{YEAR}</span>
	//   <ul>
	//     <li><a href="../{LINK}">{TITLE}</a></li>
	//   </ul>
	// </li>
	// `

	// type post struct {
	// 	date    string // expects ISO 8601 format, e.g., "2024-12-15"
	// 	year    string
	// 	content string
	// }

	var ( // nolint:prealloc
	// bufNav   bytes.Buffer
	// navLinks []post
	)

	for path := range pages {
		// if strings.Contains(path, "/index.md") {
		// 	continue // non-article pages should be skipped
		// }
		f, err := os.Open(path)
		if err != nil {
			fmt.Printf("failed to open path '%s': %s\n", path, err)
			continue
		}

		md, err := io.ReadAll(f)
		if err != nil {
			fmt.Printf("failed to read file '%s': %s\n", path, err)
			_ = f.Close()
			continue
		}
		_ = f.Close()

		h := mdToHTML(md)
		content := bytes.Replace(contentSubPage, needleMainInsert, h, 1)

		segs := strings.Split(path, "/")
		dir := segs[0]
		// date := strings.Split(segs[1], ".")[0]
		// year := strings.Split(date, ".")[0]
		// title := strings.ReplaceAll(caser.String(dir), "-", " ")
		// link := filepath.Join(dir, "index.html")
		// contentNav := strings.Replace(tplNav, "{YEAR}", year, 1)
		// contentNav = strings.Replace(contentNav, "{LINK}", link, 1)
		// contentNav = strings.Replace(contentNav, "{TITLE}", title, 1)
		// navLinks = append(navLinks, post{date: date, year: year, content: contentNav})
		// sort.Slice(navLinks, func(i, j int) bool {
		// 	// Parse dates for comparison
		// 	date1, _ := time.Parse("2006-01-02", navLinks[i].date)
		// 	date2, _ := time.Parse("2006-01-02", navLinks[j].date)
		// 	return date1.Before(date2) // Ascending order
		// })
		// for _, link := range navLinks {
		// 	_, _ = bufNav.WriteString(link.content)
		// }
		// content = bytes.Replace(content, needleNavInsert, bufNav.Bytes(), 1)
		content = bytes.Replace(content, needleNavInsert, []byte(""), 1)

		dst := filepath.Join(dir, "index.html")
		err = writeFile(dst, content)
		if err != nil {
			fmt.Printf("failed to write file '%s': %s\n", dst, err)
			continue
		}

		fmt.Printf("rendered: %s -> %s\n", path, dst)
	}
}

func renderHomepage(files <-chan string, wg *sync.WaitGroup) { // nolint:revive // function-length
	defer wg.Done()
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
	needleMainInsert := []byte("{INSERT_MAIN}")
	needleNavInsert := []byte("{INSERT_NAV}")
	tplMain := `
	<article>
	<h3>{TITLE}</h3>
	<p class="pubdate">{DATE}</p>
	<ul class="actions">
	<li><a href="{LINK}" class="button">Read</a></li>
	</ul>
	</article>
	`
	tplNav := `
	<li>
	  <span class="opener">{YEAR}</span>
	  <ul>
	    <li><a href="{LINK}">{TITLE}</a></li>
	  </ul>
	</li>
	`
	tplNavGenericPage := `
	<li><a href="{LINK}">{TITLE}</a></li>
	`

	type post struct {
		date    string // expects ISO 8601 format, e.g., "2024-12-15"
		year    string
		content string
	}

	var ( // nolint:prealloc
		bufMain bytes.Buffer
		bufNav  bytes.Buffer
		posts   []post
		links   []post
	)

	for path := range files {
		segs := strings.Split(path, "/")
		dir := segs[0]
		date := strings.Split(segs[1], ".")[0]
		year := strings.Split(date, ".")[0]
		title := strings.ReplaceAll(caser.String(dir), "-", " ")
		link := filepath.Join(dir, dst)
		contentMain := strings.Replace(tplMain, "{TITLE}", title, 1)
		contentMain = strings.Replace(contentMain, "{LINK}", link, 1)
		contentMain = strings.Replace(contentMain, "{DATE}", date, 1)
		var contentNav string
		if year == "index" {
			contentNav = strings.Replace(tplNavGenericPage, "{TITLE}", title, 1)
			contentNav = strings.Replace(contentNav, "{LINK}", link, 1)
		} else {
			contentNav = strings.Replace(tplNav, "{YEAR}", year, 1)
			contentNav = strings.Replace(contentNav, "{LINK}", link, 1)
			contentNav = strings.Replace(contentNav, "{TITLE}", title, 1)
		}
		if year != "index" {
			// Avoid adding generic pages to the home page list of pages in the
			// main section (generic pages are fine to add to side nav `links`)
			posts = append(posts, post{date: date, year: year, content: contentMain})
		}
		links = append(links, post{date: date, year: year, content: contentNav})
	}

	sort.Slice(posts, func(i, j int) bool {
		// Parse dates for comparison
		date1, _ := time.Parse("2006-01-02", posts[i].date)
		date2, _ := time.Parse("2006-01-02", posts[j].date)
		return date2.Before(date1) // Descending order
	})
	sort.Slice(links, func(i, j int) bool {
		// Parse dates for comparison
		date1, _ := time.Parse("2006-01-02", links[i].date)
		date2, _ := time.Parse("2006-01-02", links[j].date)
		return date2.Before(date1) // Descending order
	})

	for _, post := range posts {
		_, _ = bufMain.WriteString(post.content)
	}
	for _, link := range links {
		_, _ = bufNav.WriteString(link.content)
	}

	contentIndex = bytes.Replace(contentIndex, needleMainInsert, bufMain.Bytes(), 1)
	contentIndex = bytes.Replace(contentIndex, needleNavInsert, bufNav.Bytes(), 1)

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
