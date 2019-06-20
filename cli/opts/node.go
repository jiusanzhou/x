package opts

import (
	"flag"
	"fmt"
	"reflect"
)

// node presents a command nodes
// TODO:
type node struct {
	err error
	//embed item since an node can also be an item
	item
	parent     *node
	flagGroups []*itemGroup
	args       []*item
	flagNames  map[string]bool
	flagsets   []*flag.FlagSet
	envNames   map[string]bool
	cmds       map[string]*node
}

func newNode(val reflect.Value) *node {
	n := &node{
		parent:    nil,
		flagNames: map[string]bool{},
		envNames:  map[string]bool{},
		cmds:      map[string]*node{},
	}
	//all new node's MUST be an addressable struct
	t := val.Type()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		val = val.Elem()
	}
	if !val.CanAddr() || t.Kind() != reflect.Struct {
		n.errorf("must be an addressable to a struct")
		return n
	}
	n.item.val = val
	return n
}

//errorf to be stored until parse-time
func (n *node) errorf(format string, args ...interface{}) error {
	err := fmt.Errorf(format, args...)
	//only store the first error
	if n.err == nil {
		n.err = err
	}
	return err
}
