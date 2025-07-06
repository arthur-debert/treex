treex maps for your projects

We've seen (and love) a project file layout explained in it's README. treex  is  in-locus documentation that's easy to write , explore and extend:

  -- treex in action
    
    # annotate your source tree in a simple plain text file
    $ cat .info # goo-ole plain text, as simple as it gets
    cmd Command Line Utilities
    docs/guides User guides and tutorials

    $ treex 
        my-project
        ├── cmd/                    Command line utilities
        ├── docs/                   
        │   └── guides/             User guides and tutorials
  
  -- bash


These are very useful for documentation and exploration but are time consuming to generate, will out sync actual file structure and are not available when you most use it: in the shell when working on the codebase.

treex reads .info files, plain text files in the format <path> <annotation> and generates annotated trees, right in you shell as you work. .info files can be source controlled and kept next to the files they document, keeping thing local and in syn.


1. Quick Start

  treex will render .info files, plain text such as :

    -- A sample info file 
    
    src/main.py The entry point for the application
    docs/README.md Project documentation
    cmd/app Main application executable

    -- text

    --  It also has convenience tools for easier documentation:
      # generates the .info with the paths specified
      treex init src/core build scripts/deploy.sh
      # add an annotation for a given path
      treex add tests/setup "Make sure this is ran before any tests"
      # you can generate the .info file and have treex genrate the files if not present
      treex maketree 
      # verify a .info file
      tree check
      #  if you already have a hand generated map, import it
      tree import myfile

    -- bash

    -- You can render markdown or html for your docs
      
      treex --format markdown > README.md
    
    -- bash


2.Info Files

  These files can be distributed throughout your project, keeping documentation close to the code it describes. treex recursively finds and combines them when rendering your project map.
    
    -- treex uses `.info` files with a simple format:
      <path> <description>
      # For paths containing spaces, use the colon format:
      <path with spaces>: <description>
    -- text


3. Commands

    - treex: Render your project map. Works from any directory in your project.
    - treex init <path1> <path2> ... <pathN>
      Create a new `.info` file with the specified paths, ready for you to annotate.
    - treex add <path> <description>
      Adds the description to the path. Will create an .info file as needed.
    - treex maketree
      Generate the actual file/directory structure from your `.info` file. Useful for scaffolding new projects.


4. Output Formats

  - Terminal: Rich, colored output for your shell
  - Markdown: Perfect for README files and documentation
  - HTML: For web publishing
  - Plain text: Simple, universal format


5. Installation

    -- brew : 
      brew install treex
    -- bash

  Or download a `.deb` package from the [releases](https://github.com/username/treex/releases).


6. Contributing

  Bug reports, feature requests or just plain feedback is very welcome, just open an issue.


7. License

  MIT License
