package blog

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Post struct {
	Title string
	Date  string
	Year  string
	Dir   string
	Path  string
}

func NewPost(path string) (*Post, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer f.Close()

	var (
		title, date   string
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

	return &Post{
		Title: title,
		Date:  date,
		Year:  year,
		Dir:   dir,
		Path:  path,
	}, nil
}
