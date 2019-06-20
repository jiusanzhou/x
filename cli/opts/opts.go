package opts

import (
	"reflect"
)

// Opt is a single option
type Opt interface {
	Name() string
	Short() string
	Help() string
	Default() reflect.Value
	Value() reflect.Value
}

// Opts contains flags, args and commands
type Opts interface {
	Parse() ParsedOpts
	Opts() []Opt
	// TODO:
}

// ParsedOpts ...
type ParsedOpts interface {
	//Help returns the final help text
	Help() string
	//IsRunnable returns whether the matched command has a Run method
	IsRunnable() bool
	//Run assumes the matched command is runnable and executes its Run method.
	//The target Run method must be 'Run() error' or 'Run()'
	Run() error
	//RunFatal assumes the matched command is runnable and executes its Run method.
	//However, any error will be printed, followed by an exit(1).
	RunFatal()
}

// New create a Opts
func New(config interface{}) Opts {
	return newNode(reflect.ValueOf(config))
}

//Setter is any type which can be set from a string.
//This includes flag.Value.
type Setter interface {
	Set(string) error
}
