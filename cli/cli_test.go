package cli

import (
	"testing"
)

func TestNew(t *testing.T) {
	cmd := New(Name("testcmd"))
	if cmd == nil {
		t.Fatal("New() returned nil")
	}
	if cmd.Command.Use != "testcmd" {
		t.Errorf("New() Use = %q, want %q", cmd.Command.Use, "testcmd")
	}
}

func TestNew_WithVersion(t *testing.T) {
	cmd := New(Name("testcmd"), Version("1.0.0"))
	if cmd.Command.Version != "1.0.0" {
		t.Errorf("Version() = %q, want %q", cmd.Command.Version, "1.0.0")
	}
}

func TestNew_WithShort(t *testing.T) {
	cmd := New(Name("testcmd"), Short("short description"))
	if cmd.Command.Short != "short description" {
		t.Errorf("Short() = %q, want %q", cmd.Command.Short, "short description")
	}
}

func TestNew_WithLong(t *testing.T) {
	cmd := New(Name("testcmd"), Long("long description"))
	if cmd.Command.Long != "long description" {
		t.Errorf("Long() = %q, want %q", cmd.Command.Long, "long description")
	}
}

func TestNew_WithDescription(t *testing.T) {
	cmd := New(Name("testcmd"), Description("description"))
	if cmd.Command.Long != "description" {
		t.Errorf("Description() = %q, want %q", cmd.Command.Long, "description")
	}
}

func TestNew_WithExample(t *testing.T) {
	cmd := New(Name("testcmd"), Example("example usage"))
	if cmd.Command.Example != "example usage" {
		t.Errorf("Example() = %q, want %q", cmd.Command.Example, "example usage")
	}
}

func TestNew_WithAliases(t *testing.T) {
	cmd := New(Name("testcmd", "alias1", "alias2"))
	if len(cmd.Command.Aliases) != 2 {
		t.Errorf("Aliases len = %d, want 2", len(cmd.Command.Aliases))
	}
}

func TestCommand_Register(t *testing.T) {
	parent := New(Name("parent"))
	child := New(Name("child"))

	err := parent.Register(child)
	if err != nil {
		t.Errorf("Register() error = %v", err)
	}

	if child.parent != parent {
		t.Error("Register() should set child.parent")
	}
}

func TestCommand_IsRoot(t *testing.T) {
	root := New(Name("root"))
	if !root.IsRoot() {
		t.Error("IsRoot() should return true for root command")
	}

	child := New(Name("child"))
	root.Register(child)
	if child.IsRoot() {
		t.Error("IsRoot() should return false for child command")
	}
}

func TestCommand_Option(t *testing.T) {
	cmd := New(Name("testcmd"))
	cmd.Option(Version("2.0.0"))

	if cmd.Command.Version != "2.0.0" {
		t.Errorf("Option() Version = %q, want %q", cmd.Command.Version, "2.0.0")
	}
}

func TestSetFlags(t *testing.T) {
	called := false
	cmd := New(
		Name("testcmd"),
		SetFlags(func(c *Command) {
			called = true
		}),
	)

	if !called {
		t.Error("SetFlags() callback should be called")
	}

	_ = cmd
}

func TestRun_Option(t *testing.T) {
	cmd := New(
		Name("testcmd"),
		Run(func(cmd *Command, args ...string) {
		}),
	)

	if cmd.Command.Run == nil {
		t.Error("Run() should set Command.Run")
	}
}

func TestPreRun(t *testing.T) {
	cmd := New(
		Name("testcmd"),
		PreRun(func(cmd *Command, args ...string) {}),
	)

	if cmd.Command.PreRun == nil {
		t.Error("PreRun() should set Command.PreRun")
	}
}

func TestPersistentPreRun(t *testing.T) {
	cmd := New(
		Name("testcmd"),
		PersistentPreRun(func(cmd *Command, args ...string) {}),
	)

	if cmd.Command.PersistentPreRun == nil {
		t.Error("PersistentPreRun() should set Command.PersistentPreRun")
	}
}
