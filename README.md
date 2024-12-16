# integralist.co.uk

The `integralist.co.uk` website is statically generated using [Go][1].

Running the target `make build` will:

- Loop over all top-level directories (skipping `cmd`, `assets` etc).
- Convert every `.md` into a `index.html`.
- Pushes the date for every `.md` (extracted from the filename) into a queue.
- A separate process pulls from the queue and builds a HTML list.

## Non-article pages

Most pages on the website are "article" pages writing about some topic.

Some are general pages (e.g. "resume") and so they won't have a date prefixed to
the filename. In these cases we render the page as HTML but we don't include the
page dynamically in the list of articles (i.e. you have to manually link them).

## Writing Markdown

When writing Markdown, some linters such as alex, and markdownlint will complain
about various things.

For Alex, you can disable specific warnings using:

```plain
<!--alex ignore foo bar baz-->
```

For Markdownlint, you can disable specific warnings using:

```plain
<!-- markdownlint-disable -->
SOMETHING HERE TO IGNORE
<!-- markdownlint-enable -->
```

## TODO

- Fix resume content.
- Fix side nav on home page to group by year.
- Figure out how to render side nav on sub pages without doubling up a walk.
  - Might need to change channel async rendering approach for a slice of paths.
- Truncate the list of pages on the home page (once I've migrated pages).

[1]: https://go.dev/
