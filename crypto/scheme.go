package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"math/big"
)

// Sent from server i -> j
type Triple struct {
	i int
	j int
	p *big.Int
	q *big.Int
	h *big.Int
}

func (t *Triple) String() string {
	return fmt.Sprintf("[%d] -> [%d] (p=%v, q=%v, h=%v)\n", t.i, t.j, t.p, t.q, t.h)
}

type Triples []*Triple

// Broadcasted from server i
type Share struct {
	i int
	n *big.Int
}

func (share *Share) String() string {
	return fmt.Sprintf("[%d] (n=%v)\n", share.i, share.n)
}

type Shares []*Share

func NewShare(i int, n *big.Int) *Share {
	if n == nil {
		n = &big.Int{}
	}
	return &Share{i, n}
}

func PickCandidates(bits int) (*big.Int, *big.Int, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	p, q := privKey.Primes[0], privKey.Primes[1]
	return p, q, nil
}

func PickPolynomials(p, q *big.Int, k, max int) (Polynomials, error) {
	l := (k - 1) / 2
	f, err := ConstructPolynomial(l, p)
	if err != nil {
		return nil, err
	}
	err = f.RandomizeCoeffs(max, 0) //skip constant
	if err != nil {
		return nil, err
	}
	g, err := ConstructPolynomial(l, q)
	if err != nil {
		return nil, err
	}
	err = g.RandomizeCoeffs(max, 0)
	if err != nil {
		return nil, err
	}
	h, err := ConstructPolynomial(2*l, 0)
	if err != nil {
		return nil, err
	}
	err = h.RandomizeCoeffs(max, 0)
	if err != nil {
		return nil, err
	}
	return Polynomials{f, g, h}, nil
}

func ComputeTriple(i, j int, f, g, h *Polynomial) (*Triple, error) {
	t := &Triple{}
	t.i, t.j = i, j
	var err error
	t.p, err = f.Calculate(j)
	if err != nil {
		return nil, err
	}
	t.q, err = g.Calculate(j)
	if err != nil {
		return nil, err
	}
	t.h, err = h.Calculate(j)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func ComputeShare(i int, tz Triples) *Share { //len(ts) >= k
	share := NewShare(i, nil)
	pz, _ := BigInt(0)
	qz, _ := BigInt(0)
	hz, _ := BigInt(0)
	for _, t := range tz {
		AddTo(pz, t.p)
		AddTo(qz, t.q)
		AddTo(hz, t.h)
	}
	share.n.Mul(pz, qz)
	AddTo(share.n, hz)
	return share
}

func DiscoverN(shares Shares) (*big.Int, error) { //len(shares) >= k
	poly, err := Constant(0)
	if err != nil {
		return nil, err
	}
	for _, share := range shares {
		p1, err := Interpolate(share, shares)
		if err != nil {
			return nil, err
		}
		p2, err := Constant(share.n)
		if err != nil {
			return nil, err
		}
		p1.Multiply(p2)
		poly.Add(p1)
	}
	return poly.Calculate(0)
}

func Interpolate(share *Share, shares Shares) (*Polynomial, error) {
	poly, err := Constant(1)
	if err != nil {
		return nil, err
	}
	for _, sh := range shares {
		if share.i == sh.i {
			continue
		}
		p, err := ConstructPolynomial(1, -sh.i, 1)
		if err != nil {
			return nil, err
		}
		poly.Multiply(p)
		diff, err := BigInt(share.i - sh.i)
		if err != nil {
			return nil, err
		}
		MulBy(poly.denom, diff)
	}
	return poly, nil
}
