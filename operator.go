package x

import (
	"errors"
	"fmt"
	"strconv"
)

// Operator types for label selector matching.
var (
	OperatorIn        = OperatorType("In") // default one
	OperatorNotIn     = OperatorType("NotIn")
	OperatorExists    = OperatorType("Exists")
	OperatorNotExists = OperatorType("NotExists")
	OperatorLt        = OperatorType("Lt")
	OperatorGt        = OperatorType("Gt")
	OperatorRange     = OperatorType("Range")
)

var operatorFactory = SyncMap[OperatorType, OperatorTypeImpl]{}

// RegisterOperator registers a custom operator implementation.
func RegisterOperator(ot OperatorType, op OperatorTypeImpl) {
	operatorFactory.Store(ot, op)
}

// OperatorTypeImpl defines the interface for operator implementations.
type OperatorTypeImpl interface {
	Validate(values []string) error
	Match(val string, values []string) bool
}

type emptyOperatorImpl struct {
	validate func(values []string) error
	match    func(val string, values []string) bool
}

func (e emptyOperatorImpl) Validate(values []string) error {
	return e.validate(values)
}

func (e emptyOperatorImpl) Match(val string, values []string) bool {
	return e.match(val, values)
}

// OperatorImplFn creates an OperatorTypeImpl from validate and match functions.
func OperatorImplFn(validateFn func(values []string) error, matchFn func(val string, values []string) bool) OperatorTypeImpl {
	return &emptyOperatorImpl{
		validate: validateFn,
		match:    matchFn,
	}
}

// OperatorType represents an operator for value matching.
type OperatorType string

// Validate checks if the operator values are valid.
func (ot OperatorType) Validate(values []string) error {
	op, ok := operatorFactory.Load(ot)
	if !ok {
		return fmt.Errorf("unknown operator type %s", ot)
	}

	return op.Validate(values)
}

// Match returns true if val matches the operator condition with values.
func (ot OperatorType) Match(val string, values []string) bool {
	op, ok := operatorFactory.Load(ot)
	if !ok {
		return false
	}

	return op.Match(val, values)
}

func init() {

	RegisterOperator(OperatorIn, OperatorImplFn(func(values []string) error {
		if len(values) == 0 {
			return errors.New("operator In must contain at least 1 value")
		}
		return nil
	}, func(val string, values []string) bool {
		// use glob to match string
		return ContainsFunc(values, func(v string) bool {
			// get the glob of v and then match the val
			return Glob(v).Match(val)
		})
	}))

	RegisterOperator(OperatorNotIn, OperatorImplFn(func(values []string) error {
		if len(values) == 0 {
			return errors.New("operator NotIn must contain at least 1 value")
		}
		return nil
	}, func(val string, values []string) bool {
		// use glob to match string
		return !ContainsFunc(values, func(v string) bool {
			// get the glob of v and then match the val
			return Glob(v).Match(val)
		})
	}))

	RegisterOperator(OperatorExists, OperatorImplFn(func(values []string) error {
		if len(values) > 0 {
			return errors.New("operator Exists cannot have values")
		}
		return nil
	}, func(val string, values []string) bool {
		// TODO: this maybe not correct
		return val != "<nil>" && val != ""
	}))

	RegisterOperator(OperatorNotExists, OperatorImplFn(func(values []string) error {
		if len(values) > 0 {
			return errors.New("operator NotExists cannot have values")
		}
		return nil
	}, func(val string, values []string) bool {
		// TODO: this maybe not correct
		return val == "<nil>" || val == ""
	}))

	RegisterOperator(OperatorGt, OperatorImplFn(func(values []string) error {
		if len(values) != 1 {
			return errors.New("operator Gt requires exactly 1 value")
		}
		// parse string to float64
		_, err := strconv.ParseFloat(values[0], 64)
		return err
	}, func(val string, values []string) bool {
		// parse val to float again
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return false
		}
		t, _ := strconv.ParseFloat(values[0], 64)
		return f >= t
	}))

	RegisterOperator(OperatorLt, OperatorImplFn(func(values []string) error {
		if len(values) != 1 {
			return errors.New("operator Lt requires exactly 1 value")
		}
		// parse string to float64
		_, err := strconv.ParseFloat(values[0], 64)
		return err
	}, func(val string, values []string) bool {
		// parse val to float again
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return false
		}
		t, _ := strconv.ParseFloat(values[0], 64)
		return f <= t
	}))

	RegisterOperator(OperatorRange, OperatorImplFn(func(values []string) error {
		if len(values) != 2 {
			return errors.New("operator Range requires exactly 2 values")
		}
		var err error
		// parse string to float64
		// parse string to float64
		_, err = strconv.ParseFloat(values[0], 64)
		if err != nil {
			return err
		}

		_, err = strconv.ParseFloat(values[1], 64)
		if err != nil {
			return err
		}
		return err
	}, func(val string, values []string) bool {

		var err error
		var l, r, t float64

		// parse string to float64
		l, err = strconv.ParseFloat(values[0], 64)
		if err != nil {
			return false
		}

		// parse string to float64
		r, err = strconv.ParseFloat(values[1], 64)
		if err != nil {
			return false
		}

		t, err = strconv.ParseFloat(val, 64)
		if err != nil {
			return false
		}

		return t >= l && t <= r
	}))
}
