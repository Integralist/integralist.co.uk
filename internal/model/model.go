package model

import (
	"html/template"
	"math"
	"regexp"
	"strings"
	"time"
)

type Post struct {
	Author        string
	Content       template.HTML
	Date          time.Time
	Description   string
	Image         string
	ImagePosition string
	JS            []string
	Keywords      []string
	MarkdownURL   string
	Slug          string
	SourceMD      []byte
	Tags          []string
	Title         string
	URL           string
}

func (p Post) ReadingTime() int {
	words := len(strings.Fields(string(p.SourceMD)))
	return int(math.Ceil(float64(words) / 200))
}

type Page struct {
	Content       template.HTML
	Description   string
	Image         string
	ImagePosition string
	Keywords      []string
	MarkdownURL   string
	NavOrder      int
	Slug          string
	SourceMD      []byte
	Title         string
	URL           string
}

type Tag struct {
	Name  string
	Slug  string
	Posts []*Post
	URL   string
	Color string
}

type Site struct {
	BaseURL string
	Posts   []*Post
	Pages   []*Page
	Tags    []*Tag
}

var (
	nonAlphaNum = regexp.MustCompile(`[^a-z0-9-]+`)
	multiDash   = regexp.MustCompile(`-{2,}`)
)

// Slugify converts a string into a URL-safe slug.
func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonAlphaNum.ReplaceAllString(s, "-")
	s = multiDash.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}
