package cli

import (
	"reflect"
	"strings"
	"unicode"
)

type Runner interface {
	Run() error
}

type Describer interface {
	Description() string
}

type ShortDescriber interface {
	ShortDescription() string
}

type ExampleProvider interface {
	Example() string
}

func FromStruct(v any) *Command {
	return fromStructValue(reflect.ValueOf(v), "")
}

func fromStructValue(rv reflect.Value, name string) *Command {
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return nil
	}

	rt := rv.Type()

	if name == "" {
		name = toKebabCase(rt.Name())
	}

	opts := []Option{
		Name(name),
		Config(rv.Addr().Interface()),
	}

	if desc, ok := rv.Addr().Interface().(Describer); ok {
		opts = append(opts, Description(desc.Description()))
	}

	if shortDesc, ok := rv.Addr().Interface().(ShortDescriber); ok {
		opts = append(opts, Short(shortDesc.ShortDescription()))
	}

	if example, ok := rv.Addr().Interface().(ExampleProvider); ok {
		opts = append(opts, Example(example.Example()))
	}

	if runner, ok := rv.Addr().Interface().(Runner); ok {
		opts = append(opts, Run(func(cmd *Command, args ...string) {
			if err := runner.Run(); err != nil {
				cmd.PrintErr(err)
			}
		}))
	}

	cmd := New(opts...)

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fieldValue := rv.Field(i)

		if !field.IsExported() {
			continue
		}

		if field.Anonymous && fieldValue.Kind() == reflect.Struct {
			subCmd := fromStructValue(fieldValue, toKebabCase(field.Name))
			if subCmd != nil {
				cmd.Register(subCmd)
			}
		}

		if field.Type.Kind() == reflect.Struct {
			tag := field.Tag.Get("cmd")
			if tag == "" {
				continue
			}

			tagParts := strings.Split(tag, ",")
			cmdName := tagParts[0]
			if cmdName == "" {
				cmdName = toKebabCase(field.Name)
			}

			subCmd := fromStructValue(fieldValue, cmdName)
			if subCmd != nil {
				cmd.Register(subCmd)
			}
		}
	}

	return cmd
}

func toKebabCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteRune('-')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func RegisterStruct(parent *Command, v any) error {
	cmd := FromStruct(v)
	if cmd == nil {
		return nil
	}
	return parent.Register(cmd)
}
