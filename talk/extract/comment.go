package extract

import (
	"regexp"
	"strings"

	"go.zoe.im/x/talk"
)

// Annotation represents a parsed @talk annotation from a comment.
type Annotation struct {
	Path       string
	Method     string
	StreamMode talk.StreamMode
	Skip       bool
	Tags       map[string]string
}

var annotationRegex = regexp.MustCompile(`@talk\s*(.*)`)
var kvRegex = regexp.MustCompile(`(\w+)=([^\s]+)`)

// ParseAnnotation extracts @talk annotation from a comment string.
// Format: @talk path=/users/{id} method=GET tag=value
// Use @talk skip or @talk ignore to exclude a method from registration.
func ParseAnnotation(comment string) *Annotation {
	match := annotationRegex.FindStringSubmatch(comment)
	if match == nil {
		return nil
	}

	ann := &Annotation{
		Tags: make(map[string]string),
	}

	content := strings.TrimSpace(match[1])

	if content == "skip" || content == "ignore" || content == "-" {
		ann.Skip = true
		return ann
	}

	kvMatches := kvRegex.FindAllStringSubmatch(content, -1)
	for _, kv := range kvMatches {
		key := strings.ToLower(kv[1])
		value := kv[2]

		switch key {
		case "path":
			ann.Path = value
		case "method":
			ann.Method = strings.ToUpper(value)
		case "stream":
			ann.StreamMode = parseStreamMode(value)
		case "skip", "ignore":
			ann.Skip = value == "true" || value == "1"
		default:
			ann.Tags[key] = value
		}
	}

	return ann
}

// ParseAnnotations extracts all @talk annotations from a multi-line comment.
func ParseAnnotations(comments []string) *Annotation {
	for _, comment := range comments {
		if ann := ParseAnnotation(comment); ann != nil {
			return ann
		}
	}
	return nil
}

func parseStreamMode(s string) talk.StreamMode {
	switch strings.ToLower(s) {
	case "server", "server-side", "sse":
		return talk.StreamServerSide
	case "client", "client-side":
		return talk.StreamClientSide
	case "bidi", "bidirectional", "duplex":
		return talk.StreamBidirect
	default:
		return talk.StreamNone
	}
}

// HasAnnotation checks if the comment contains a @talk annotation.
func HasAnnotation(comment string) bool {
	return strings.Contains(comment, "@talk")
}
