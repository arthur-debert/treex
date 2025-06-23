package format

import (
	"strings"
	"testing"
)

func TestParseFormatString(t *testing.T) {
	tests := []struct {
		input    string
		expected OutputFormat
		valid    bool
	}{
		// Valid exact formats
		{"json", FormatJSON, true},
		{"yaml", FormatYAML, true},
		{"html", FormatHTML, true},
		{"markdown", FormatMarkdown, true},

		// Valid aliases
		{"yml", FormatYAML, true},
		{"md", FormatMarkdown, true},
		{"plain", FormatNoColor, true},
		{"minimal", FormatMinimal, true},

		// Case insensitive
		{"JSON", FormatJSON, true},
		{"YAML", FormatYAML, true},
		{"Html", FormatHTML, true},
		{"MARKDOWN", FormatMarkdown, true},

		// Invalid formats
		{"invalid", "", false},
		{"", "", false},
		{"xmll", "", false}, // Close but not exact
		{"jsn", "", false},  // Close but not exact

		// Whitespace handling
		{" json ", FormatJSON, true},
		{"\tjson\n", FormatJSON, true},

		// Special formats
		{"compact-json", FormatCompactJSON, true},
		{"flat-json", FormatFlatJSON, true},
		{"nested-markdown", FormatNestedMarkdown, true},
		{"table-markdown", FormatTableMarkdown, true},
		{"compact-html", FormatCompactHTML, true},
		{"table-html", FormatTableHTML, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := ParseFormatString(test.input)

			if test.valid {
				if err != nil {
					t.Errorf("Expected %q to be valid, got error: %v", test.input, err)
				}
				if result != test.expected {
					t.Errorf("Expected %q to parse to %q, got %q", test.input, test.expected, result)
				}
			} else {
				if err == nil {
					t.Errorf("Expected %q to be invalid, got result: %q", test.input, result)
				}
			}
		})
	}
}

func TestListAvailableFormats(t *testing.T) {
	formats := ListAvailableFormats()

	// Should return a non-empty map
	if len(formats) == 0 {
		t.Error("ListAvailableFormats should return at least one format")
	}

	// Should contain expected main formats
	expectedFormats := []OutputFormat{
		FormatJSON,
		FormatYAML,
		FormatHTML,
		FormatMarkdown,
		FormatColor,
	}

	for _, expected := range expectedFormats {
		if _, found := formats[expected]; !found {
			t.Errorf("Expected format %q not found in available formats", expected)
		}
	}

	// All formats should have descriptions
	for format, description := range formats {
		if string(format) == "" {
			t.Error("Found empty format string in available formats")
		}
		if description == "" {
			t.Errorf("Format %q has empty description", format)
		}
	}
}

func TestGetFormatHelp(t *testing.T) {
	help := GetFormatHelp()

	// Should return non-empty help text
	if help == "" {
		t.Error("GetFormatHelp should return non-empty help text")
	}

	// Should contain information about main formats
	expectedFormats := []string{
		"json",
		"yaml",
		"html",
		"markdown",
		"color",
	}

	for _, format := range expectedFormats {
		if !strings.Contains(strings.ToLower(help), format) {
			t.Errorf("Help text should mention %q format", format)
		}
	}

	// Help should be informative but doesn't need to mention all aliases
	// Just check that it's not empty and has basic formatting
	if len(help) < 50 {
		t.Error("Help text should be reasonably detailed")
	}

	// Should be formatted nicely
	if !strings.Contains(help, "\n") {
		t.Error("Help text should contain line breaks for formatting")
	}

	// Should contain some explanation
	if !strings.Contains(strings.ToLower(help), "format") {
		t.Error("Help text should explain what formats are")
	}
}

func TestFormatAliases(t *testing.T) {
	// Test that aliases work correctly for all expected mappings
	aliasTests := []struct {
		alias    string
		expected OutputFormat
	}{
		{"yml", FormatYAML},
		{"md", FormatMarkdown},
		{"plain", FormatNoColor},
		{"minimal", FormatMinimal},
		{"color", FormatColor},
		{"no-color", FormatNoColor},
		{"interactive", FormatHTML},
		{"compact-web", FormatCompactHTML},
		{"nested-md", FormatNestedMarkdown},
		{"table-md", FormatTableMarkdown},
	}

	for _, test := range aliasTests {
		t.Run(test.alias, func(t *testing.T) {
			result, err := ParseFormatString(test.alias)
			if err != nil {
				t.Errorf("Alias %q should be valid, got error: %v", test.alias, err)
			}
			if result != test.expected {
				t.Errorf("Alias %q should map to %q, got %q", test.alias, test.expected, result)
			}
		})
	}
}

func TestFormatValidation(t *testing.T) {
	// Test edge cases in format validation
	edgeCases := []struct {
		input       string
		shouldError bool
		description string
	}{
		{"", true, "empty string"},
		{" ", true, "whitespace only"},
		{"json ", false, "trailing space"},
		{" json", false, "leading space"},
		{" json ", false, "surrounding spaces"},
		{"\tjson\n", false, "tabs and newlines"},
		{"JSON", false, "uppercase"},
		{"Json", false, "mixed case"},
		{"json-", true, "invalid suffix"},
		{"-json", true, "invalid prefix"},
		{"json_format", true, "underscore variant"},
		{"jsonformat", true, "concatenated"},
		{"json format", true, "space in middle"},
		{"json\tjson", true, "multiple formats"},
	}

	for _, test := range edgeCases {
		t.Run(test.description, func(t *testing.T) {
			_, err := ParseFormatString(test.input)

			if test.shouldError && err == nil {
				t.Errorf("Input %q should cause error but didn't", test.input)
			}
			if !test.shouldError && err != nil {
				t.Errorf("Input %q should not cause error but did: %v", test.input, err)
			}
		})
	}
}

func TestOutputFormatString(t *testing.T) {
	// Test that OutputFormat values convert to proper strings
	formats := []OutputFormat{
		FormatJSON,
		FormatYAML,
		FormatHTML,
		FormatMarkdown,
		FormatColor,
		FormatNoColor,
		FormatMinimal,
		FormatCompactJSON,
		FormatFlatJSON,
		FormatNestedMarkdown,
		FormatTableMarkdown,
		FormatCompactHTML,
		FormatTableHTML,
	}

	for _, format := range formats {
		str := string(format)

		// Should not be empty
		if str == "" {
			t.Errorf("Format %v should not convert to empty string", format)
		}

		// Should be parseable back to the same format
		parsed, err := ParseFormatString(str)
		if err != nil {
			t.Errorf("Format string %q should be parseable back, got error: %v", str, err)
		}
		if parsed != format {
			t.Errorf("Format %q should parse back to itself, got %q", format, parsed)
		}
	}
}

func TestFormatRegistry_Completeness(t *testing.T) {
	// Ensure all formats we expect are actually registered
	availableFormats := ListAvailableFormats()

	// Expected formats that should always be available
	expectedFormats := []OutputFormat{
		FormatJSON,
		FormatYAML,
		FormatHTML,
		FormatMarkdown,
		FormatColor,
		FormatNoColor,
		FormatMinimal,
		FormatCompactJSON,
		FormatFlatJSON,
		FormatNestedMarkdown,
		FormatTableMarkdown,
		FormatCompactHTML,
		FormatTableHTML,
	}

	for _, expected := range expectedFormats {
		if _, found := availableFormats[expected]; !found {
			t.Errorf("Expected format %q should be available but wasn't found", expected)
		}

		// Also test that each format can be parsed
		_, err := ParseFormatString(string(expected))
		if err != nil {
			t.Errorf("Format %q should be parseable, got error: %v", expected, err)
		}
	}
}

func TestFormatHelp_Completeness(t *testing.T) {
	help := GetFormatHelp()
	availableFormats := ListAvailableFormats()

	// Help should mention all available formats
	for format := range availableFormats {
		formatStr := string(format)
		if !strings.Contains(help, formatStr) {
			t.Errorf("Help text should mention format %q", formatStr)
		}
	}
}

func TestParseFormatString_ErrorMessages(t *testing.T) {
	// Test that error messages are helpful
	invalidInputs := []string{
		"invalid",
		"jsn",
		"xmll",
		"",
		"json yaml", // Multiple formats
	}

	for _, input := range invalidInputs {
		t.Run(input, func(t *testing.T) {
			_, err := ParseFormatString(input)
			if err == nil {
				t.Errorf("Input %q should produce an error", input)
				return
			}

			errMsg := err.Error()

			// Error message should be helpful
			if errMsg == "" {
				t.Error("Error message should not be empty")
			}

			// Should mention what formats are available (or similar helpful info)
			if !strings.Contains(strings.ToLower(errMsg), "format") {
				t.Errorf("Error message should mention 'format': %s", errMsg)
			}
		})
	}
}
