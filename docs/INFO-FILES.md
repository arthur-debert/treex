# INFO-FILES

treex reads it's data from .info files in the file system. These can be in any directory, and treex will merge the final tree. If there are multiple entries for the same file, the deepest level .info (the file closer to the entry's path) takes precedence. 

# .info Files Guide

## Overview

`.info` files are the core feature of treex that allow you to annotate files and directories with descriptive text. Each directory can contain its own `.info` file describing the files and subdirectories within it.

## Basic Format

Each `.info` file contains path-description pairs:

```
<path>
<description>

<path>
<description>
...
```

### Example
```
README.md
Main project documentation file

src/main.go
Application entry point with command line handling

config.json
Configuration settings for the application
```

## Nested .info Files

**treex supports nested `.info` files** - any directory in your project can have its own `.info` file:

```
project/
├── .info              # Describes files in project/
├── README.md
├── main.go
└── internal/
    ├── .info          # Describes files in internal/
    ├── parser.go
    ├── utils.go
    └── deep/
        ├── .info      # Describes files in deep/
        ├── config.json
        └── data.db
```

### Path Resolution

Each `.info` file can **only describe paths within its own directory**:

```
# In project/.info
README.md              ✅ Valid (in same directory)
internal/parser.go     ❌ Invalid (should be in internal/.info)
../parent.txt          ❌ Invalid (security violation)

# In project/internal/.info  
parser.go              ✅ Valid (in same directory)
utils.go               ✅ Valid (in same directory)
deep/config.json       ✅ Valid (subdirectory)
../README.md           ❌ Invalid (parent directory)
```

## Description Formats

### Single Line Description
```
file.txt
A simple text file
```

### Multi-line Description (with title)
```
complex-file.js
Complex JavaScript Module
This file handles advanced data processing
and includes multiple utility functions
for data transformation and validation.
```

The first line becomes the **title** (shown inline), and subsequent lines provide detailed description.

### Multi-line Description (no title)  
```
simple-script.sh
This is a simple shell script
that performs basic file operations
```

When there's no clear title structure, the first line is used as the inline annotation.

## Best Practices

### 1. Organize by Directory Scope
Place `.info` files close to the files they describe:

```
✅ Good:
src/
├── .info          # Describes src/ contents
├── main.go
└── utils/
    ├── .info      # Describes utils/ contents  
    └── helper.go

❌ Avoid:
.info              # Trying to describe everything from root
```

### 2. Keep Descriptions Focused
Each `.info` file should focus on its directory's contents:

```
✅ Good - internal/.info:
parser.go
Handles .info file parsing

builder.go
Constructs file trees

❌ Avoid - internal/.info:
parser.go
Handles .info file parsing

../README.md           # Don't reference parent files
```

### 3. Use Meaningful Titles
For multi-line descriptions, make the first line a clear title:

```
✅ Good:
api-client.go
HTTP API Client
Provides methods for communicating with external APIs
including authentication, retry logic, and error handling.

❌ Less clear:
api-client.go
This file contains functions that are used to make HTTP requests
to various external services and handle the responses appropriately.
```

## Security Features

### Directory Traversal Protection
treex automatically filters out dangerous paths:

```
# These paths are automatically ignored:
../../../etc/passwd    # ❌ Filtered out
../parent.txt          # ❌ Filtered out  
./file.txt             # ✅ Allowed
subdir/file.txt        # ✅ Allowed
```

### Scope Enforcement
Each `.info` file can only affect its own directory tree:

- ✅ Files in the same directory
- ✅ Files in subdirectories  
- ❌ Files in parent directories
- ❌ Files in sibling directories

## Command Line Usage

```bash
# Use nested .info files (default)
treex .

# See verbose output showing all parsed .info files
treex --verbose .

# Alternative: use only root .info file (legacy mode)
# This requires using the older API directly
```

## Examples

### Project Documentation Structure
```
docs/
├── .info
├── README.md
├── api/
│   ├── .info
│   ├── endpoints.md
│   └── authentication.md
└── guides/
    ├── .info
    ├── quickstart.md
    └── deployment.md
```

### docs/.info
```
README.md
Main documentation index

api/
API documentation and reference

guides/
User guides and tutorials
```

### docs/api/.info  
```
endpoints.md
Complete API endpoint reference

authentication.md
Authentication and authorization guide
```

### docs/guides/.info
```
quickstart.md
Quick Start Guide
Get up and running in 5 minutes with basic setup
and configuration examples.

deployment.md
Production Deployment Guide
Step-by-step instructions for deploying to production
including security considerations and monitoring setup.
```

## Migration from Single .info Files

If you have an existing single `.info` file, you can gradually migrate to nested files:

1. **Start with existing root `.info`**
2. **Create subdirectory `.info` files** for areas that need more detail
3. **Move relevant entries** from root to appropriate subdirectory files
4. **Remove entries from root** that are now covered by nested files

The system will automatically merge all `.info` files, so you can migrate incrementally. 
