
We're making the final adjustments for v1 now that we've got feedback.
For each, make the change, verify that it works and commit.

1. let's remove the code that adds extra spacing between annotated items:
      --extra-spacing            Add extra vertical spacing between annotated items (default true)
not only the option, but the actual code.

2. remove the --safe-mode flag.

3. the command treex info-files display helps on them. currently it renders hard coded strings into the source codce. alter that to read fromm cmd/treex/commands/infofiles.txt

4. remove the options -p (since one can pass the argument for the path to analyze

5. keep but do not print in usage the -v or --verbose options

6. remove the check command. treex will already outout issues as warning, no need. if the regular $treex
show uses the code currently in treex check, keep it, also remove the code for the command too

7. 
Change the treex default help usage. 
It's paramount thta we don not use a custom template, but use cobra's built in. 
Alter the help, best-effort to be as like the one that follows as possible, as long as that doesn't require a custom tmeplate or forgoing cobra's built int help generation

--- THis is the ideal help output: 
treex : renders a documented file tree in your shell. 

Usage:
  treex [path...] [flags]
  treex [command]

Authoring Annotations: 
  init        Initialize a .info file for a directory or specific paths
  add         Add or update an entry in the current directory's .info file
  check       Validate .info files in a directory
  import      Generate .info files from annotated tree structure from markdown

File-system:
  make-tree   Create file/directory structure from .info file

Help and learning:
  formats     List available output formats (--format=NAME)
  help        Help about any command
  info-files  Show information about .info file format and usage

Flags:
  -d, --depth int                Maximum depth to traverse (default 10)
      --format string            color(deault), no-color, markdown (see formats command) 
  -h, --help                     help for treex
      --show string              View mode: mix, annotated, all (default 'mix') (default "mix")
      --use-ignore-file string   Use specified ignore file (default is .gitignore) (default ".gitignore")
      --version                  version for treex

Use "treex [command] --help" for more information about a command.
