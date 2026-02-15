package codec

import (
	"testing"
)

type testData struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestJSONCodec_MarshalUnmarshal(t *testing.T) {
	c, err := Get("json")
	if err != nil {
		t.Fatalf("Get(json) failed: %v", err)
	}

	if c.Name() != "json" {
		t.Errorf("Name() = %q, want %q", c.Name(), "json")
	}

	if c.ContentType() != "application/json" {
		t.Errorf("ContentType() = %q, want %q", c.ContentType(), "application/json")
	}

	original := &testData{Name: "test", Value: 42}

	data, err := c.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded testData
	if err := c.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Name != original.Name || decoded.Value != original.Value {
		t.Errorf("decoded = %+v, want %+v", decoded, original)
	}
}

func TestMustGet(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustGet panicked unexpectedly: %v", r)
		}
	}()

	c := MustGet("json")
	if c == nil {
		t.Error("MustGet returned nil")
	}
}

func TestMustGet_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustGet should panic for unknown codec")
		}
	}()

	MustGet("unknown-codec")
}

func TestGet_Unknown(t *testing.T) {
	_, err := Get("unknown-codec")
	if err == nil {
		t.Error("Get should return error for unknown codec")
	}
}
