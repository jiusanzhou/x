package memory

import (
	"context"
	"testing"

	"go.zoe.im/x/micro/store"
)

func TestCreateRecord(t *testing.T) {

	ctx := context.Background()

	s := NewStore()

	r := store.NewSimpleRecord("foo", "Value")
	s.Create(ctx, r)

	rr, err := s.Get(ctx, "foo")
	if err != nil {
		t.Fatalf("unexpect error %s", err)
	}

	if r.GetValue() != rr.GetValue() {
		t.Fatalf("except %v, but got %v", r.GetValue(), rr.GetValue())
	}
}