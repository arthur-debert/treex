treex displays directory trees with annotations from .info files.

Annotations are read from .info files in directories and displayed
alongside the file tree structure, similar to the unix tree command
but with additional context and descriptions for files and directories.

.INFO FILES:

.info files are simple text files that describe the contents of directories.
Each directory can contain its own .info file to document files and subdirectories.

Basic format:
    filename
    Description of the file

    directory/
    Description of the directory

Example .info file:
    README.md
    Main project documentation

    src/main.go
    Application Entry Point
    Handles command line arguments and initializes the application.

    config/
    Configuration files and settings

FORMATS:

treex supports multiple output formats for different use cases. Use --format=<name> to specify:

Terminal formats (for display):
  color           Full color terminal output with beautiful styling (default)
                  Aliases: colorful, full
  minimal         Minimal color styling for basic terminals  
                  Aliases: simple
  no-color        Plain text output without colors
                  Aliases: plain, text

Data formats (for automation and processing):
  json            JSON structured data format
  yaml            YAML structured data format
                  Aliases: yml
  compact-json    Compact JSON format (no indentation)
                  Aliases: compact
  flat-json       Flat JSON array of paths with metadata
                  Aliases: flat

Markdown formats (for documentation):
  markdown        Markdown format with clickable file links
                  Aliases: md
  nested-markdown Nested Markdown with sections and table of contents
                  Aliases: nested-md
  table-markdown  Markdown with table layout
                  Aliases: table-md

HTML formats (for web display):
  html            Interactive HTML with expandable tree
                  Aliases: interactive
  compact-html    Compact HTML format
                  Aliases: compact-web
  table-html      HTML with table layout

Special formats:
  simplelist      Simple indented list of file and directory names
                  Aliases: slist

Examples:
  treex                           # Default color output
  treex --format=json > tree.json # Export as JSON
  treex --format=minimal .        # Minimal colors for basic terminals
  treex --format=markdown > README.md  # Generate markdown documentation
  treex --format=no-color > tree.txt   # Plain text for files
  treex --format=yaml | less      # YAML output with pager

NESTED .INFO FILES:

treex supports nested .info files - any directory can have its own .info file:
    project/.info          # Describes project/ contents
    project/src/.info      # Describes src/ contents  
    project/docs/.info     # Describes docs/ contents

Each .info file can only describe paths within its own directory for security.

GENERATING .INFO FILES:

Use 'treex import <file>' to generate .info files from annotated tree structures.
The input can be simple:
    myproject/cmd The CLI utilities
    myproject/docs Documentation

Use 'treex info-files' for a quick reference guide. 