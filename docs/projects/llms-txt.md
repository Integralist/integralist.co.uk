# llms.txt Integration

- **Status**: Planning
- **Author**: Integralist
- **Created**: 2026-01-14

## Summary

Integrate the [llms.txt](https://llmstxt.org/) standard into `integralist.co.uk` to provide a standardized interface for Large Language Models (LLMs). This involves generating a root-level `/llms.txt` manifest file that lists available content and ensuring that every rendered HTML page has a corresponding raw Markdown version available at `index.html.md` for efficient consumption by AI agents.

> [!NOTE]
> As this is a statically generated website hosted on Netlify, dynamic Content Negotiation using the `Accept` HTTP request header is not possible. To provide LLM-friendly content, we follow the `llms.txt` recommendation of providing a Markdown version of each page at the same URL path but with a `.md` extension (e.g., `index.html.md`).

## Background

The `llms.txt` proposal suggests two key features for AI-friendly websites: a manifest file (`/llms.txt`) providing a roadmap of the site's content, and co-located Markdown files for every HTML page. This allows LLMs to retrieve clean, structured text without parsing complex HTML.

### Current State

The website is statically generated using a custom Go program located in `cmd/blog/main.go`.

- **File Discovery**: `main()` (lines 37-53) walks the directory tree to find `.md` files.
- **Rendering**:
    - `renderPosts` (lines 80-121) reads a source `.md` file, converts it to HTML, and writes it to `path/to/dir/index.html`.
    - It currently *only* writes the `.html` file.
- **Navigation**: `renderSideNav` (lines 145-225) generates the navigation HTML. It contains logic to extract titles and dates from filenames, which is currently coupled to the HTML generation logic.

## Implementation Tasks

### Phase 1: Preparation & Refactoring

- [ ] **Task 1.1**: Refactor `cmd/blog/main.go` to extract the logic for parsing page metadata (Title, Date, Year, Link) from `renderSideNav` into a reusable helper function or struct method. This is necessary because `llms.txt` needs the same metadata.

### Phase 2: Core Implementation

- [ ] **Task 2.1**: Update `renderPosts` in `cmd/blog/main.go` to write a copy of the raw Markdown content to `index.html.md` in the same directory as the generated `index.html`.
- [ ] **Task 2.2**: Implement a new function `renderLLMsTxt(pages []string)` in `cmd/blog/main.go`.
    - This function will generate the `llms.txt` content following the spec:
        - Header with Site Title and Description.
        - "Pages" section.
        - "Posts" section.
    - It will link to the `index.html.md` versions of the files.
- [ ] **Task 2.3**: Call `renderLLMsTxt` within `main()` after page collection.

### Phase 3: Verification

- [ ] **Task 3.1**: Run `make build` to generate the site.
- [ ] **Task 3.2**: Verify `llms.txt` exists at the project root and contains valid links.
- [ ] **Task 3.3**: Verify that for a sample post, an `index.html.md` file exists alongside `index.html`.

## File Changes

| File | Change |
|------|--------|
| `cmd/blog/main.go` | Add `renderLLMsTxt` function; update `renderPosts` to write `.md` files; refactor metadata extraction. |
| `llms.txt` | New generated file at root. |
| `*/index.html.md` | New generated files corresponding to every article. |

## Notes

- The `llms.txt` spec recommends the format: `- [Title](url): Optional description`. We will use the post title and link to the `.md` version.
- We need to ensure that the "Pages" vs "Posts" distinction used in the current side nav is preserved or adapted appropriately for the `llms.txt` structure.