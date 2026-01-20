package foundations

import (
	"math/big"
)

// ----------------------------------------------------------------------------
// Number Theory Functions
// ----------------------------------------------------------------------------

// Gcd computes the greatest common divisor of two integers.
// Returns an error if either argument is not an integer.
func Gcd(a, b Value) (Value, error) {
	ai, ok := a.(Int)
	if !ok {
		return nil, &OpError{Message: "gcd requires integer arguments, got " + a.Type()}
	}
	bi, ok := b.(Int)
	if !ok {
		return nil, &OpError{Message: "gcd requires integer arguments, got " + b.Type()}
	}

	result := gcdInt(int64(ai), int64(bi))
	return Int(result), nil
}

// gcdInt computes gcd using the Euclidean algorithm.
// Always returns a non-negative result.
func gcdInt(a, b int64) int64 {
	// Handle negative inputs
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}

	// Euclidean algorithm
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// Lcm computes the least common multiple of two integers.
// Returns an error if either argument is not an integer, or if the result overflows.
func Lcm(a, b Value) (Value, error) {
	ai, ok := a.(Int)
	if !ok {
		return nil, &OpError{Message: "lcm requires integer arguments, got " + a.Type()}
	}
	bi, ok := b.(Int)
	if !ok {
		return nil, &OpError{Message: "lcm requires integer arguments, got " + b.Type()}
	}

	// lcm(0, x) = lcm(x, 0) = 0
	if ai == 0 || bi == 0 {
		return Int(0), nil
	}

	result, ok := lcmInt(int64(ai), int64(bi))
	if !ok {
		return nil, &OpError{Message: "integer overflow in lcm"}
	}
	return Int(result), nil
}

// lcmInt computes lcm using: lcm(a,b) = |a*b| / gcd(a,b)
// Returns (result, ok) where ok is false if overflow occurred.
func lcmInt(a, b int64) (int64, bool) {
	g := gcdInt(a, b)
	if g == 0 {
		return 0, true
	}

	// Divide first to avoid overflow: (a / gcd) * b
	quotient := a / g

	// Check for overflow before multiplication
	if quotient > 0 && b > 0 && quotient > (1<<63-1)/b {
		return 0, false
	}
	if quotient < 0 && b < 0 && quotient < (1<<63-1)/(-b) {
		return 0, false
	}
	if quotient > 0 && b < 0 && -b > (1<<63-1)/quotient {
		return 0, false
	}
	if quotient < 0 && b > 0 && -quotient > (1<<63-1)/b {
		return 0, false
	}

	result := quotient * b
	if result < 0 {
		result = -result
	}
	return result, true
}

// Fact computes the factorial of a non-negative integer.
// Returns a big integer for large results.
// Returns an error if the argument is negative or not an integer.
func Fact(n Value) (Value, error) {
	ni, ok := n.(Int)
	if !ok {
		return nil, &OpError{Message: "fact requires an integer argument, got " + n.Type()}
	}

	if ni < 0 {
		return nil, &OpError{Message: "fact requires a non-negative integer"}
	}

	// For small values, compute directly
	if ni <= 20 {
		result := int64(1)
		for i := int64(2); i <= int64(ni); i++ {
			result *= i
		}
		return Int(result), nil
	}

	// For larger values, use big.Int
	result := factorial(int64(ni))
	// Check if result fits in int64
	if result.IsInt64() {
		return Int(result.Int64()), nil
	}
	// Return as BigInt value
	return &BigInt{Value: result}, nil
}

// factorial computes n! using big.Int for arbitrary precision.
func factorial(n int64) *big.Int {
	if n <= 1 {
		return big.NewInt(1)
	}
	result := big.NewInt(1)
	for i := int64(2); i <= n; i++ {
		result.Mul(result, big.NewInt(i))
	}
	return result
}

// Perm computes the number of permutations: n! / (n-r)!
// Also known as falling factorial or n choose r arrangements.
// Returns an error if arguments are invalid.
func Perm(n, r Value) (Value, error) {
	ni, ok := n.(Int)
	if !ok {
		return nil, &OpError{Message: "perm requires integer arguments, got " + n.Type()}
	}
	ri, ok := r.(Int)
	if !ok {
		return nil, &OpError{Message: "perm requires integer arguments, got " + r.Type()}
	}

	if ni < 0 {
		return nil, &OpError{Message: "perm requires n >= 0"}
	}
	if ri < 0 {
		return nil, &OpError{Message: "perm requires r >= 0"}
	}
	if ri > ni {
		return Int(0), nil
	}

	// For small values, compute directly
	nInt, rInt := int64(ni), int64(ri)
	if nInt <= 20 {
		result := int64(1)
		for i := nInt; i > nInt-rInt; i-- {
			result *= i
		}
		return Int(result), nil
	}

	// For larger values, use big.Int
	result := permutation(nInt, rInt)
	if result.IsInt64() {
		return Int(result.Int64()), nil
	}
	return &BigInt{Value: result}, nil
}

// permutation computes n!/(n-r)! using big.Int.
func permutation(n, r int64) *big.Int {
	result := big.NewInt(1)
	for i := n; i > n-r; i-- {
		result.Mul(result, big.NewInt(i))
	}
	return result
}

// Binom computes the binomial coefficient: n! / (r! * (n-r)!)
// Also known as "n choose r" or combinations.
// Returns an error if arguments are invalid.
func Binom(n, r Value) (Value, error) {
	ni, ok := n.(Int)
	if !ok {
		return nil, &OpError{Message: "binom requires integer arguments, got " + n.Type()}
	}
	ri, ok := r.(Int)
	if !ok {
		return nil, &OpError{Message: "binom requires integer arguments, got " + r.Type()}
	}

	if ni < 0 {
		return nil, &OpError{Message: "binom requires n >= 0"}
	}
	if ri < 0 || ri > ni {
		return Int(0), nil
	}

	nInt, rInt := int64(ni), int64(ri)

	// Use symmetry: C(n,r) = C(n, n-r)
	if rInt > nInt-rInt {
		rInt = nInt - rInt
	}

	// For small values, compute directly using multiplicative formula
	// C(n,r) = n*(n-1)*...*(n-r+1) / r!
	if nInt <= 62 && rInt <= 30 {
		result := int64(1)
		for i := int64(0); i < rInt; i++ {
			result = result * (nInt - i) / (i + 1)
		}
		return Int(result), nil
	}

	// For larger values, use big.Int
	result := binomial(nInt, rInt)
	if result.IsInt64() {
		return Int(result.Int64()), nil
	}
	return &BigInt{Value: result}, nil
}

// binomial computes n choose r using big.Int.
func binomial(n, r int64) *big.Int {
	if r > n-r {
		r = n - r
	}
	if r == 0 {
		return big.NewInt(1)
	}

	// Use multiplicative formula: C(n,r) = n*(n-1)*...*(n-r+1) / r!
	result := big.NewInt(1)
	for i := int64(0); i < r; i++ {
		result.Mul(result, big.NewInt(n-i))
		result.Div(result, big.NewInt(i+1))
	}
	return result
}

// ----------------------------------------------------------------------------
// Parity Functions
// ----------------------------------------------------------------------------

// Even checks if an integer is even.
// Returns an error if the argument is not an integer.
func Even(n Value) (Value, error) {
	ni, ok := n.(Int)
	if !ok {
		return nil, &OpError{Message: "even requires an integer argument, got " + n.Type()}
	}
	return Bool(int64(ni)%2 == 0), nil
}

// Odd checks if an integer is odd.
// Returns an error if the argument is not an integer.
func Odd(n Value) (Value, error) {
	ni, ok := n.(Int)
	if !ok {
		return nil, &OpError{Message: "odd requires an integer argument, got " + n.Type()}
	}
	return Bool(int64(ni)%2 != 0), nil
}

// ----------------------------------------------------------------------------
// Division Functions
// ----------------------------------------------------------------------------

// Rem computes the remainder of division (truncated division).
// The result has the same sign as the dividend.
// For integers: rem(a, b) = a - quo(a, b) * b
// For floats: uses fmod semantics.
func Rem(dividend, divisor Value) (Value, error) {
	switch a := dividend.(type) {
	case Int:
		switch b := divisor.(type) {
		case Int:
			if b == 0 {
				return nil, &OpError{Message: "cannot compute remainder with divisor of zero"}
			}
			return Int(int64(a) % int64(b)), nil
		case Float:
			if b == 0 {
				return nil, &OpError{Message: "cannot compute remainder with divisor of zero"}
			}
			return Float(float64(int64(a)) - float64(int64(float64(int64(a))/float64(b)))*float64(b)), nil
		}
	case Float:
		switch b := divisor.(type) {
		case Int:
			if b == 0 {
				return nil, &OpError{Message: "cannot compute remainder with divisor of zero"}
			}
			fa, fb := float64(a), float64(b)
			return Float(fa - float64(int64(fa/fb))*fb), nil
		case Float:
			if b == 0 {
				return nil, &OpError{Message: "cannot compute remainder with divisor of zero"}
			}
			fa, fb := float64(a), float64(b)
			return Float(fa - float64(int64(fa/fb))*fb), nil
		}
	}
	return nil, &OpError{Message: "rem requires numeric arguments"}
}

// DivEuclid computes the quotient of Euclidean division.
// The result is floor(a/b) for positive divisor, ceil(a/b) for negative divisor.
// This ensures that: a = div_euclid(a,b) * b + rem_euclid(a,b)
// where rem_euclid(a,b) is always in [0, |b|).
func DivEuclid(dividend, divisor Value) (Value, error) {
	switch a := dividend.(type) {
	case Int:
		switch b := divisor.(type) {
		case Int:
			if b == 0 {
				return nil, &OpError{Message: "cannot divide by zero"}
			}
			ai, bi := int64(a), int64(b)
			q := ai / bi
			r := ai % bi
			// Adjust quotient if remainder is negative
			if r < 0 {
				if bi > 0 {
					q--
				} else {
					q++
				}
			}
			return Int(q), nil
		case Float:
			if b == 0 {
				return nil, &OpError{Message: "cannot divide by zero"}
			}
			fa, fb := float64(a), float64(b)
			return Float(divEuclidFloat(fa, fb)), nil
		}
	case Float:
		switch b := divisor.(type) {
		case Int:
			if b == 0 {
				return nil, &OpError{Message: "cannot divide by zero"}
			}
			fa, fb := float64(a), float64(b)
			return Float(divEuclidFloat(fa, fb)), nil
		case Float:
			if b == 0 {
				return nil, &OpError{Message: "cannot divide by zero"}
			}
			return Float(divEuclidFloat(float64(a), float64(b))), nil
		}
	}
	return nil, &OpError{Message: "div_euclid requires numeric arguments"}
}

// divEuclidFloat computes Euclidean division quotient for floats.
func divEuclidFloat(a, b float64) float64 {
	q := a / b
	r := a - float64(int64(q))*b
	if r < 0 {
		if b > 0 {
			return float64(int64(q) - 1)
		}
		return float64(int64(q) + 1)
	}
	return float64(int64(q))
}

// RemEuclid computes the remainder of Euclidean division.
// The result is always non-negative and in [0, |divisor|).
// This ensures that: a = div_euclid(a,b) * b + rem_euclid(a,b)
func RemEuclid(dividend, divisor Value) (Value, error) {
	switch a := dividend.(type) {
	case Int:
		switch b := divisor.(type) {
		case Int:
			if b == 0 {
				return nil, &OpError{Message: "cannot compute remainder with divisor of zero"}
			}
			ai, bi := int64(a), int64(b)
			r := ai % bi
			if r < 0 {
				if bi > 0 {
					r += bi
				} else {
					r -= bi
				}
			}
			return Int(r), nil
		case Float:
			if b == 0 {
				return nil, &OpError{Message: "cannot compute remainder with divisor of zero"}
			}
			fa, fb := float64(a), float64(b)
			return Float(remEuclidFloat(fa, fb)), nil
		}
	case Float:
		switch b := divisor.(type) {
		case Int:
			if b == 0 {
				return nil, &OpError{Message: "cannot compute remainder with divisor of zero"}
			}
			fa, fb := float64(a), float64(b)
			return Float(remEuclidFloat(fa, fb)), nil
		case Float:
			if b == 0 {
				return nil, &OpError{Message: "cannot compute remainder with divisor of zero"}
			}
			return Float(remEuclidFloat(float64(a), float64(b))), nil
		}
	}
	return nil, &OpError{Message: "rem_euclid requires numeric arguments"}
}

// remEuclidFloat computes Euclidean remainder for floats.
func remEuclidFloat(a, b float64) float64 {
	r := a - float64(int64(a/b))*b
	if r < 0 {
		if b > 0 {
			r += b
		} else {
			r -= b
		}
	}
	return r
}

// Quo computes the quotient of truncated division.
// The result is truncated toward zero.
// For integers: quo(a, b) = trunc(a / b)
func Quo(dividend, divisor Value) (Value, error) {
	switch a := dividend.(type) {
	case Int:
		switch b := divisor.(type) {
		case Int:
			if b == 0 {
				return nil, &OpError{Message: "cannot divide by zero"}
			}
			return Int(int64(a) / int64(b)), nil
		case Float:
			if b == 0 {
				return nil, &OpError{Message: "cannot divide by zero"}
			}
			return Float(float64(int64(float64(a) / float64(b)))), nil
		}
	case Float:
		switch b := divisor.(type) {
		case Int:
			if b == 0 {
				return nil, &OpError{Message: "cannot divide by zero"}
			}
			return Float(float64(int64(float64(a) / float64(b)))), nil
		case Float:
			if b == 0 {
				return nil, &OpError{Message: "cannot divide by zero"}
			}
			return Float(float64(int64(float64(a) / float64(b)))), nil
		}
	}
	return nil, &OpError{Message: "quo requires numeric arguments"}
}

// ----------------------------------------------------------------------------
// BigInt Type for Large Factorials
// ----------------------------------------------------------------------------

// BigInt represents an arbitrary-precision integer.
// Used for large factorials that overflow int64.
type BigInt struct {
	Value *big.Int
}

func (*BigInt) valueMarker()  {}
func (*BigInt) Type() string  { return "int" }
func (b *BigInt) String() string {
	if b == nil || b.Value == nil {
		return "0"
	}
	return b.Value.String()
}
