# integralist.co.uk

A technical blog featuring a bespoke Static Site Generator written in Go.

> [!TIP]
> Can also be accessed via [integralist.netlify.app](https://integralist.netlify.app/)

## Architecture

This project uses a custom SSG to convert Markdown files into a
high-performance, SEO-friendly static website.

- **Frontend**: Vanilla HTML and CSS (no frameworks, no JavaScript).
- **Backend**: Custom SSG written in Go (v1.26+), utilizing
  `github.com/gomarkdown/markdown`.
- **Templates**: Standard Go `html/template`.
- **Deployment**: Automatically built and hosted on
  [Netlify](https://www.netlify.com/).

## Getting Started

### Prerequisites

- Go 1.26 or higher
- Make

### Local Development

1. **Build the site**:

   ```bash
   make build
   ```

   This compiles the SSG and generates the static files into the `public/`
   directory.

1. **Preview the site**:

   ```bash
   make serve
   ```

   This builds the site and starts a local development server at
   `http://localhost:8080`.

1. **Run tests**:

   ```bash
   make test
   ```

1. **Clean build artifacts**:

   ```bash
   make clean
   ```

## Project Structure

- `assets/`: CSS, images, and HTML templates.
- `cmd/ssg/`: Entry point for the Static Site Generator.
- `cmd/server/`: Local development server.
- `content/posts/`: Markdown source files for blog posts.
- `content/pages/`: Markdown source files for static pages (nav items).
- `internal/`: Core logic for parsing, rendering, and site building.
- `public/`: The generated static site (Git ignored).

## Writing

### Blog Post

```yaml
---
title: "Post Title"
date: 2026-04-12
description: "A short description for SEO and post listings."
tags: [go, ssg]
keywords: [go, static site generator]
author: "Mark"
image: /assets/img/hero.jpg
image_position: top
---

Content here.
```

- `keywords` is optional — defaults to `tags` if omitted.
- `tags` generate coloured pill badges and index pages at `/tags/{slug}/`.
- `author` is optional. When set, it appears in Twitter Card metadata.
- `image` is optional. When set, it renders as a clickable hero image at the
  top of the post and populates `og:image` / `twitter:image` meta tags (the
  path is relative to site root, e.g. `/assets/img/hero.jpg`).
- `image_position` is optional. Controls `object-position` for the hero image
  crop (default: `center`). Use `top` to crop from the bottom upward.

### Static Page

```yaml
---
title: "About"
description: "About integralist.co.uk and its author."
keywords: [about, personal, blog]
nav_order: 1
image: /assets/img/photo.jpg
image_position: top
---

Content here.
```

- `nav_order` controls the ordering in the top navigation.
- Pages render at the root level (e.g. `about.md` becomes `/about/`).
- `image` and `image_position` work the same as for posts.

## Writing Markdown

When writing Markdown, some linters such as alex, and markdownlint will
complain about various things.

For Alex, you can disable specific warnings using:

```txt
<!--alex ignore foo bar baz-->
```

For Markdownlint, you can disable specific warnings using:

```txt
<!-- markdownlint-disable -->
SOMETHING HERE TO IGNORE
<!-- markdownlint-enable -->
```

GitHub-flavored alert blockquotes are supported:

```md
> [!NOTE]
> Useful information.

> [!WARNING]
> Be careful here.
```

Supported types: `NOTE`, `TIP`, `IMPORTANT`, `WARNING`, `CAUTION`.

## Agent and LLM Support

The site is designed to be easily consumed by AI agents and LLMs. The build
generates several discovery and content files:

### Companion Markdown

Every HTML page has a companion `index.md` file containing the raw Markdown
source. The HTML head includes a
`<link rel="alternate" type="text/markdown">` tag pointing to it.

### Discovery Files

- **`robots.txt`** - Allows all crawlers and includes a `Sitemap:` directive.
- **`sitemap.xml`** - Lists all posts (with `lastmod` dates), pages, tag pages,
  and the homepage.
- **`llms.txt`** - Describes the site and lists every post and page with direct
  links to their companion Markdown files. Follows the
  [llms.txt](https://llmstxt.org/) convention.
- **`rss.xml`** - RSS 2.0 feed with full HTML content for each post. Every
  page includes a `<link rel="alternate" type="application/rss+xml">` tag for
  auto-discovery by feed readers.

## Deployment

The site is deployed via GitHub integration with Netlify. Every push to the
`main` branch triggers a build using `make run`, which compiles the Go binary,
runs it to generate static HTML into `public/`, and then Netlify serves that
directory. The Go binary and templates are only used during the build step and
are not deployed. The generated HTML and companion Markdown files (see
[Agent and LLM Support](#agent-and-llm-support)) are what Netlify serves.

## DNS

Domain is registered with SquareSpace. Two custom DNS records point to
Netlify:

| Type  | Name | Data                          |
| ----- | ---- | ----------------------------- |
| ALIAS | @    | apex-loadbalancer.netlify.com |
| CNAME | www  | integralist.netlify.app       |

## Side Blocks

This static generator supports GFM-style alert blocks. Use the type that best
matches the _intent_ of the aside:

```md
> [!NOTE]
> Supplementary context the reader should know, even when skimming.

> [!TIP]
> Actionable advice that helps the reader do something better.

> [!HINT]
> A playful nudge or Easter egg for attentive readers.

> [!INFO]
> Interesting trivia or background facts (fun facts, history, etymology).

> [!IMPORTANT]
> Key information the reader must not overlook.

> [!ATTENTION]
> Draws the reader's eye to a pattern or detail they might miss.

> [!WARNING]
> Potential pitfall or risk that could cause problems if ignored.

> [!CAUTION]
> Proceed carefully — real consequences if mishandled.

> [!DANGER]
> Serious risk of harm, data loss, or irreversible damage.

> [!ABSTRACT]
> High-level overview or thesis statement for the section that follows.

> [!SUMMARY]
> Condensed recap of preceding content.

> [!TLDR]
> Ultra-short version for readers in a hurry.

> [!EXAMPLE]
> A concrete illustration of the concept just discussed.

> [!CITE]
> A direct quotation with attribution (lyrics, literature, speeches).

> [!QUOTE]
> A paraphrased or indirect reference to another source.

> [!FAQ]
> A commonly asked question and its answer.

> [!QUESTION]
> An open question posed to the reader for reflection.

> [!HELP]
> Guidance for readers who may be stuck or confused.

> [!BUG]
> A known issue or defect being called out.

> [!ERROR]
> Something went wrong — describes the failure.

> [!FAIL]
> A specific test, check, or attempt that did not pass.

> [!FAILURE]
> A broader or more significant breakdown (alias for FAIL).

> [!MISSING]
> Expected content or functionality that is absent.

> [!CHECK]
> A verification step that passed or a condition confirmed.

> [!DONE]
> A task or milestone that has been completed.

> [!SUCCESS]
> A positive outcome worth celebrating.

> [!TODO]
> A task or item that still needs to be addressed.
```

### NOTE vs INFO

- **NOTE** — supplementary context that clarifies or qualifies the surrounding
  text. The reader needs this to correctly understand what they just read.
- **INFO** — interesting but non-essential background (trivia, fun facts,
  historical context). Skipping it doesn't change comprehension.
