package x

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

var ErrNoMatchingType = errors.New("no matching type found for value")
var ErrTypeNotRegistered = errors.New("type not registered in OneOf registry")

type TypeCreator func() any

type OneOfRegistry struct {
	types     map[string]TypeCreator
	typeField string
}

func NewOneOfRegistry(typeField string) *OneOfRegistry {
	if typeField == "" {
		typeField = "type"
	}
	return &OneOfRegistry{
		types:     make(map[string]TypeCreator),
		typeField: typeField,
	}
}

func (r *OneOfRegistry) Register(typeName string, creator TypeCreator) *OneOfRegistry {
	r.types[typeName] = creator
	return r
}

func (r *OneOfRegistry) RegisterType(typeName string, example any) *OneOfRegistry {
	t := reflect.TypeOf(example)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	finalType := t
	r.types[typeName] = func() any {
		return reflect.New(finalType).Interface()
	}
	return r
}

func (r *OneOfRegistry) Unmarshal(data []byte) (any, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	typeData, ok := raw[r.typeField]
	if !ok {
		return nil, fmt.Errorf("type field '%s' not found in data", r.typeField)
	}

	var typeName string
	if err := json.Unmarshal(typeData, &typeName); err != nil {
		return nil, fmt.Errorf("failed to parse type field: %w", err)
	}

	creator, ok := r.types[typeName]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrTypeNotRegistered, typeName)
	}

	instance := creator()
	if err := json.Unmarshal(data, instance); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to type %s: %w", typeName, err)
	}

	return instance, nil
}

func (r *OneOfRegistry) UnmarshalInto(data []byte, target any) error {
	result, err := r.Unmarshal(data)
	if err != nil {
		return err
	}

	targetVal := reflect.ValueOf(target)
	if targetVal.Kind() != reflect.Ptr {
		return errors.New("target must be a pointer")
	}
	targetVal = targetVal.Elem()

	resultVal := reflect.ValueOf(result)
	if resultVal.Kind() == reflect.Ptr {
		resultVal = resultVal.Elem()
	}

	if !resultVal.Type().AssignableTo(targetVal.Type()) {
		return fmt.Errorf("cannot assign %v to %v", resultVal.Type(), targetVal.Type())
	}

	targetVal.Set(resultVal)
	return nil
}

type OneOf[T any] struct {
	registry *OneOfRegistry
	Value    T
}

func NewOneOf[T any](registry *OneOfRegistry) *OneOf[T] {
	return &OneOf[T]{registry: registry}
}

func (o *OneOf[T]) UnmarshalJSON(data []byte) error {
	if o.registry == nil {
		return errors.New("OneOf registry not initialized")
	}

	result, err := o.registry.Unmarshal(data)
	if err != nil {
		return err
	}

	if typed, ok := result.(T); ok {
		o.Value = typed
		return nil
	}

	if reflect.TypeOf(result).Kind() == reflect.Ptr {
		if typed, ok := reflect.ValueOf(result).Elem().Interface().(T); ok {
			o.Value = typed
			return nil
		}
	}

	return fmt.Errorf("unmarshaled value is not of expected type %T", o.Value)
}

func (o OneOf[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.Value)
}

type OneOfSlice[T any] struct {
	registry *OneOfRegistry
	Values   []T
}

func NewOneOfSlice[T any](registry *OneOfRegistry) *OneOfSlice[T] {
	return &OneOfSlice[T]{registry: registry}
}

func (o *OneOfSlice[T]) UnmarshalJSON(data []byte) error {
	if o.registry == nil {
		return errors.New("OneOfSlice registry not initialized")
	}

	var rawItems []json.RawMessage
	if err := json.Unmarshal(data, &rawItems); err != nil {
		return fmt.Errorf("failed to parse JSON array: %w", err)
	}

	o.Values = make([]T, 0, len(rawItems))
	for i, rawItem := range rawItems {
		result, err := o.registry.Unmarshal(rawItem)
		if err != nil {
			return fmt.Errorf("failed to unmarshal item %d: %w", i, err)
		}

		if typed, ok := result.(T); ok {
			o.Values = append(o.Values, typed)
			continue
		}

		if reflect.TypeOf(result).Kind() == reflect.Ptr {
			if typed, ok := reflect.ValueOf(result).Elem().Interface().(T); ok {
				o.Values = append(o.Values, typed)
				continue
			}
		}

		return fmt.Errorf("item %d is not of expected type", i)
	}

	return nil
}

func (o OneOfSlice[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.Values)
}
