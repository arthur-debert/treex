# treex: document your project
  
  **treex**  is  in-locus documentation that's easy to write , explore and extend:

**treex** renders annotated file trees for documentation on the shell, like  those
in a projects readmes. it allows you to use this information in the command line,
to export to markdown though `.info` files: dead simple plain files co-located with the files they document, making it easy to keep in sync.

```bash
# annotate your source tree in a simple plain text file
$ echo "cmd/ Command Line Utilities" >> .info
# or use treex helpers
$ treex info-add docs/guides "In-depth guides for development"
$ treex 
    my-project
    ├─ cmd/                    Command line utilities
    ├─ docs/                   
    │  └─ guides/             In-depth guides for development
        
# export to mardown 
$ treex --format markdown >> README.md
# update your .info to reflect changes in the file system 
$ treex info-sync
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
   **treex** info-init src/core build scripts/deploy.sh
   # add an annotation for a given path
   treex info-add tests/setup "Make sure this is ran before any tests"
   # verify a .info file
   treex check
   # verify it it's out of sync, pruning removed paths: 
   treex info-sync
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

* **`treex info-init <path1> <path2> ... <pathN>`**:  Create a new `.info` file with the specified paths, ready for you to annotate (use `--force` to overwrite existing file without confirmation).
* **`treex info-add <path> <description>`**: Add or update an annotation for a specific path. Multi-word descriptions no longer require quotes.
* **`treex info-del <path>`**: Delete the annotation for a specific path from the `.info` file (only removes the annotation, not the file itself).
* **`treex info-sync`**: Delete annotations for non-existent paths from all `.info` files (use `--force` to skip confirmation).
* **`treex info-search <term>`**: Search for a term in all `.info` files (searches both paths and annotations).
* **`treex info-edit [path]`**: Open `.info` file(s) in your editor, optionally jumping to a specific annotation line.
* **`treex config`**: Output the default configuration file with all options documented.

### `treex info-draw`

Create tree diagrams from .info format without requiring filesystem paths to exist. Perfect for documentation diagrams, organizational charts, family trees, and conceptual structures.

#### Use Cases
- **Organization Charts**: Visualize company hierarchies and reporting structures
- **Family Trees**: Document genealogy and family relationships  
- **Project Planning**: Design directory structures before implementation
- **Concept Maps**: Create hierarchical documentation diagrams
- **System Architecture**: Illustrate component relationships

#### Examples

```bash
# Create a simple org chart
echo -e "CEO Alice Thompson, Chief Executive\nCTO\tDavid Kim, Chief Technology\nCTO/Engineering/\tEngineering Team" | treex info-draw

# Render from a file
treex info-draw --info-file organization.info

# Export to markdown for documentation
treex info-draw --info-file project-plan.info --format markdown > ARCHITECTURE.md

# Family tree example
cat > family.info << EOF
Grandparents/ The elders
Grandparents/John Born 1950
Grandparents/Mary Born 1952
Parents/ Our generation
Parents/Dad Michael, born 1980
Parents/Mom Sarah, born 1982
EOF
treex info-draw --info-file family.info
```

#### Key Features
- No filesystem validation - paths are purely conceptual
- Supports all treex output formats (color, no-color, markdown)
- Read from files or stdin for maximum flexibility
- Perfect for creating diagrams in documentation
- See example files in `docs/examples/` directory

## Known Issues

- **Working Directory Resolution**: In some environments (like certain Docker containers or CI systems), treex now has improved handling for cases where the working directory cannot be determined. The tool will fall back to using the PWD environment variable or work with relative paths when possible.

## Installation

```bash
brew install treex
```

Or download a `.deb` package from the [releases](https://github.com/username/treex/releases).

## Contributing

Bug reports, feature requests or just plain feedback is very welcome, just open an issue.

## License

[MIT](LICENSE)
