# treex: document your project
  
  **treex**  is  in-locus documentation that's easy to write , explore and extend:

**treex** renders annotated file trees for documentation on the shell, like  those
in a projects readmes. it allows you to use this information in the command line,
to export to markdown though `.info` files: dead simple plain files co-located with the files they document, making it easy to keep in sync.

```bash
# annotate your source tree in a simple plain text file
$ echo "cmd/ Command Line Utilities" >> .info
# or use treex helpers
$ treex add docs/guides "In-depth guides for development"
$ treex 
    my-project
    ├─ cmd/                    Command line utilities
    ├─ docs/                   
    │  └─ guides/             In-depth guides for development
        
# export to mardown 
$ treex --format markdown >> README.md
# update your .info to reflect changes in the file system 
$ treex sync
# .info file are as simple as they come: 
$ cat .info 
cmd/ Command line utilities
docs/guides In-depth guides for development
```

Annotated trees are useful for documentation but being a part of the codebase they become a shore to maintain and not available where you need it: exploraing though the shell.

**treex** keeps annotations reads .info files where each line is a  <path> <annotation>  and generates annotated trees  right in you shell as you work. `.info` files can be source controlled and kept next to the files they document, keeping thing local and in sync

## Quick Start

**treex** will render .info files, plain text such as :

```text
   src/main.py The entry point for the application
   docs/README.md Project documentation
   cmd/app Main application executable
```

It also has convenience tools for easier documentation:

   ```bash
   # generates the .info with the paths specified
   **treex** init src/core build scripts/deploy.sh
   # add an annotation for a given path
   treex add tests/setup "Make sure this is ran before any tests"
   # verify a .info file
   treex check
   # verify it it's out of sync, pruning removed paths: 
   treex sync
   ```

You can render markdown

```bash
   treex --format markdown > README.md
```

## Info Files

**treex** uses `.info` files with a simple format:

```text
<path> <description>
```

For paths containing spaces, use the colon format:

```text
<path with spaces>: <description>
```

These files can be distributed throughout your project, keeping documentation close to the code it describes.

**treex** recursively finds and combines them when rendering your project map.

## Customizing output

### Filtering

By default, if a .gitignore file is found, **treex** will honor it and won't show any file in it's patterns. You can change this with:

* `--ignore-file <file>` to use a different ignore file
* `--no-ignore` to not use any ignore file

By default, **treex** looks for `.info` files. You can use a different filename with:

* `--info-file <filename>` to use a custom info file name (e.g., `--info-file .project-info`)

Most trees are long and deep, and we rarely want to document **everything**. Hence **treex** has three modes that define how it shows trees:

### View Modes (`--show`)

* **mix** (default): Shows all annotated paths plus 2 nodes per directory for context, truncating if needed.
  * Always displays all files and directories that have annotations
  * For each directory, shows 2  unannotated items to give you a sense of what else is there
  * Intelligently selects which unannotated items to show, preferring files over directories
* **annotated**: Shows only paths that have annotations. This mode:
  * Displays a minimal tree containing only annotated items
  * Hides all unannotated files and directories
  * Useful when you want a clean view of just your documented components
  
* **all**: Shows every file and directory. This mode:
  * Displays the complete tree structure
  * Can produce very long output for large projects
  * Useful when you need to see everything or are exploring a new codebase

### Output Formats

* **color** (default): Rich, colored output for your shell
* **no-color**: Plain text output without colors
* **markdown**: Perfect for README files and documentation

### Customizing Colors and Styles

**treex** supports full customization of colors and text styles through a `treex.yaml` configuration file. You can:

* Customize colors for both light and dark terminal themes
* Control text styling (bold, faint) for different elements
* Override specific colors while keeping others as defaults

See the included `treex.yaml` file for a fully documented example of all available options, or check `docs/configuration.md` for details.

## Commands

### `treex`

Render your project map. Works from any directory in your project.

* **`treex init <path1> <path2> ... <pathN>`**:  Create a new `.info` file with the specified paths, ready for you to annotate.
* **`treex add <path> <description>`**: Add or update an annotation for a specific path.
* **`treex rm <path>`**: Remove the annotation for a specific path from the `.info` file.
* **`treex sync`**: Remove annotations for non-existent paths from all `.info` files (use `--force` to skip confirmation).
* **`treex search <term>`**: Search for a term in all `.info` files (searches both paths and annotations).
* **`treex config`**: Output the default configuration file with all options documented.

### `treex draw`

Create tree diagrams from .info format without requiring filesystem paths to exist. Perfect for documentation diagrams and conceptual structures.

* **`treex draw --info-file family.txt`**: Draw a tree from a specific info file
* **`treex draw < organization.txt`**: Draw from stdin input  
* **`cat diagram.info | treex draw`**: Draw from piped input

The draw command uses the same rendering pipeline as the main treex command, supporting all output formats (color, no-color, markdown) but bypasses filesystem warnings since paths are conceptual rather than real filesystem paths.

## Installation

```bash
brew install treex
```

Or download a `.deb` package from the [releases](https://github.com/username/treex/releases).

## Contributing

Bug reports, feature requests or just plain feedback is very welcome, just open an issue.

## License

[MIT](LICENSE)
