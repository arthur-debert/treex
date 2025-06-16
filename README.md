# treex

treex is a file tree documentation viewer, displaying annotations over a file tree.
This provides a visual map of where things are in a project that can helpful when navigating new projects.

```bash
$ treex 
brew

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

## Installation

You can install `treex` using a package manager or by downloading a pre-compiled binary.

### Package Managers

#### Homebrew (macOS / Linux)

If you are on macOS or Linux, you can install `treex` using [Homebrew](https://brew.sh/):

```bash
# First, add the custom tap
brew tap arthur-debert/tools

# Now, install treex
brew install treex
```

#### APT (Debian / Ubuntu)

If you are on a Debian-based Linux distribution like Ubuntu, you can install `treex` from our APT repository.

_Note: You will need to replace `your-apt-repo.com` with the actual domain of your repository._

```bash
# 1. Add the repository's GPG key
curl -sS https://your-apt-repo.com/gpg.key | sudo gpg --dearmor -o /usr/share/keyrings/treex-archive-keyring.gpg

# 2. Add the repository to your sources
echo "deb [signed-by=/usr/share/keyrings/treex-archive-keyring.gpg] https://your-apt-repo.com/ ./" | sudo tee /etc/apt/sources.list.d/treex.list > /dev/null

# 3. Update package lists and install treex
sudo apt-get update
sudo apt-get install treex
```

### Manual Installation

You can always download the latest pre-compiled binary for your operating system and architecture from the [GitHub Releases](https://github.com/arthur-debert/treex/releases) page.

1. Download the appropriate archive (e.g., `treex_Linux_x86_64.tar.gz`).
2. Extract the archive: `tar -xzf treex_*.tar.gz`
3. Move the `treex` binary to a directory in your `$PATH`: `sudo mv treex /usr/local/bin/`

### From Source

If you have Go installed, you can build and install `treex` from source:

```bash
go install github.com/arthur-debert/treex/cmd/treex@latest
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
