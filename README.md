# integralist.co.uk

The `integralist.co.uk` website is statically generated using [Go][1].

Running the target `make build` will:

- Loop over all top-level directories (skipping `cmd`, `assets`).
- Convert every `.md` into a `index.html`.
- Pushes the date for every `.md` (extracted from the filename) into a queue.
- A separate process pulls from the queue and builds a HTML list.

[1]: https://go.dev/
