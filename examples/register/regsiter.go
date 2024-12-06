package register

import (
	"go.zoe.im/x"
	"go.zoe.im/x/factory"
)

type Calc interface {
	Add(n1, n2 int) int
}

var (
	myFactory = factory.NewFactory[Calc, int]()
)

func init() {
	myFactory.Register("calc", func(cfg x.TypedLazyConfig, opts ...int) (Calc, error) {
		return nil, nil
	})
}
