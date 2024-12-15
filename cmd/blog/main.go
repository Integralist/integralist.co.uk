package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

const fileBufferSize = 10

var skipDirs = []string{".git", "assets", "cmd"}

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	files := make(chan string, fileBufferSize)
	go renderHTML(files, &wg)

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
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	close(files) // forces renderHTML to complete
	wg.Wait()
}

func renderHTML(files <-chan string, wg *sync.WaitGroup) {
	defer wg.Done() // Mark this goroutine as done when it exits

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
		fmt.Printf("processed: %s: %s\n", path, h)
	}
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
