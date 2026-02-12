package x

import (
	"encoding/json"
	"testing"
)

type TestAnimal interface {
	Speak() string
}

type TestDog struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Bark string `json:"bark"`
}

func (d TestDog) Speak() string { return d.Bark }

type TestCat struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Meow string `json:"meow"`
}

func (c TestCat) Speak() string { return c.Meow }

func TestOneOfRegistry_Unmarshal(t *testing.T) {
	registry := NewOneOfRegistry("type").
		RegisterType("dog", TestDog{}).
		RegisterType("cat", TestCat{})

	tests := []struct {
		name     string
		json     string
		wantType string
		wantName string
	}{
		{
			name:     "dog",
			json:     `{"type":"dog","name":"Buddy","bark":"woof"}`,
			wantType: "dog",
			wantName: "Buddy",
		},
		{
			name:     "cat",
			json:     `{"type":"cat","name":"Whiskers","meow":"meow"}`,
			wantType: "cat",
			wantName: "Whiskers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := registry.Unmarshal([]byte(tt.json))
			if err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			switch v := result.(type) {
			case *TestDog:
				if v.Type != tt.wantType || v.Name != tt.wantName {
					t.Errorf("got Dog{Type: %q, Name: %q}, want {%q, %q}", v.Type, v.Name, tt.wantType, tt.wantName)
				}
			case *TestCat:
				if v.Type != tt.wantType || v.Name != tt.wantName {
					t.Errorf("got Cat{Type: %q, Name: %q}, want {%q, %q}", v.Type, v.Name, tt.wantType, tt.wantName)
				}
			default:
				t.Errorf("unexpected type %T", result)
			}
		})
	}
}

func TestOneOfRegistry_UnmarshalUnknownType(t *testing.T) {
	registry := NewOneOfRegistry("type").
		RegisterType("dog", TestDog{})

	_, err := registry.Unmarshal([]byte(`{"type":"unknown","name":"test"}`))
	if err == nil {
		t.Error("expected error for unknown type")
	}
}

func TestOneOfRegistry_MissingTypeField(t *testing.T) {
	registry := NewOneOfRegistry("type").
		RegisterType("dog", TestDog{})

	_, err := registry.Unmarshal([]byte(`{"name":"test"}`))
	if err == nil {
		t.Error("expected error for missing type field")
	}
}

func TestOneOfRegistry_CustomTypeField(t *testing.T) {
	registry := NewOneOfRegistry("kind").
		RegisterType("dog", TestDog{})

	data := `{"kind":"dog","name":"Rex"}`
	result, err := registry.Unmarshal([]byte(data))
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	dog, ok := result.(*TestDog)
	if !ok {
		t.Fatalf("expected *TestDog, got %T", result)
	}

	if dog.Name != "Rex" {
		t.Errorf("got name %q, want %q", dog.Name, "Rex")
	}
}

func TestOneOf_UnmarshalJSON(t *testing.T) {
	registry := NewOneOfRegistry("type").
		RegisterType("dog", TestDog{}).
		RegisterType("cat", TestCat{})

	type Config struct {
		Animal *OneOf[TestAnimal] `json:"animal"`
	}

	jsonData := `{"animal":{"type":"dog","name":"Max","bark":"woof!"}}`

	var cfg Config
	cfg.Animal = NewOneOf[TestAnimal](registry)

	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if cfg.Animal.Value == nil {
		t.Fatal("Animal.Value is nil")
	}

	if cfg.Animal.Value.Speak() != "woof!" {
		t.Errorf("Speak() = %q, want %q", cfg.Animal.Value.Speak(), "woof!")
	}
}

func TestOneOfSlice_UnmarshalJSON(t *testing.T) {
	registry := NewOneOfRegistry("type").
		RegisterType("dog", TestDog{}).
		RegisterType("cat", TestCat{})

	type Zoo struct {
		Animals *OneOfSlice[TestAnimal] `json:"animals"`
	}

	jsonData := `{"animals":[{"type":"dog","name":"Max","bark":"woof!"},{"type":"cat","name":"Whiskers","meow":"meow!"}]}`

	var zoo Zoo
	zoo.Animals = NewOneOfSlice[TestAnimal](registry)

	if err := json.Unmarshal([]byte(jsonData), &zoo); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if len(zoo.Animals.Values) != 2 {
		t.Fatalf("expected 2 animals, got %d", len(zoo.Animals.Values))
	}

	sounds := []string{"woof!", "meow!"}
	for i, animal := range zoo.Animals.Values {
		if animal.Speak() != sounds[i] {
			t.Errorf("animal %d: Speak() = %q, want %q", i, animal.Speak(), sounds[i])
		}
	}
}

func TestOneOf_MarshalJSON(t *testing.T) {
	registry := NewOneOfRegistry("type")
	oneof := NewOneOf[TestAnimal](registry)
	oneof.Value = TestDog{Type: "dog", Name: "Rex", Bark: "woof"}

	data, err := json.Marshal(oneof)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	expected := `{"type":"dog","name":"Rex","bark":"woof"}`
	if string(data) != expected {
		t.Errorf("got %s, want %s", string(data), expected)
	}
}
