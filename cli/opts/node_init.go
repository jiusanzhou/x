package opts

func (n *node) Parse() ParsedOpts {
	return nil
}

func (n *node) Opts() []Opt {
	if err := n.load(); err != nil {
		return nil
	}
	opts := []Opt{}
	for _, f := range n.flags() {
		opts = append(opts, f)
	}
	return opts
}

func (n *node) flagGroup(name string) *itemGroup {
	// NOTE: the default group is the empty string
	// get existing group
	for _, g := range n.flagGroups {
		if g.name == name {
			return g
		}
	}
	// otherwise, create and append
	g := &itemGroup{name: name}
	n.flagGroups = append(n.flagGroups, g)
	return g
}

func (n *node) flags() []*item {
	flags := []*item{}
	for _, g := range n.flagGroups {
		flags = append(flags, g.flags...)
	}
	return flags
}
