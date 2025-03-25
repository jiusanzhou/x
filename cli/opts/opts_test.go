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
