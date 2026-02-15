package extract

import (
	"testing"

	"go.zoe.im/x/talk"
)

func TestParseAnnotation(t *testing.T) {
	tests := []struct {
		name     string
		comment  string
		expected *Annotation
	}{
		{
			name:    "full annotation",
			comment: "// @talk path=/users/{id} method=GET",
			expected: &Annotation{
				Path:   "/users/{id}",
				Method: "GET",
				Tags:   map[string]string{},
			},
		},
		{
			name:    "path only",
			comment: "// @talk path=/items",
			expected: &Annotation{
				Path: "/items",
				Tags: map[string]string{},
			},
		},
		{
			name:    "with stream mode",
			comment: "// @talk path=/events method=GET stream=sse",
			expected: &Annotation{
				Path:       "/events",
				Method:     "GET",
				StreamMode: talk.StreamServerSide,
				Tags:       map[string]string{},
			},
		},
		{
			name:    "with custom tags",
			comment: "// @talk path=/api method=POST auth=required",
			expected: &Annotation{
				Path:   "/api",
				Method: "POST",
				Tags:   map[string]string{"auth": "required"},
			},
		},
		{
			name:    "bidirectional stream",
			comment: "// @talk path=/chat stream=bidi",
			expected: &Annotation{
				Path:       "/chat",
				StreamMode: talk.StreamBidirect,
				Tags:       map[string]string{},
			},
		},
		{
			name:     "no annotation",
			comment:  "// This is a regular comment",
			expected: nil,
		},
		{
			name:     "empty comment",
			comment:  "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseAnnotation(tt.comment)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Fatal("expected annotation, got nil")
			}

			if result.Path != tt.expected.Path {
				t.Errorf("Path = %q, want %q", result.Path, tt.expected.Path)
			}

			if result.Method != tt.expected.Method {
				t.Errorf("Method = %q, want %q", result.Method, tt.expected.Method)
			}

			if result.StreamMode != tt.expected.StreamMode {
				t.Errorf("StreamMode = %v, want %v", result.StreamMode, tt.expected.StreamMode)
			}

			for key, want := range tt.expected.Tags {
				if got := result.Tags[key]; got != want {
					t.Errorf("Tags[%q] = %q, want %q", key, got, want)
				}
			}
		})
	}
}

func TestParseAnnotations(t *testing.T) {
	comments := []string{
		"// GetUser retrieves a user by ID.",
		"// @talk path=/users/{id} method=GET",
		"// Returns NotFound if user doesn't exist.",
	}

	result := ParseAnnotations(comments)
	if result == nil {
		t.Fatal("expected annotation, got nil")
	}

	if result.Path != "/users/{id}" {
		t.Errorf("Path = %q, want %q", result.Path, "/users/{id}")
	}

	if result.Method != "GET" {
		t.Errorf("Method = %q, want %q", result.Method, "GET")
	}
}

func TestHasAnnotation(t *testing.T) {
	tests := []struct {
		comment  string
		expected bool
	}{
		{"// @talk path=/users", true},
		{"// Regular comment", false},
		{"@talk", true},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.comment, func(t *testing.T) {
			if got := HasAnnotation(tt.comment); got != tt.expected {
				t.Errorf("HasAnnotation(%q) = %v, want %v", tt.comment, got, tt.expected)
			}
		})
	}
}

func TestParseStreamMode(t *testing.T) {
	tests := []struct {
		input    string
		expected talk.StreamMode
	}{
		{"server", talk.StreamServerSide},
		{"server-side", talk.StreamServerSide},
		{"sse", talk.StreamServerSide},
		{"client", talk.StreamClientSide},
		{"client-side", talk.StreamClientSide},
		{"bidi", talk.StreamBidirect},
		{"bidirectional", talk.StreamBidirect},
		{"duplex", talk.StreamBidirect},
		{"none", talk.StreamNone},
		{"", talk.StreamNone},
		{"unknown", talk.StreamNone},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := parseStreamMode(tt.input); got != tt.expected {
				t.Errorf("parseStreamMode(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
