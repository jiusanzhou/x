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
func (d Duration) ToUnstructured() interface{} {
	return time.Duration(d).String()
}

// RunWithTimeout execute func with timeout
// return if timeout, TODO: exit shuld be more fance
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
