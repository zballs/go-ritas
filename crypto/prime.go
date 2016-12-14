package crypto

import (
	. "github.com/zballs/goRITAS/util"
	"math"
	"math/big"
)

var s1 = []int{1, 13, 17, 29, 37, 41, 49, 53}
var s2 = []int{7, 19, 31, 43}
var s3 = []int{11, 23, 47, 59}

// Generates prime numbers in range [7, upto]

// TODO: optimize // eliminate multiples of prime squares

func SieveOfAtkin(upto int) []int {

	var x, y, n, r, s int
	sqrt := math.Sqrt(float64(upto))
	is_prime := make([]bool, upto)

	for x = 1; float64(x) <= sqrt; x++ {
		for y = x; float64(y) <= sqrt; y++ {
			if y%2 != 0 {
				if n = 4*x*x + y*y; n <= upto {
					r = n % 60
					for _, s = range s1 {
						if r == s {
							is_prime[n] = !is_prime[n]
							break
						}
					}
				}
			} else if x%2 != 0 && y%2 == 0 {
				if n = 3*x*x + y*y; n <= upto {
					r = n % 60
					for _, s = range s2 {
						if r == s {
							is_prime[n] = !is_prime[n]
							break
						}
					}
				}
			}
			if x > y && ((x%2 != 0 && y%2 == 0) || (x%2 == 0 && y%2 != 0)) {
				if n = 3*x*x - y*y; n <= upto {
					r = n % 60
					for _, s = range s3 {
						if r == s {
							is_prime[n] = !is_prime[n]
							break
						}
					}
				}
			}
		}
	}

	var primes []int

	for n, prime := range is_prime {
		if prime {
			primes = append(primes, n)
		}
	}

	return primes
}

// Distributed Primality // Fermat Test
func ComputeValue1(g, N, p, q *big.Int) *big.Int {
	exp, _ := BigInt(0)
	one, _ := BigInt(1)
	AddTos(exp, N, one)
	SubFroms(exp, p, q)
	result := &big.Int{}
	return result.Exp(g, exp, N)
}

func ComputeValue(g, N, p, q *big.Int) *big.Int {
	exp := &big.Int{}
	exp.Add(p, q)
	result := &big.Int{}
	return result.Exp(g, exp, N)
}

func ProductOfPrimes(N, value1 *big.Int, values []*big.Int) bool { //len(values) = k-1
	value, _ := BigInt(1)
	for _, v := range values {
		MulBy(value, v)
	}
	ModOf(value, N)
	return BytesEqual(value1.Bytes(), value.Bytes())
}
