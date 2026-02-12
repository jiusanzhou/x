package opts

import (
	"flag"
	"fmt"
	"reflect"
)

// node presents a command nodes
// TODO: add cmds here
type node struct {
	err error
	item
	args       []*item
	flagGroups []*itemGroup
	flagNames  map[string]bool
	envNames   map[string]bool
	flagsets   []*flag.FlagSet

	loaded   bool
	cmds     map[string]*node
	prompter Prompter

	//pretend these are in the user struct :)
	internalOpts struct {
		Help       bool
		Version    bool
		ConfigPath string
	}
}

func newNode(val reflect.Value) *node {
	n := &node{
		flagNames: map[string]bool{},
		envNames:  map[string]bool{},

		cmds: map[string]*node{},
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

// errorf to be stored until parse-time
func (n *node) errorf(format string, args ...interface{}) error {
	err := fmt.Errorf(format, args...)
	//only store the first error
	if n.err == nil {
		n.err = err
	}
	return err
}
