```text
my-project/
├─ cmd/
│  └─ myapp/
│     └─ main.go
├─ internal/
│  ├─ server/
│  │  ├─ server.go
│  │  └─ server_test.go
│  └─ parser/
│     ├─ parser.go
│     ├─ parser_test.go
│     └─ testdata/
│        └─ fixture.json
├─ pkg/
│  └─ public-api/
│     └─ client.go
├─ go.mod
├─ go.sum
└─ Makefile
```

cmd/ - Your Application's Entry Points
Purpose: Contains the main packages for your executables.
Rule: Create one subdirectory for each binary you want to build.
Example: cmd/myapp/main.go is the entry point for the myapp binary.
Code here should be minimal. It should parse flags/config and call into packages in internal/ or pkg/ to do the real work.
internal/ - Private Application Code
Purpose: The bulk of your project's logic. This is the default place for most of your code.
Rule: The Go toolchain prevents other projects from importing packages from your internal/ directory. It's private to your project.
Structure: Group code by functionality. For example, internal/server for HTTP server logic, internal/database for database interaction.
pkg/ - Public Library Code (Use With Caution)
Purpose: Code that is safe to be imported and used by external applications.
Rule: Do not put your code here by default. Only move a package from internal/ to pkg/ when you have a clear requirement for another project to import it. Most applications will never need a pkg/ directory.
