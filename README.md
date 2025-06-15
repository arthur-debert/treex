# treex

treex is a file tree documentation viewer, displaying annotations over a file tree.
This provides a visual map of where things are in a project that can be refreshing.

## Treex

### Annotations

Annotations are powered by a .info file inside a directory

The .info file has the following format:

```text
<path>
<description>

That is, a path (relative to the .info file) in a single line, folowed by any number of lines of text about
the file.

If the description has the form
<text>\n
<text>
....<text>
```

Then the first line (which has a line break ) is taken to be the title ,  short intro for the file.
Descriptions can be preceeded and followed by blank lines, which are ignored .

Paths can be deep, that is in /some/path we can have paths to /some/path/file-a.txt and /some/path/deep/into/file-b.txt

Linebreaks between paths are not required.
See the [info file provided](.info)

### TUI

The ui looks like the unix tree ui, with the annotations:

```text
.
├── .github
│   └── workflows
│       └── go.yml          CI Unit test workflow
│                                        This makes usage of go action, that does pretty much all go setup.
│                                         Note that his has no caching just yet.
├── .gitignore
├── .info
├── LICENSE                 MIT, like most things.
├── README.md           Like the title says, that useful little readme.
├── cmd
├── go.mod
├── go.sum
├── internal
└── pkg
```

### Installation

```bash
go install github.com/adebert/treex/cmd/treex@latest
```

## Usage

```bash
treex --help
```

## Development

### Prerequisites

- Go 1.21 or higher

### Running tests

```bash
go test ./...
```

## License

[MIT](LICENSE)
