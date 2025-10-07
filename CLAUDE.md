# treex

`treex` is a command line utility for displaying file trees, a modernized view of the unix tree tool.
This phase of the project will not tourch or deal wiith the cli app, but instead focus on the core engine and data structures.

We'll be working on the tree builder part, the code that recives the root path we want to view and set of filters and conditions and returns a tree of `treex/types.go Node` structs..

## FileMap (generated with previous version of treex)

Treex will store annotations in .info files , so always check and use `.info` when relevant.

```text
treex
├─ dist                                 All build / distribution should be done inside this dir
│  ├─ completions                       Shell completions
│  ├─ man                               Man page
│  └─ bin
│     └─ treex                          Binary treex file
├─ scripts                              Codebase utilities / dayliy functions
│  ├─ build                             Builds the binary into dist/bin
│  ├─ gen-completion                    IGNORE for now: will generate the cli completion into dis/completions
│  ├─ gen-manpage                       IGNORE for now: will generate the cli man into dis/man
│  ├─ lint                              Runs linting, called by pre-commit
│  ├─ loc-tree                          Shows loc separating app and test code
│  ├─ pre-commit                        Hook for git: lint + tests
│  ├─ release                           Uses goreleaser to generate the full dist
│  ├─ test                              Runs tests, called by pre-commit
│  ├─ coverage
│  └─ fmt
├─ docs
│  └─ dev
│     └─ architecture.md
└─ treex                                 Codebase entry point
   ├─ cmd                        .       Command line entry point
   ├─ treebuilder                       Package for generating the node tree
   │  └─ filesystem.go
   └─ types                             Package for types, only the core 3 types should go here, watch for type creep
      ├─ annotations.go
      └─ node.go
```

## Tooling

- scripts/build -> builds should
- scripts/test -> runs tests
- scripts/lint

## NO CLI

  For now there will be no shell work, we will implement the core enine at the data structure level.
  The only reson we want the build, release and other infrascture setup is to have the layout ready .

## No Backwards Compatibility

  This is unreleased code. There are no integrations, no third parties, no users. Hence it makes no sense for us to build backard compatibility, adapters, and on son. Each of these has a cost on complexity that compounds and in this codebase it has no gain to be had.

## Testing

- Its critical that we do unit testing. The entire code base is designed functionally.
- For this we need to dependency inject the file systme operations so we can mock this smartly on tests.
- Whenever possible file system can be in one isolated function that's injectable, and all other logic cna be tested easily.

## Project Scope

  This is detailed in the [architecture docs](./docs/dev/architecture.txt)

## Standards

  All documentation for the project is to be written in .txt , no markdown. See the [architecture] (./docs/dev/architecture.txt) for a formatting reference
