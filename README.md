# treex

**treex** is a file viewer that displays annotations visually.

Ever joined a new project and felt lost in a sea of files and directories? `treex` provides a living map of your codebase, helping you and your team understand the architecture at a glance.

Imagine exploring a new project for the first time. Instead of just a list of files, you get this:

```text
my-web-app
├── .github/                CI/CD workflows
│   └── workflows/
│       └── release.yml     Handles automated deployments to production
├── .gitignore
├── Dockerfile              Containerizes the app for production. Uses a multi-stage build.
├── README.md               You are here!
├── api/                    Backend services (Express.js)
│   ├── .info
│   ├── package.json
│   └── server.js           Main API server file. Defines all routes.
├── package.json            Manages Node.js dependencies for both frontend and backend.
└── web/                    Frontend application (React)
    ├── .info
    ├── package.json
    └── src/
        ├── App.js          The root of our React app.
        └── components/
            └── Login.js    The main login component. Connects to the `/api/server.js` endpoint.
```

This annotated view is powered by simple `.info` files you can check into your repository, making project knowledge accessible and easy to maintain.

## How It Works

`treex` looks for `.info` files in the directories it scans. These files contain simple, Markdown-like annotations for files and directories.

Here's the content of the `web/.info` file from the example above:

```plaintext
# web/

# This is the main directory for our React single-page application.
# It has its own package.json for managing frontend dependencies.

App.js
The root of our React app.

components/Login.js
The main login component. Connects to the `/api/server.js` endpoint.
```

It's just a path followed by its description. That's it!

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

#### Debian Package (.deb)

If you are on a Debian-based Linux distribution like Ubuntu, you can install the `.deb` package directly from our GitHub releases:

```bash
# Download the latest .deb package (replace with the latest version)
wget https://github.com/arthur-debert/treex/releases/latest/download/treex_*_Linux_x86_64.deb

# Install the package
sudo dpkg -i treex_*_Linux_x86_64.deb

# If there are dependency issues, fix them with:
sudo apt-get install -f
```

You can also browse all releases at [GitHub Releases](https://github.com/arthur-debert/treex/releases) and download the specific version you need.

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
# Show the annotated tree for the current directory
treex

# Show the tree for a specific path
treex path/to/your/project

# Get help on all available flags
treex --help
```

### Working with .info Files

`treex` provides several ways to create and manage `.info` files for your projects:

#### 1. Manual Editing

The simplest way is to create `.info` files manually in any directory:

```bash
# Create a .info file in the current directory
nano .info
```

Example `.info` content:

```text
cmd/
Command line utilities and main application entry points.

docs/
All project documentation including user guides and API references.

README.md
Main project documentation. Start here for an overview.
```

#### 2. Interactive Addition with `add-info`

Add descriptions for specific files or directories interactively:

```bash
# Add a description for a specific file or directory
treex add-info src/main.go "Main application entry point with CLI setup"

# Add a description for a directory
treex add-info config/ "Configuration files and environment settings"
```

This command will:

- Create a `.info` file in the appropriate directory if it doesn't exist
- Add or update the entry for the specified path
- Prompt you if an entry already exists (replace, append, or skip)

#### 3. Bulk Generation with `gen-info`

Generate multiple `.info` files from a hand-written annotated tree structure:

```bash
# Generate .info files from an annotated tree structure
treex gen-info my-tree-structure.txt
```

This command takes a tree-like input file (like those commonly found in project documentation) and automatically creates `.info` files in the appropriate directories.

**Example input file:**

```text
my-project
├── cmd/                    Command line utilities
├── docs/                   All documentation
│   └── guides/             User guides and tutorials
├── pkg/                    Core application code
├── scripts/                Build and deployment scripts
└── README.md               Main project documentation
```

**Generated output:**

- `.info` (root level)
- `my-project/.info` (with entries for cmd/, docs/, pkg/, scripts/, README.md)
- `my-project/docs/.info` (with entry for guides/)

**Features:**

- **Flexible parsing**: Handles various tree connector styles (├──, └──, |, etc.) and spacing
- **Path validation**: Provides clear error messages if referenced paths don't exist
- **Smart organization**: Creates `.info` files in the correct parent directories
- **Directory detection**: Automatically detects directories (with trailing `/`) vs files
- **Loose format support**: Works with hand-crafted trees from documentation

## Development

Interested in contributing? Check out the [Development Guide](docs/DEVELOPMENT.md) to get started.

## License

[MIT](LICENSE)
