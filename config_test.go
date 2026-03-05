package x

import (
	"encoding/json"
	"testing"
)

func TestTypedLazyConfig_String(t *testing.T) {
	cfg := TypedLazyConfig{
		Name:   "test",
		Type:   "postgres",
		Config: json.RawMessage(`{"dsn":"localhost"}`),
	}
	s := cfg.String()
	if s != `{test@postgres {"dsn":"localhost"}}` {
		t.Errorf("String() = %q", s)
	}
}

func TestTypedLazyConfig_Unmarshal(t *testing.T) {
	cfg := TypedLazyConfig{
		Type:   "test",
		Config: json.RawMessage(`{"host":"localhost","port":5432}`),
	}

	var target struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}
	if err := cfg.Unmarshal(&target); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if target.Host != "localhost" || target.Port != 5432 {
		t.Errorf("Unmarshal() = %+v", target)
	}
}

func TestTypedLazyConfig_Validate(t *testing.T) {
	type DBConfig struct {
		DSN string `json:"dsn"`
	}

	// valid
	cfg := TypedLazyConfig{
		Type:   "postgres",
		Config: json.RawMessage(`{"dsn":"postgres://localhost"}`),
	}
	if err := cfg.Validate(DBConfig{}); err != nil {
		t.Errorf("Validate() valid config error = %v", err)
	}

	// empty type
	empty := TypedLazyConfig{}
	if err := empty.Validate(DBConfig{}); err == nil {
		t.Error("Validate() should fail when type is empty")
	}

	// no config is OK
	noConfig := TypedLazyConfig{Type: "memory"}
	if err := noConfig.Validate(DBConfig{}); err != nil {
		t.Errorf("Validate() no config should be OK, got %v", err)
	}

	// invalid JSON
	bad := TypedLazyConfig{
		Type:   "test",
		Config: json.RawMessage(`{invalid`),
	}
	if err := bad.Validate(DBConfig{}); err == nil {
		t.Error("Validate() should fail for invalid JSON")
	}
}

func TestTypedLazyConfig_MustUnmarshal(t *testing.T) {
	cfg := TypedLazyConfig{
		Type:   "test",
		Config: json.RawMessage(`{"value":42}`),
	}

	var target struct {
		Value int `json:"value"`
	}
	cfg.MustUnmarshal(&target)
	if target.Value != 42 {
		t.Errorf("MustUnmarshal() Value = %d, want 42", target.Value)
	}
}

func TestTypedLazyConfig_MustUnmarshal_Panics(t *testing.T) {
	cfg := TypedLazyConfig{
		Type:   "test",
		Config: json.RawMessage(`{bad`),
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustUnmarshal() should panic on invalid config")
		}
	}()

	var target struct{}
	cfg.MustUnmarshal(&target)
}
