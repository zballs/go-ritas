package crypto

import (
	"crypto/rand"
	"fmt"
	"github.com/pkg/errors"
	"math/big"
)

// BigInt polynomials

type Polynomial struct {
	degree int
	coeffs []*big.Int
	denom  *big.Int
}

type Polynomials []*Polynomial

func ConstructPolynomial(degree int, values ...interface{}) (*Polynomial, error) {
	if degree < 0 {
		return nil, errors.New("Degree cannot be less than 0")
	}
	var err error
	coeffs := make([]*big.Int, degree+1)
	for d, value := range values {
		coeffs[d], err = BigInt(value)
		if err != nil {
			return nil, err
		}
	}
	one, err := BigInt(1)
	if err != nil {
		return nil, err
	}
	return &Polynomial{
		degree: degree,
		coeffs: coeffs,
		denom:  one,
	}, nil
}

func Constant(value interface{}) (*Polynomial, error) {
	return ConstructPolynomial(0, value)
}

func (poly *Polynomial) Denominate(denom *big.Int) {
	for d, _ := range poly.coeffs {
		MulBy(poly.coeffs[d], denom)
	}
	MulBy(poly.denom, denom)
}

func (poly *Polynomial) Equalize(p *Polynomial) {
	if poly.denom == p.denom {
		return
	}
	r, div := &big.Int{}, &big.Int{}
	if r.Mod(poly.denom, p.denom).Int64() == int64(0) {
		div.Div(poly.denom, p.denom)
		p.Denominate(div)
	} else if r.Mod(p.denom, poly.denom).Int64() == int64(0) {
		div.Div(p.denom, poly.denom)
		poly.Denominate(div)
	} else {
		poly.Denominate(p.denom)
		p.Denominate(poly.denom)
	}
}

func (poly *Polynomial) RandomizeCoeffs(max int, skips ...int) error {
FOR_LOOP:
	for d := 0; d <= poly.degree; d++ {
		for _, skip := range skips {
			if d == skip {
				continue FOR_LOOP
			}
		}
		bigint, err := BigInt(max)
		if err != nil {
			return err
		}
		poly.coeffs[d], err = rand.Int(rand.Reader, bigint)
		if err != nil {
			return err
		}
	}
	return nil
}

func (poly *Polynomial) Calculate(x int) (*big.Int, error) {
	result, err := BigInt(0)
	if err != nil {
		return nil, err
	}
	product := &big.Int{}
	for d, coeff := range poly.coeffs {
		bigx, err := pow(x, d)
		if err != nil {
			return nil, err
		}
		product.Mul(coeff, bigx)
		AddTo(result, product)
	}
	DivBy(result, poly.denom)
	return result, nil
}

func (poly *Polynomial) Add(p *Polynomial) {
	poly.Equalize(p)
	if p.degree > poly.degree {
		poly.degree = p.degree
	}
	if len(poly.coeffs) >= len(p.coeffs) {
		for d, _ := range poly.coeffs {
			if d == len(p.coeffs) {
				break
			}
			AddTo(poly.coeffs[d], p.coeffs[d])
		}
	} else {
		for d, _ := range p.coeffs {
			if d == len(poly.coeffs) {
				break
			}
			AddTo(p.coeffs[d], poly.coeffs[d])
		}
		poly.coeffs = p.coeffs
	}
}

func (poly *Polynomial) Multiply(p *Polynomial) {
	poly.degree += p.degree
	coeffs := make([]*big.Int, poly.degree+1)
	for d, _ := range coeffs {
		coeffs[d] = &big.Int{}
	}
	product := &big.Int{}
	for d1, coeff1 := range poly.coeffs {
		for d2, coeff2 := range p.coeffs {
			product.Mul(coeff1, coeff2)
			AddTo(coeffs[d1+d2], product)
		}
	}
	poly.coeffs = coeffs
	MulBy(poly.denom, p.denom)
}

func (poly *Polynomial) String() string {
	var str string
	for d := poly.degree; d >= 0; d-- {
		coeff := poly.coeffs[d]
		if d == 0 {
			str += fmt.Sprintf("%v (d=%v)", coeff, poly.denom)
			break
		}
		switch int(coeff.Int64()) {
		case 0:
			// Do nothing
		case 1:
			str += fmt.Sprintf("x^%d + ", d)
		case -1:
			str += fmt.Sprintf("-x^%d + ", d)
		default:
			str += fmt.Sprintf("%vx^%d + ", coeff, d)
		}
	}
	return str
}

func pow(x, d int) (*big.Int, error) {
	exp, err := BigInt(1)
	if err != nil {
		return nil, err
	}
	bigx, err := BigInt(x)
	if err != nil {
		return nil, err
	}
	for i := 0; i < d; i++ {
		MulBy(exp, bigx)
	}
	return exp, nil
}
