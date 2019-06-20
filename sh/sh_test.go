package sh

import (
	"testing"
)

func TestRun(t *testing.T) {
	if err := Run("echo 111"); err != nil {
		t.Fatalf("%#v", err)
	}
}