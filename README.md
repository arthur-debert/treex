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

You can install `treex` using brew or  a release .deb.

```bash
# First, add the custom tap, then install
brew tap arthur-debert/tools
brew install treex

# For .deb  download the latest .deb package (replace with the latest version)
wget https://github.com/arthur-debert/treex/releases/latest/download/treex_*_Linux_x86_64.deb
sudo dpkg -i treex_*_Linux_x86_64.deb
sudo apt-get install -f
```

## Usage

```bash
# Show the annotated tree for the current directory, by default respecting .gitignore
treex # honors gitignores
treex path/to/your/project  --depth-4 #  specify path, depth can be changed
treex --help # for more

# adding annotations
treex init # defaults to --depth of 3 not to create a monster, can be overwrittern
treex add <path> <info> # adds info to the ,info
trexx import <path>  # if you have a hand-generated text like this
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

#### 2. Interactive Addition with `add`

Add descriptions for specific files or directories interactively:

```bash
# Add a description for a specific file or directory
treex add src/main.go "Main application entry point with CLI setup"

# Add a description for a directory
treex add config/ "Configuration files and environment settings"
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
