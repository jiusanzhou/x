package opts

import (
	"strings"
	"testing"
)

func TestInteractiveInput(t *testing.T) {
	type Config struct {
		Name string
	}
	c := &Config{}
	n := testNew(c)
	if err := n.load(); err != nil {
		t.Fatal(err)
	}

	input := strings.NewReader("Alice\n")
	output := &strings.Builder{}
	prompter := NewPrompterWithIO(input, output)

	err := n.ParseInteractive(&InteractiveOpts{Prompter: prompter})
	if err != nil {
		t.Fatal(err)
	}

	check(t, c.Name, "Alice")
}

func TestInteractiveInputWithDefault(t *testing.T) {
	type Config struct {
		Name string `opts:"default=Bob"`
	}
	c := &Config{}
	n := testNew(c)
	if err := n.load(); err != nil {
		t.Fatal(err)
	}

	input := strings.NewReader("\n")
	output := &strings.Builder{}
	prompter := NewPrompterWithIO(input, output)

	err := n.ParseInteractive(&InteractiveOpts{Prompter: prompter})
	if err != nil {
		t.Fatal(err)
	}

	check(t, c.Name, "Bob")
}

func TestInteractiveConfirm(t *testing.T) {
	type Config struct {
		Verbose bool
	}
	c := &Config{}
	n := testNew(c)
	if err := n.load(); err != nil {
		t.Fatal(err)
	}

	input := strings.NewReader("y\n")
	output := &strings.Builder{}
	prompter := NewPrompterWithIO(input, output)

	err := n.ParseInteractive(&InteractiveOpts{Prompter: prompter})
	if err != nil {
		t.Fatal(err)
	}

	check(t, c.Verbose, true)
}

func TestInteractiveSelect(t *testing.T) {
	type Config struct {
		Color string `opts:"enum=red|green|blue"`
	}
	c := &Config{}
	n := testNew(c)
	if err := n.load(); err != nil {
		t.Fatal(err)
	}

	input := strings.NewReader("2\n")
	output := &strings.Builder{}
	prompter := NewPrompterWithIO(input, output)

	err := n.ParseInteractive(&InteractiveOpts{Prompter: prompter})
	if err != nil {
		t.Fatal(err)
	}

	check(t, c.Color, "green")
}

func TestInteractiveSkipSet(t *testing.T) {
	type Config struct {
		Name string
		Age  int
	}
	c := &Config{}
	n := testNew(c)

	err := n.parse("--name", "Alice")
	if err != nil {
		t.Fatal(err)
	}

	input := strings.NewReader("25\n")
	output := &strings.Builder{}
	prompter := NewPrompterWithIO(input, output)

	err = n.ParseInteractive(&InteractiveOpts{Prompter: prompter, SkipSet: true})
	if err != nil {
		t.Fatal(err)
	}

	check(t, c.Name, "Alice")
	check(t, c.Age, 25)
}

func TestInteractiveArgs(t *testing.T) {
	type Config struct {
		File string `opts:"mode=arg"`
	}
	c := &Config{}
	n := testNew(c)
	if err := n.load(); err != nil {
		t.Fatal(err)
	}

	input := strings.NewReader("myfile.txt\n")
	output := &strings.Builder{}
	prompter := NewPrompterWithIO(input, output)

	err := n.ParseInteractive(&InteractiveOpts{Prompter: prompter})
	if err != nil {
		t.Fatal(err)
	}

	check(t, c.File, "myfile.txt")
}

func TestPrompterSelect(t *testing.T) {
	input := strings.NewReader("2\n")
	output := &strings.Builder{}
	p := NewPrompterWithIO(input, output)

	idx, err := p.Select("Choose:", []string{"a", "b", "c"}, 0)
	if err != nil {
		t.Fatal(err)
	}
	check(t, idx, 1)
}

func TestPrompterMultiSelect(t *testing.T) {
	input := strings.NewReader("1,3\n")
	output := &strings.Builder{}
	p := NewPrompterWithIO(input, output)

	indices, err := p.MultiSelect("Choose multiple:", []string{"a", "b", "c"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	check(t, len(indices), 2)
	check(t, indices[0], 0)
	check(t, indices[1], 2)
}

func TestPrompterConfirmDefault(t *testing.T) {
	input := strings.NewReader("\n")
	output := &strings.Builder{}
	p := NewPrompterWithIO(input, output)

	result, err := p.Confirm("Continue?", true)
	if err != nil {
		t.Fatal(err)
	}
	check(t, result, true)
}
