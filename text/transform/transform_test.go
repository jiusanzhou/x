package transform

import (
	"bytes"
	"io"
	"regexp"
	"strings"
	"testing"

	"golang.org/x/text/transform"
)

func TestRegexLiteral(t *testing.T) {
	re := regexp.MustCompile(`foo`)
	tr := RegexLiteral(re, []byte("bar"))

	result, err := String(tr, "foo baz foo")
	if err != nil {
		t.Fatalf("String() error = %v", err)
	}

	expected := "bar baz bar"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestRegexStringLiteral(t *testing.T) {
	tr, err := RegexStringLiteral(`\d+`, "NUM")
	if err != nil {
		t.Fatalf("RegexStringLiteral() error = %v", err)
	}

	result, err := String(tr, "test123abc456")
	if err != nil {
		t.Fatalf("String() error = %v", err)
	}

	expected := "testNUMabcNUM"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestRegex_WithFunc(t *testing.T) {
	re := regexp.MustCompile(`[a-z]+`)
	tr := Regex(re, func(match []byte) []byte {
		return bytes.ToUpper(match)
	})

	result, err := String(tr, "hello world 123")
	if err != nil {
		t.Fatalf("String() error = %v", err)
	}

	expected := "HELLO WORLD 123"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestChain(t *testing.T) {
	tr1 := RegexLiteral(regexp.MustCompile(`foo`), []byte("bar"))
	tr2 := RegexLiteral(regexp.MustCompile(`bar`), []byte("baz"))

	chain := NewChain(tr1, tr2)
	result, err := chain.String("foo foo foo")
	if err != nil {
		t.Fatalf("String() error = %v", err)
	}

	expected := "baz baz baz"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestChain_Add(t *testing.T) {
	chain := NewChain().
		Add(RegexLiteral(regexp.MustCompile(`a`), []byte("b"))).
		Add(RegexLiteral(regexp.MustCompile(`b`), []byte("c")))

	result, err := chain.String("aaa")
	if err != nil {
		t.Fatalf("String() error = %v", err)
	}

	expected := "ccc"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestChain_Reader(t *testing.T) {
	chain := NewChain(RegexLiteral(regexp.MustCompile(`hello`), []byte("hi")))
	reader := chain.Reader(strings.NewReader("hello world"))

	result, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	expected := "hi world"
	if string(result) != expected {
		t.Errorf("got %q, want %q", string(result), expected)
	}
}

func TestChain_Writer(t *testing.T) {
	var buf bytes.Buffer
	chain := NewChain(RegexLiteral(regexp.MustCompile(`hello`), []byte("hi")))
	writer := chain.Writer(&buf)

	_, err := writer.Write([]byte("hello world"))
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	writer.Close()

	expected := "hi world"
	if buf.String() != expected {
		t.Errorf("got %q, want %q", buf.String(), expected)
	}
}

func TestChain_Empty(t *testing.T) {
	chain := NewChain()
	if chain.Transformer() != transform.Nop {
		t.Error("empty chain should return Nop transformer")
	}
}

func TestChain_Single(t *testing.T) {
	tr := RegexLiteral(regexp.MustCompile(`a`), []byte("b"))
	chain := NewChain(tr)

	if chain.Transformer() != tr {
		t.Error("single-item chain should return the transformer directly")
	}
}

func TestBytes(t *testing.T) {
	tr := RegexLiteral(regexp.MustCompile(`foo`), []byte("bar"))
	result, err := Bytes(tr, []byte("foo"))
	if err != nil {
		t.Fatalf("Bytes() error = %v", err)
	}

	expected := []byte("bar")
	if !bytes.Equal(result, expected) {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestReader(t *testing.T) {
	tr := RegexLiteral(regexp.MustCompile(`test`), []byte("TEST"))
	reader := Reader(strings.NewReader("test123test"), tr)

	result, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	expected := "TEST123TEST"
	if string(result) != expected {
		t.Errorf("got %q, want %q", string(result), expected)
	}
}

func TestWriter(t *testing.T) {
	var buf bytes.Buffer
	tr := RegexLiteral(regexp.MustCompile(`test`), []byte("TEST"))
	writer := Writer(&buf, tr)

	_, err := writer.Write([]byte("test123test"))
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	writer.Close()

	expected := "TEST123TEST"
	if buf.String() != expected {
		t.Errorf("got %q, want %q", buf.String(), expected)
	}
}
