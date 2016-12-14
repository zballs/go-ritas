package crypto

import (
	"github.com/pkg/errors"
	"math/big"
)

// BigInt

func BigInt(value interface{}) (*big.Int, error) {
	switch value.(type) {
	case *big.Int:
		return value.(*big.Int), nil
	case int:
		i64 := int64(value.(int))
		return big.NewInt(i64), nil
	default:
		return nil, errors.New("Unexpected type")
	}
}

func MulBy(x, y *big.Int) {
	x.Mul(x, y)
}

func DivBy(x, y *big.Int) {
	x.Div(x, y)
}

func ExpTo(x, y, m *big.Int) {
	x.Exp(x, y, m)
}

func ModOf(x, y *big.Int) {
	x.Mod(x, y)
}

func AddTo(x, y *big.Int) {
	x.Add(x, y)
}

func SubFrom(x, y *big.Int) {
	x.Sub(x, y)
}

func AddTos(x *big.Int, yz ...*big.Int) {
	for _, y := range yz {
		AddTo(x, y)
	}
}

func SubFroms(x *big.Int, yz ...*big.Int) {
	for _, y := range yz {
		SubFrom(x, y)
	}
}
