package x

import (
	"encoding/json"
	"time"
)

// Duration is a wrapper around time.Duration which supports correct
// marshaling to YAML and JSON. In particular, it marshals into strings, which
// can be used as map keys in json.
type Duration time.Duration

func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (d *Duration) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}

	pd, err := time.ParseDuration(str)
	if err != nil {
		return err
	}
	*d = Duration(pd)
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// ToUnstructured implements the value.UnstructuredConverter interface.
func (d Duration) ToUnstructured() any {
	return time.Duration(d).String()
}

// MarshalYAML implements the yaml.Marshaler interface.
func (d Duration) MarshalYAML() (any, error) {
	return d.String(), nil
}

// String returns a string representation of the duration.
func (d Duration) String() string {
	return time.Duration(d).String()
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (d *Duration) UnmarshalYAML(unmarshal func(any) error) error {
	var str string
	err := unmarshal(&str)
	if err != nil {
		return err
	}
	return d.DurationFromString(str)
}

// DurationFromString parses a duration from a string.
func (d *Duration) DurationFromString(str string) error {
	duration, err := time.ParseDuration(str)
	if err != nil {
		return err
	}
	*d = Duration(duration)
	return nil
}

// RunWithTimeout executes f with a timeout. Returns true if the timeout was reached.
func RunWithTimeout(f func(exit *bool), timeout time.Duration) bool {
	exit := false
	done := make(chan struct{}, 1)
	// TODO: with gorooutine pool
	go func() {
		f(&exit)
		done <- struct{}{}
	}()
	select {
	case <-done:
		return false
	case <-time.After(timeout):
		exit = true
		return true
	}
}
