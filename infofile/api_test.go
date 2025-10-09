package infofile

// This file has been reorganized. Tests have been moved to thematic files:
//
// - edit_test.go: InfoFile Add/Remove/Update operations (pure unit tests)
// - validation_test.go: InfoFileSet validation and cleaning operations (pure unit tests)
// - multifile_test.go: InfoFileSet gather/distribute/merge operations (pure unit tests)
// - api_fs_test.go: InfoAPI filesystem integration tests
//
// All core logic is now tested as pure functions without filesystem dependencies.
// The api_fs_test.go file contains only integration tests that verify the API
// correctly calls the underlying InfoFileSet operations and handles filesystem I/O.
