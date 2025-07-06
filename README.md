# treex maps for your projects

We've seen (and appreciate) when project's README give an overiewe of the project by showing a commented file layout. While great, these are usually hand crafted and separate from the files.

**treex**  is  in-locus documentation that's easy to write , explore and extend:

```bash
# annotate your source tree in a simple plain text file
$ cat .info # goo-ole plain text, as simple as it gets
cmd Command Line Utilities
docs/guides User guides and tutorials

$ treex 
    my-project
    ├── cmd/                    Command line utilities
    ├── docs/                   
    │   └── guides/             User guides and tutorials

```

These are very useful for documentation and exploration but are time consuming to generate, will out sync actual file structure and are not available when you most use it: in the shell when working on the codebase.

treex reads .info files, plain text files in the format <path> <annotation> and generates annotated trees, right in you shell as you work. .info files can be source controlled and kept next to the files they document, keeping thing local and in syn.

## Quick Start

treex will render .info files, plain text such as :

```text
   src/main.py The entry point for the application
   docs/README.md Project documentation
   cmd/app Main application executable
```

It also has convenience tools for easier documentation:

   ```bash
   # generates the .info with the paths specified
   treex init src/core build scripts/deploy.sh
   # add an annotation for a given path
   treex add tests/setup "Make sure this is ran before any tests"
   # verify a .info file
   treex check
   #  if you already have a hand generated map, import it
   treex import myfile
   ```

You can render markdown or html for your docs

```bash
   treex --format markdown > README.md
```

## Info Files

treex uses `.info` files with a simple format:

```text
<path> <description>
```

For paths containing spaces, use the colon format:

```text
<path with spaces>: <description>
```

These files can be distributed throughout your project, keeping documentation close to the code it describes. treex recursively finds and combines them when rendering your project map.

## Customizing output

### Filtering

By default, if a .gitignore file is found, treex will honor it and won't show any file in it's patterns. You can change this with:
* `--ignore-file <file>` to use a different ignore file
* `--no-ignore` to not use any ignore file

By default, treex looks for `.info` files. You can use a different filename with:
* `--info-file <filename>` to use a custom info file name (e.g., `--info-file .project-info`)

Most trees are long and deep, and we rearely want to document **everything**. Hence treex has three modes that define how it shows trees:

* mix: (default) show all anottated paths, plus a few others per dir for context
* annotated: only shows annotated paths
* all: shows all paths (output can be very long )

### Output Formats

* **color** (default): Rich, colored output for your shell
* **no-color**: Plain text output without colors
* **markdown**: Perfect for README files and documentation

## Commands

### `treex`

Render your project map. Works from any directory in your project.

* **`treex init <path1> <path2> ... <pathN>`**:  Create a new `.info` file with the specified paths, ready for you to annotate.
* **`treex add <path> <description>`**: Add or update an annotation for a specific path.
* **`treex rm <path>`**: Remove the annotation for a specific path from the `.info` file.
* **`treex sync`**: Remove annotations for non-existent paths from all `.info` files (use `--force` to skip confirmation).
* **`treex search <term>`**: Search for a term in all `.info` files (searches both paths and annotations).

## Installation

```bash
brew install treex
```

Or download a `.deb` package from the [releases](https://github.com/username/treex/releases).

## Contributing

Bug reports, feature requests or just plain feedback is very welcome, just open an issue.

## License

[MIT](LICENSE)
