# Refactor: Source Files & Front Matter

- **Status**: Completed
- **Author**: Integralist
- **Created**: 2026-01-14

## Summary

Refactor the blog's content structure to eliminate file duplication and decouple metadata from filenames. We will rename source files from `YYYY-MM-DD.index.md` to `index.html.md` and move the publish date into YAML Front Matter. This allows the source file to serve double duty: as the source for generating `index.html` and as the direct download target for `llms.txt`.

## Background

Currently, the build process generates a duplicate `index.html.md` file for every post to satisfy the `llms.txt` requirement. This wastes disk space and creates clutter. The user identified that renaming the source file to `index.html.md` would solve this, but the date metadata is currently trapped in the filename.

### Current State
- Source: `posts/topic/2024-01-01.index.md`
- logic: `main.go` parses date from filename.

### Desired State
- Source: `posts/topic/index.html.md`
- Content:
  ```markdown
  ---
  date: 2024-01-01
  ---
  # Post Title
  ```
- logic: `main.go` parses date from file content.

## Implementation Tasks

### Phase 1: Migration Script

- [x] **Task 1.1**: Create a temporary Go program (or function in `main.go`) to walk the directory.
- [x] **Task 1.2**: For every `YYYY-MM-DD.index.md`:
    - Read content.
    - Extract date from filename.
    - Prepend `date: YYYY-MM-DD` front matter.
    - Rename file to `index.html.md`.

### Phase 2: Update Generator (`cmd/blog/main.go`)

- [x] **Task 2.1**: Update `parsePage` to read the first few lines of the file to extract the `date` field from Front Matter.
- [x] **Task 2.2**: Implement a basic Front Matter stripper so the `date` block doesn't appear in the generated HTML.
- [x] **Task 2.3**: Update file discovery to look for `index.html.md` instead of `.md` (or just prioritize it).
- [x] **Task 2.4**: Remove the step in `renderPosts` that writes `index.html.md` (as it is now the source).
- [x] **Task 2.5**: Update `main.go` to add `docs` to the `skipDirs` list or explicitly ignore `docs/projects` to prevent HTML generation for project plans.

### Phase 3: Verification

- [x] **Task 3.1**: Run `make build`.
- [x] **Task 3.2**: Verify `index.html` generates correctly (date is correct, no YAML visible).
- [x] **Task 3.3**: Verify `llms.txt` links still work (they point to the source files now).
