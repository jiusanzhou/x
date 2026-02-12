package opts

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestSimple(t *testing.T) {
	//config
	type Config struct {
		Foo string
		Bar string
		Age int
		Old bool `opts:"short=o"`
	}
	c := &Config{}
	//flag example parse
	err := testNew(c).parse("--foo", "hello", "--bar", "world", "--age", "18", "-o")
	if err != nil {
		t.Fatal(err)
	}
	//check config is filled
	check(t, c.Foo, "hello")
	check(t, c.Bar, "world")
	check(t, c.Age, 18)
	check(t, c.Old, true)
}

func TestIngoreUnknown(t *testing.T) {
	type InnerConfig struct {
		Context context.Context
	}
	//config
	type Config struct {
		Foo        string
		Bar        string
		Age        int
		Old        bool        `opts:"short=o"`
		IngoredOne InnerConfig `opts:"-"`
	}
	c := &Config{}
	//flag example parse
	err := testNew(c).parse("--foo", "hello", "--bar", "world", "--age", "18", "-o")
	if err != nil {
		t.Fatal(err)
	}
	//check config is filled
	check(t, c.Foo, "hello")
	check(t, c.Bar, "world")
	check(t, c.Age, 18)
	check(t, c.Old, true)
}

var spaces = regexp.MustCompile(`\ `)
var newlines = regexp.MustCompile(`\n`)

func readable(s string) string {
	s = spaces.ReplaceAllString(s, "•")
	s = newlines.ReplaceAllString(s, "⏎\n")
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = fmt.Sprintf("%5d: %s", i+1, l)
	}
	s = strings.Join(lines, "\n")
	return s
}

func check(t *testing.T, a, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		stra := readable(fmt.Sprintf("%v", a))
		strb := readable(fmt.Sprintf("%v", b))
		typea := reflect.ValueOf(a)
		typeb := reflect.ValueOf(b)
		extra := ""
		if out, ok := diffstr(stra, strb); ok {
			extra = "\n\n" + out
			stra = "\n" + stra + "\n"
			strb = "\n" + strb + "\n"
		} else {
			stra = "'" + stra + "'"
			strb = "'" + strb + "'"
		}
		t.Fatalf("got %s (%s), expected %s (%s)%s", stra, typea.Kind(), strb, typeb.Kind(), extra)
	}
}

func diffstr(a, b interface{}) (string, bool) {
	stra, oka := a.(string)
	strb, okb := b.(string)
	if !oka || !okb {
		return "", false
	}
	ra := []rune(stra)
	rb := []rune(strb)
	line := 1
	char := 1
	var diff rune
	for i, a := range ra {
		if a == '\n' {
			line++
			char = 1
		} else {
			char++
		}
		var b rune
		if i < len(rb) {
			b = rb[i]
		}
		if a != b {
			a = diff
			break
		}
	}
	return fmt.Sprintf("Diff on line %d char %d (%d)", line, char, diff), true
}

func testNew(config interface{}) *node {
	o := New(config)
	n := o.(*node)
	return n
}

func TestEnumTag(t *testing.T) {
	type Config struct {
		Format string `opts:"enum=list|table|grid"`
	}
	c := &Config{}
	n := testNew(c)

	err := n.parse("--format", "list")
	if err != nil {
		t.Fatal(err)
	}
	check(t, c.Format, "list")
}

func TestEnvTag(t *testing.T) {
	type Config struct {
		Host string `opts:"env=TEST_HOST"`
	}
	c := &Config{}
	_ = testNew(c)
}

func TestNameTag(t *testing.T) {
	type Config struct {
		DatabaseURL string `opts:"name=db-url"`
	}
	c := &Config{}
	n := testNew(c)

	err := n.parse("--db-url", "postgres://localhost")
	if err != nil {
		t.Fatal(err)
	}
	check(t, c.DatabaseURL, "postgres://localhost")
}

func TestNestedStruct(t *testing.T) {
	type Database struct {
		Host string
		Port int
	}
	type Config struct {
		DB Database
	}
	c := &Config{}
	n := testNew(c)

	err := n.parse("--db-host", "localhost", "--db-port", "5432")
	if err != nil {
		t.Fatal(err)
	}
	check(t, c.DB.Host, "localhost")
	check(t, c.DB.Port, 5432)
}

func TestFloatType(t *testing.T) {
	type Config struct {
		Rate float64
	}
	c := &Config{}
	n := testNew(c)

	err := n.parse("--rate", "3.14")
	if err != nil {
		t.Fatal(err)
	}
	if c.Rate != 3.14 {
		t.Errorf("Rate = %f, want 3.14", c.Rate)
	}
}

func TestIntTypes(t *testing.T) {
	type Config struct {
		Int8Val  int8
		Int16Val int16
		Int32Val int32
		Int64Val int64
	}
	c := &Config{}
	n := testNew(c)

	err := n.parse("--int8-val", "8", "--int16-val", "16", "--int32-val", "32", "--int64-val", "64")
	if err != nil {
		t.Fatal(err)
	}
	check(t, c.Int8Val, int8(8))
	check(t, c.Int16Val, int16(16))
	check(t, c.Int32Val, int32(32))
	check(t, c.Int64Val, int64(64))
}

func TestUintTypes(t *testing.T) {
	type Config struct {
		UintVal   uint
		Uint8Val  uint8
		Uint16Val uint16
		Uint32Val uint32
		Uint64Val uint64
	}
	c := &Config{}
	n := testNew(c)

	err := n.parse("--uint-val", "1", "--uint8-val", "8", "--uint16-val", "16", "--uint32-val", "32", "--uint64-val", "64")
	if err != nil {
		t.Fatal(err)
	}
	check(t, c.UintVal, uint(1))
	check(t, c.Uint8Val, uint8(8))
}
