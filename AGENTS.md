## Why

integralist.co.uk is a technical blog covering software engineering topics
(Go, Python, Rust, networking, security, infrastructure, DevOps, and more). It
runs on a bespoke static site generator (SSG) written in Go, with no frameworks
and no JavaScript.

## What

The SSG reads Markdown from `content/`, applies Go `html/template` templates
from `assets/templates/`, and writes static HTML to `public/` (git-ignored).

Key entry points:

- `cmd/ssg/main.go` — build binary; generates the site.
- `cmd/server/main.go` — local dev server (`localhost:8080`).
- `internal/builder/` — orchestrates the build pipeline.
- `internal/content/` — loads posts/pages from disk, assembles tags.
- `internal/parser/` — YAML frontmatter extraction + Markdown-to-HTML.
- `internal/renderer/` — template execution and file writing.
- `internal/model/` — domain types: `Post`, `Page`, `Tag`, `Site`.

See `README.md` for the full frontmatter schema, project structure, deployment
details, and agent/LLM support.

## How

```bash
make build   # compile the SSG binary
make run     # build + generate site into public/
make serve   # build + generate + start dev server at :8080
make test    # go test ./...
make clean   # rm -rf public ssg
```

Deployed via Netlify on push to `main` (`netlify.toml` runs `make run`).

## Writing style

When editing or rewriting blog content, follow these rules:

- No em dashes. Use commas, parentheses, or restructure the sentence instead.
- Keep swear words. They're intentional and convey genuine feeling.
- Light sarcasm and humour are welcome, especially in section headers.
- Don't remove the author's voice. Conversational tone throughout.
- Don't invent facts. Only improve flow, clarity, and structure.
- Parenthetical asides are fine but keep them short. If one runs past two
  lines, break it into its own sentence.

## Gotchas

- Every HTML page gets a companion `index.md` with raw Markdown source (for
  LLM consumption). The HTML `<head>` includes a
  `<link rel="alternate" type="text/markdown">` pointing to it.
- Tags auto-generate index pages at `/tags/{slug}/` and cycle through four
  fixed colours (coral, amber, teal, blue).
- Pages render at the root (`about.md` -> `/about/`), posts under `/posts/`.
- The build generates `robots.txt`, `sitemap.xml`, `llms.txt`, and `rss.xml`
  as discovery files. RSS includes full post HTML content.
- The optional `image` frontmatter field on posts sets both the OG/Twitter
  meta image and renders a clickable hero image at the top of the post.
