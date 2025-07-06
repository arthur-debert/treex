# treex maps for your projects

In locus documentation that's easy to write , explore and extend:

```bash
# annotate your source tree in a simple plain text file
$ cat .info # goo-ole plain text, as simple as it gets
cmd: Command Line Utilities
docs/guides: User guides and tutorials

$ treex 
    my-project
    ├── cmd/                    Command line utilities
    ├── docs/                   
    │   └── guides/             User guides and tutorials

```

These are very useful for documentation and exploration but are time consuming to generate, will out sync actual file structure and are not available when you most use it: in the shell when working on the codebase.

treex reads .info files, plain text files in the format <path>:<annotation> and generates annotated trees, right in you shell as you work. .info files can be source controlled and kept next to the files they document, keeping thing local and in syn.

## Installation

```bash
brew install treex
```

Or download a `.deb` package from the [releases](https://github.com/username/treex/releases).

## Quick Start

treex will render .info files, like :

1. **Initialize** your project documentation:

   ```bash
   treex init src/core build scripts/deploy.sh
   ```

2. **Edit** the generated `.info` file:

   ```text
   src/core: Core application code
   build: Build scripts and artifacts
   scripts/deploy.sh: Production deployment script
   ```

3. **View** your project map:

   ```bash
   treex
   ```

   ```text
   my-project
   ├── src/
   │   └── core/               Core application code
   ├── build/                  Build scripts and artifacts
   ├── scripts/
   │   └── deploy.sh           Production deployment script
   └── README.md
   ```

## How It Works

treex uses `.info` files with a simple format:

```text
<path>: <description>
```

These files can be distributed throughout your project, keeping documentation close to the code it describes. treex recursively finds and combines them when rendering your project map.

## Commands

### `treex`

Render your project map. Works from any directory in your project.

### `treex init <path1> <path2> ... <pathN>`

Create a new `.info` file with the specified paths, ready for you to annotate.

### `treex add <path>: <description>`

Add or update an annotation for a specific path.

### `treex maketree`

Generate the actual file/directory structure from your `.info` file. Useful for scaffolding new projects.

## Output Formats

- **Terminal**: Rich, colored output for your shell
- **Markdown**: Perfect for README files and documentation
- **HTML**: For web publishing
- **Plain text**: Simple, universal format

Use `treex --help` for format options and more commands.

## Examples

```bash
treex init cmd pkg internal docs
# add detailed annotations
treex add cmd: command line applications
treex add pkg: public library code
treex add internal: private application code
treex add docs: project documentation
# generate markdown for your readme
treex --format markdown >> readme.md
```

## Why treex?

- **Simple**: Just paths and descriptions, nothing fancy
- **Flexible**: Distribute `.info` files anywhere in your project
- **Accessible**: View your project map from any directory
- **Maintainable**: Keep documentation close to code
- **Universal**: Works with any programming language or project type

## Documentation

For more details, see the [documentation](docs/) directory:

- [Installation Guide](docs/INSTALLATION.txxt)
- [Feature Overview](docs/OPTIONS.txxt)
- [Info Files Format](docs/INFO-FILES.txxt)
- [Development Guide](docs/DEVELOPMENT.txxt)

## Contributing

See [DEVELOPMENT.txxt](docs/DEVELOPMENT.txxt) for development setup and contribution guidelines.

## License

[MIT](LICENSE)
