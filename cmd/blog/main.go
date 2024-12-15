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

	files := make(chan string, channelBufferSize)
	index := make(chan string, channelBufferSize)

	go renderHTML(files, &wg)
	go renderIndex(index, &wg)

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
			files <- path
			if !strings.Contains(path, "/index.md") {
				index <- path
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	close(files) // forces renderHTML to complete
	close(index) // forces renderIndex to complete
	wg.Wait()
}

func renderHTML(files <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	f, err := os.Open("assets/templates/page.tpl")
	if err != nil {
		err = fmt.Errorf("failed to open page template: %w", err)
		panic(err)
	}

	bs, err := io.ReadAll(f)
	if err != nil {
		err = fmt.Errorf("failed to read page template: %w", err)
		panic(err)
	}

	needle := []byte("{INSERT_HERE}")

	for path := range files {
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
		content := bytes.Replace(bs, needle, h, 1)

		segs := strings.Split(path, "/")
		dir := segs[0]
		dst := filepath.Join(dir, "index.html")

		err = writeFile(dst, content)
		if err != nil {
			fmt.Printf("failed to write file '%s': %s\n", dst, err)
			continue
		}

		fmt.Printf("rendered: %s -> %s\n", path, dst)
	}
}

func renderIndex(files <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	caser := cases.Title(language.BritishEnglish)

	idx := "assets/templates/index.tpl"
	f, err := os.Open(idx)
	if err != nil {
		err = fmt.Errorf("failed to open index template: %w", err)
		panic(err)
	}

	page, err := io.ReadAll(f)
	if err != nil {
		err = fmt.Errorf("failed to read index template: %w", err)
		panic(err)
	}

	needleMainInsert := []byte("{INSERT_HERE}")
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

	for path := range files {
		segs := strings.Split(path, "/")
		dir := segs[0]
		date := strings.Split(segs[1], ".")[0]
		title := strings.ReplaceAll(caser.String(dir), "-", " ")
		link := filepath.Join(dir, dst)
		content := strings.Replace(tplMain, "{TITLE}", title, 1)
		content = strings.Replace(content, "{LINK}", link, 1)
		content = strings.Replace(content, "{DATE}", date, 1)
		posts = append(posts, post{date: date, content: content})
	}

	sort.Slice(posts, func(i, j int) bool {
		// Parse dates for comparison
		date1, _ := time.Parse("2006-01-02", posts[i].date)
		date2, _ := time.Parse("2006-01-02", posts[j].date)
		return date1.Before(date2) // Ascending order
	})

	for _, post := range posts {
		_, _ = bufMain.WriteString(post.content)
	}

	page = bytes.Replace(page, needleMainInsert, bufMain.Bytes(), 1)

	err = writeFile("index.html", page)
	if err != nil {
		fmt.Printf("failed to write index file: %s\n", err)
		return
	}

	fmt.Printf("rendered: %s -> %s\n", idx, dst)
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
