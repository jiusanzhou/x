package opts

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func (n *node) load() error {
	//add this node and its fields (recurses if has sub-commands)
	n.loaded = true
	return n.addStructFields(defaultGroup, n.item.val)
}

func (n *node) parse(args ...string) error {
	if err := n.load(); err != nil {
		return err
	}

	// create a new flagset, and link each item
	flagset := flag.NewFlagSet(n.item.name, flag.ContinueOnError)
	flagset.SetOutput(ioutil.Discard)
	for _, item := range n.flags() {
		// add shortnames where possible
		if item.shortName == "" && len(item.name) >= 2 {
			if s := item.name[0:1]; !n.flagNames[s] {
				item.shortName = s
				n.flagNames[s] = true
			}
		}

		flagset.Var(item, item.name, "")
		if sn := item.shortName; sn != "" {
			flagset.Var(item, sn, "")
		}

		k := item.envName
		if item.set() || k == "" {
			continue
		}
		v := os.Getenv(k)
		if v == "" {
			continue
		}
		err := item.Set(v)
		if err != nil {
			return fmt.Errorf("flag '%s' cannot set invalid env var (%s): %s", item.name, k, err)
		}

		// TODO: supported config path
	}

	if err := flagset.Parse(args); err != nil {
		//insert flag errors into help text
		n.err = err
		// n.internalOpts.Help = true
		return err
	}

	return nil
}

func (n *node) addStructFields(group string, sv reflect.Value) error {
	if sv.Kind() != reflect.Struct {
		return n.errorf("opts: %s should be a pointer to a struct (got %s)", sv.Type().Name(), sv.Kind())
	}
	for i := 0; i < sv.NumField(); i++ {
		sf := sv.Type().Field(i)
		val := sv.Field(i)
		if err := n.addStructField(group, sf, val); err != nil {
			return fmt.Errorf("field '%s' %s", sf.Name, err)
		}
	}
	return nil
}

func (n *node) addStructField(group string, sf reflect.StructField, val reflect.Value) error {
	kv := newKV(sf.Tag.Get("opts"))
	help := sf.Tag.Get("help")
	mode := sf.Tag.Get("type") // legacy versions of this package used "type"
	if m := sf.Tag.Get("mode"); m != "" {
		mode = m // allow "mode" to be used directly, undocumented!
	}
	if err := n.addKVField(kv, sf.Name, help, mode, group, val); err != nil {
		return err
	}
	if ks := kv.keys(); len(ks) > 0 {
		return fmt.Errorf("unused opts keys: %s", strings.Join(ks, ", "))
	}
	return nil
}

func (n *node) addKVField(kv *kv, fName, help, mode, group string, val reflect.Value) error {
	//ignore unaddressed/unexported fields
	if !val.CanSet() {
		return nil
	}
	//parse key-values
	//ignore `opts:"-"`
	if _, ok := kv.take("-"); ok {
		return nil
	}
	//get field name and mode
	name, _ := kv.take("name")
	if name == "" {
		//default to struct field name
		name = camel2dash(fName)
		//slice? use singular, usage of
		//Foos []string should be: --foo bar --foo bazz
		if val.Type().Kind() == reflect.Slice {
			name = getSingular(name)
		}
	}
	//new kv mode supercede legacy mode
	if t, ok := kv.take("mode"); ok {
		mode = t
	}
	//default opts mode from go type
	if mode == "" {
		switch val.Type().Kind() {
		case reflect.Struct:
			mode = "embedded"
		default:
			mode = "flag"
		}
	}
	// use the specified group
	if g, ok := kv.take("group"); ok {
		group = g
	}

	// new kv help defs supercede legacy defs
	if h, ok := kv.take("help"); ok {
		help = h
	}

	//from this point, we must have a flag or an arg
	i, err := newItem(val)
	if err != nil {
		return err
	}
	i.mode = mode
	i.name = name
	i.help = help
	// insert either as flag or as argument
	switch mode {
	case "flag":
		//set default text
		if d, ok := kv.take("default"); ok {
			i.defstr = d
		} else if !i.slice {
			v := val.Interface()
			t := val.Type()
			z := reflect.Zero(t)
			zero := reflect.DeepEqual(v, z.Interface())
			if !zero {
				i.defstr = fmt.Sprintf("%v", v)
			}
		}
		if e, ok := kv.take("env"); ok || n.useEnv {
			explicit := true
			if e == "" {
				explicit = false
				e = camel2const(i.name)
			}
			_, set := n.envNames[e]
			if set && explicit {
				return n.errorf("env name '%s' already in use", e)
			}
			if !set {
				n.envNames[e] = true
				i.envName = e
				i.useEnv = true
			}
		}
		//cannot have duplicates
		if n.flagNames[name] {
			return n.errorf("flag '%s' already exists", name)
		}
		//flags can also set short names
		if short, ok := kv.take("short"); ok {
			if len(short) != 1 {
				return n.errorf("short name '%s' on flag '%s' must be a single character", short, name)
			}
			if n.flagNames[short] {
				return n.errorf("short name '%s' on flag '%s' already exists", short, name)
			}
			n.flagNames[short] = true
			i.shortName = short
		}
		//add to this command's flags
		n.flagNames[name] = true
		g := n.flagGroup(group)
		g.flags = append(g.flags, i)
	case "arg":
		// minimum number of items
		if i.slice {
			if m, ok := kv.take("min"); ok {
				min, err := strconv.Atoi(m)
				if err != nil {
					return n.errorf("min not an integer")
				}
				i.min = min
			}
			if m, ok := kv.take("max"); ok {
				max, err := strconv.Atoi(m)
				if err != nil {
					return n.errorf("max not an integer")
				}
				i.max = max
			}
		}
		// validations
		if group != "" {
			return n.errorf("args cannot be placed into a group")
		}
		for _, item := range n.args {
			if item.slice {
				return n.errorf("cannot come after arg list '%s'", item.name)
			}
		}
		// add to this command's arguments
		n.args = append(n.args, i)
	default:
		return fmt.Errorf("invalid opts mode '%s'", mode)
	}
	return nil
}
