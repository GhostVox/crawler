package main

import (
	"reflect"
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	var testcases = []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normalize http url",
			input:    "http://example.com",
			expected: "example.com",
		},
		{
			name:     "normalize https url",
			input:    "https://example.com",
			expected: "example.com",
		},
		{
			name:     "normalize relative url",
			input:    "http://example.com/path/to/page",
			expected: "example.com/path/to/page",
		},
		{
			name:     "normalize absolute url",
			input:    "http://example.com/path/to/page",
			expected: "example.com/path/to/page",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actual := NormalizeURL(tc.input)
			if actual != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, actual)
			}
		})
	}
}

func TestParseHtml(t *testing.T) {
	testcases := []struct {
		name      string
		inputURL  string
		inputBody string
		expected  []string
	}{
		{
			name:     "absolute and relative URLs",
			inputURL: "https://blog.boot.dev",
			inputBody: `
<html>
	<body>
		<a href="/path/one">
			<span>Boot.dev</span>
		</a>
		<a href="https://other.com/path/one">
			<span>Boot.dev</span>
		</a>
	</body>
</html>
`,
			expected: []string{"https://blog.boot.dev/path/one", "https://other.com/path/one"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actual, _ := getURLsFromHTML(tc.inputBody, tc.inputURL)
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, actual)
			}
		})
	}
}
