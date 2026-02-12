package factory

import (
	"encoding/json"
	"testing"

	"go.zoe.im/x"
)

type TestPlugin interface {
	Name() string
}

type testPlugin struct {
	name string
}

func (p *testPlugin) Name() string {
	return p.name
}

type TestOption struct {
	Prefix string
}

func TestNewFactory(t *testing.T) {
	f := NewFactory[TestPlugin, TestOption]()
	if f == nil {
		t.Fatal("NewFactory() returned nil")
	}
}

func TestFactory_Register(t *testing.T) {
	f := NewFactory[TestPlugin, TestOption]()

	creator := func(cfg x.TypedLazyConfig, opts ...TestOption) (TestPlugin, error) {
		return &testPlugin{name: cfg.Name}, nil
	}

	err := f.Register("test", creator)
	if err != nil {
		t.Errorf("Register() error = %v", err)
	}
}

func TestFactory_Register_Duplicate(t *testing.T) {
	f := NewFactory[TestPlugin, TestOption]()

	creator := func(cfg x.TypedLazyConfig, opts ...TestOption) (TestPlugin, error) {
		return &testPlugin{name: cfg.Name}, nil
	}

	f.Register("test", creator)
	err := f.Register("test", creator)
	if err == nil {
		t.Error("Register() should return error for duplicate registration")
	}
}

func TestFactory_Register_WithAlias(t *testing.T) {
	f := NewFactory[TestPlugin, TestOption]()

	creator := func(cfg x.TypedLazyConfig, opts ...TestOption) (TestPlugin, error) {
		return &testPlugin{name: cfg.Name}, nil
	}

	err := f.Register("test", creator, "alias1", "alias2")
	if err != nil {
		t.Errorf("Register() with aliases error = %v", err)
	}

	cfg := x.TypedLazyConfig{Type: "alias1", Name: "via-alias"}
	plugin, err := f.Create(cfg)
	if err != nil {
		t.Errorf("Create() via alias error = %v", err)
	}
	if plugin.Name() != "via-alias" {
		t.Errorf("Create() via alias Name() = %q, want %q", plugin.Name(), "via-alias")
	}
}

func TestFactory_Create(t *testing.T) {
	f := NewFactory[TestPlugin, TestOption]()

	creator := func(cfg x.TypedLazyConfig, opts ...TestOption) (TestPlugin, error) {
		prefix := ""
		if len(opts) > 0 {
			prefix = opts[0].Prefix
		}
		return &testPlugin{name: prefix + cfg.Name}, nil
	}

	f.Register("test", creator)

	cfg := x.TypedLazyConfig{
		Type:   "test",
		Name:   "myPlugin",
		Config: json.RawMessage(`{}`),
	}

	plugin, err := f.Create(cfg)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if plugin.Name() != "myPlugin" {
		t.Errorf("Create() Name() = %q, want %q", plugin.Name(), "myPlugin")
	}
}

func TestFactory_Create_WithOptions(t *testing.T) {
	f := NewFactory[TestPlugin, TestOption]()

	creator := func(cfg x.TypedLazyConfig, opts ...TestOption) (TestPlugin, error) {
		prefix := ""
		if len(opts) > 0 {
			prefix = opts[0].Prefix
		}
		return &testPlugin{name: prefix + cfg.Name}, nil
	}

	f.Register("test", creator)

	cfg := x.TypedLazyConfig{Type: "test", Name: "plugin"}
	opt := TestOption{Prefix: "prefix_"}

	plugin, err := f.Create(cfg, opt)
	if err != nil {
		t.Fatalf("Create() with options error = %v", err)
	}
	if plugin.Name() != "prefix_plugin" {
		t.Errorf("Create() with options Name() = %q, want %q", plugin.Name(), "prefix_plugin")
	}
}

func TestFactory_Create_UnknownType(t *testing.T) {
	f := NewFactory[TestPlugin, TestOption]()

	cfg := x.TypedLazyConfig{Type: "unknown", Name: "plugin"}

	_, err := f.Create(cfg)
	if err == nil {
		t.Error("Create() should return error for unknown type")
	}
}
