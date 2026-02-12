package httputil

import (
	"net/http"
	"testing"
)

func TestCloneRequest(t *testing.T) {
	original, _ := http.NewRequest("GET", "http://example.com/path", nil)
	original.Header.Set("X-Custom", "value")
	original.Header.Add("X-Multi", "v1")
	original.Header.Add("X-Multi", "v2")

	cloned := CloneRequest(original)

	if cloned == original {
		t.Error("CloneRequest should return a new request")
	}

	if cloned.Method != original.Method {
		t.Errorf("CloneRequest Method = %q, want %q", cloned.Method, original.Method)
	}

	if cloned.URL.String() != original.URL.String() {
		t.Errorf("CloneRequest URL = %q, want %q", cloned.URL.String(), original.URL.String())
	}

	if cloned.Header.Get("X-Custom") != "value" {
		t.Errorf("CloneRequest Header X-Custom = %q, want 'value'", cloned.Header.Get("X-Custom"))
	}

	multiValues := cloned.Header.Values("X-Multi")
	if len(multiValues) != 2 {
		t.Errorf("CloneRequest Header X-Multi values = %d, want 2", len(multiValues))
	}
}

func TestCloneRequest_HeaderIndependence(t *testing.T) {
	original, _ := http.NewRequest("GET", "http://example.com", nil)
	original.Header.Set("X-Test", "original")

	cloned := CloneRequest(original)
	cloned.Header.Set("X-Test", "modified")

	if original.Header.Get("X-Test") != "original" {
		t.Error("Modifying cloned header should not affect original")
	}
}

func TestCloneHeader(t *testing.T) {
	original := http.Header{
		"Content-Type": []string{"application/json"},
		"X-Multi":      []string{"a", "b", "c"},
	}

	cloned := CloneHeader(original)

	if len(cloned) != len(original) {
		t.Errorf("CloneHeader len = %d, want %d", len(cloned), len(original))
	}

	if cloned.Get("Content-Type") != "application/json" {
		t.Errorf("CloneHeader Content-Type = %q, want 'application/json'", cloned.Get("Content-Type"))
	}

	multiValues := cloned.Values("X-Multi")
	if len(multiValues) != 3 {
		t.Errorf("CloneHeader X-Multi values = %d, want 3", len(multiValues))
	}
}

func TestCloneHeader_Independence(t *testing.T) {
	original := http.Header{
		"X-Test": []string{"original"},
	}

	cloned := CloneHeader(original)
	cloned.Set("X-Test", "modified")
	cloned.Set("X-New", "added")

	if original.Get("X-Test") != "original" {
		t.Error("Modifying cloned header should not affect original")
	}

	if original.Get("X-New") != "" {
		t.Error("Adding to cloned header should not affect original")
	}
}

func TestCloneHeader_Empty(t *testing.T) {
	original := http.Header{}
	cloned := CloneHeader(original)

	if len(cloned) != 0 {
		t.Errorf("CloneHeader of empty header len = %d, want 0", len(cloned))
	}
}
