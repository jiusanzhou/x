package sh

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func TestRun(t *testing.T) {
	var buf = new(bytes.Buffer)
	fmt.Printf("Start ...")
	if err := Run("echo 111", StdIO(os.Stdin, buf, buf)); err != nil {
		t.Fatalf("%#v", err)
	}
	fmt.Printf("STDOUT: %s\n", buf.Bytes())
}
