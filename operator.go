package x

import (
	"errors"
	"fmt"
	"strconv"
)

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

func RegisterOperator(ot OperatorType, op OperatorTypeImpl) {
	operatorFactory.Store(ot, op)
}

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

func OperatorImplFn(validateFn func(values []string) error, matchFn func(val string, values []string) bool) OperatorTypeImpl {
	return &emptyOperatorImpl{
		validate: validateFn,
		match:    matchFn,
	}
}

type OperatorType string

func (ot OperatorType) Validate(values []string) error {
	op, ok := operatorFactory.Load(ot)
	if !ok {
		return fmt.Errorf("unknown operator type %s", ot)
	}

	return op.Validate(values)
}

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
			return errors.New("operator In must contians 1 value")
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
			return errors.New("operator NotIn must contians 1 value")
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
			return errors.New("operator Exists can not has values")
		}
		return nil
	}, func(val string, values []string) bool {
		// TODO: this maybe not correct
		return val != "<nil>" && val != ""
	}))

	RegisterOperator(OperatorNotExists, OperatorImplFn(func(values []string) error {
		if len(values) > 0 {
			return errors.New("operator NotExists can not has values")
		}
		return nil
	}, func(val string, values []string) bool {
		// TODO: this maybe not correct
		return val == "<nil>" || val == ""
	}))

	RegisterOperator(OperatorGt, OperatorImplFn(func(values []string) error {
		if len(values) != 1 {
			return errors.New("operator Gt values should only 1")
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
			return errors.New("operator Lt values should only 1")
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
			return errors.New("operator Range values should only 2")
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
