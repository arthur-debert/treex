# E2E Sandbox for Treex

This directory contains the end-to-end testing sandbox infrastructure for Treex.

## Overview

The e2e-sandbox provides a controlled testing environment where you can:
- Test treex against various filesystem structures
- Run interactive testing sessions
- Develop and test scripts against treex

## Files

- run              Main sandbox script
- README.txt       This file
- file-trees/      Sample filesystem structures in JSON format
- test-script.sh   Basic sandbox functionality test
- treex-test.sh    Treex functionality test

## Usage

Basic usage:
  ./run                                    # Interactive shell with empty filesystem
  ./run file-trees/simple.json             # Interactive shell with simple filesystem
  ./run file-trees/complex.json treex-test.sh  # Run script with complex filesystem

## File Trees

The file-trees/ directory contains sample filesystem structures:

- empty.json       Empty filesystem (just the temp directory)
- simple.json      Basic project structure with src/, docs/, .gitignore
- complex.json     Complex project with multiple modules, tests, configs
- with-hidden.json Project with hidden files and directories

## How It Works

1. Builds treex binary (if needed)
2. Builds internal-treex-test-data utility
3. Creates a temporary directory
4. Sets up filesystem structure from JSON (if provided)
5. Sets HOME to temp directory and adds treex to PATH
6. Runs script or interactive shell

## Creating New File Trees

File trees are JSON objects where:
- Keys are file/directory names
- String values are file contents
- Object values are subdirectories
- null values are empty directories

Example:
{
  "src": {
    "main.go": "package main\n\nfunc main() {}\n",
    "lib": {
      "utils.go": "package lib\n"
    },
    "empty-dir": null
  },
  "README.txt": "Project documentation\n"
}

## Environment

In the sandbox:
- HOME is set to the temporary directory
- PATH includes the treex binary first
- Working directory is the temporary directory
- All filesystem changes are isolated and cleaned up automatically

## Testing Scripts

Create scripts in the e2e-sandbox directory to test specific functionality.
Scripts have access to:
- treex command in PATH
- Created filesystem structure
- Standard Unix utilities

The sandbox automatically makes scripts executable and cleans up after execution.