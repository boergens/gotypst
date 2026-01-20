package foundations

import (
	"math/big"
	"testing"
)

// --- GCD Tests ---

func TestGcd(t *testing.T) {
	tests := []struct {
		name    string
		a       Value
		b       Value
		want    Value
		wantErr bool
	}{
		{"gcd(12, 8)", Int(12), Int(8), Int(4), false},
		{"gcd(48, 18)", Int(48), Int(18), Int(6), false},
		{"gcd(0, 5)", Int(0), Int(5), Int(5), false},
		{"gcd(5, 0)", Int(5), Int(0), Int(5), false},
		{"gcd(0, 0)", Int(0), Int(0), Int(0), false},
		{"gcd(17, 13)", Int(17), Int(13), Int(1), false},
		{"gcd(-12, 8)", Int(-12), Int(8), Int(4), false},
		{"gcd(12, -8)", Int(12), Int(-8), Int(4), false},
		{"gcd(-12, -8)", Int(-12), Int(-8), Int(4), false},
		{"gcd(1, 1)", Int(1), Int(1), Int(1), false},
		{"gcd with float error", Int(12), Float(8.0), nil, true},
		{"gcd with string error", Str("12"), Int(8), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Gcd(tt.a, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("Gcd() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equal(got, tt.want) {
				t.Errorf("Gcd() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- LCM Tests ---

func TestLcm(t *testing.T) {
	tests := []struct {
		name    string
		a       Value
		b       Value
		want    Value
		wantErr bool
	}{
		{"lcm(4, 6)", Int(4), Int(6), Int(12), false},
		{"lcm(12, 18)", Int(12), Int(18), Int(36), false},
		{"lcm(0, 5)", Int(0), Int(5), Int(0), false},
		{"lcm(5, 0)", Int(5), Int(0), Int(0), false},
		{"lcm(1, 1)", Int(1), Int(1), Int(1), false},
		{"lcm(3, 7)", Int(3), Int(7), Int(21), false},
		{"lcm(-4, 6)", Int(-4), Int(6), Int(12), false},
		{"lcm(4, -6)", Int(4), Int(-6), Int(12), false},
		{"lcm with float error", Int(4), Float(6.0), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Lcm(tt.a, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("Lcm() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equal(got, tt.want) {
				t.Errorf("Lcm() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- Factorial Tests ---

func TestFact(t *testing.T) {
	tests := []struct {
		name    string
		n       Value
		want    int64
		wantBig bool
		wantErr bool
	}{
		{"fact(0)", Int(0), 1, false, false},
		{"fact(1)", Int(1), 1, false, false},
		{"fact(5)", Int(5), 120, false, false},
		{"fact(10)", Int(10), 3628800, false, false},
		{"fact(20)", Int(20), 2432902008176640000, false, false},
		{"fact(21)", Int(21), 0, true, false}, // Too big for int64
		{"fact(-1) error", Int(-1), 0, false, true},
		{"fact with float error", Float(5.0), 0, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Fact(tt.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("Fact() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if tt.wantBig {
				if _, ok := got.(*BigInt); !ok {
					t.Errorf("Fact() expected BigInt, got %T", got)
				}
			} else {
				if i, ok := got.(Int); !ok || int64(i) != tt.want {
					t.Errorf("Fact() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestFactLarge(t *testing.T) {
	// Test that fact(100) works and produces a big integer
	got, err := Fact(Int(100))
	if err != nil {
		t.Fatalf("Fact(100) error = %v", err)
	}
	bi, ok := got.(*BigInt)
	if !ok {
		t.Fatalf("Fact(100) expected BigInt, got %T", got)
	}
	// 100! has 158 digits
	str := bi.String()
	if len(str) != 158 {
		t.Errorf("Fact(100) expected 158 digits, got %d", len(str))
	}
}

// --- Permutation Tests ---

func TestPerm(t *testing.T) {
	tests := []struct {
		name    string
		n       Value
		r       Value
		want    int64
		wantBig bool
		wantErr bool
	}{
		{"perm(5, 2)", Int(5), Int(2), 20, false, false},
		{"perm(5, 0)", Int(5), Int(0), 1, false, false},
		{"perm(5, 5)", Int(5), Int(5), 120, false, false},
		{"perm(10, 3)", Int(10), Int(3), 720, false, false},
		{"perm(5, 6) out of range", Int(5), Int(6), 0, false, false},
		{"perm(-1, 2) error", Int(-1), Int(2), 0, false, true},
		{"perm(5, -1) error", Int(5), Int(-1), 0, false, true},
		{"perm with float error", Float(5.0), Int(2), 0, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Perm(tt.n, tt.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Perm() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if tt.wantBig {
				if _, ok := got.(*BigInt); !ok {
					t.Errorf("Perm() expected BigInt, got %T", got)
				}
			} else {
				if i, ok := got.(Int); !ok || int64(i) != tt.want {
					t.Errorf("Perm() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// --- Binomial Coefficient Tests ---

func TestBinom(t *testing.T) {
	tests := []struct {
		name    string
		n       Value
		r       Value
		want    int64
		wantBig bool
		wantErr bool
	}{
		{"binom(5, 2)", Int(5), Int(2), 10, false, false},
		{"binom(5, 0)", Int(5), Int(0), 1, false, false},
		{"binom(5, 5)", Int(5), Int(5), 1, false, false},
		{"binom(10, 3)", Int(10), Int(3), 120, false, false},
		{"binom(20, 10)", Int(20), Int(10), 184756, false, false},
		{"binom(5, 6) out of range", Int(5), Int(6), 0, false, false},
		{"binom(5, -1) out of range", Int(5), Int(-1), 0, false, false},
		{"binom(-1, 2) error", Int(-1), Int(2), 0, false, true},
		{"binom with float error", Float(5.0), Int(2), 0, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Binom(tt.n, tt.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Binom() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if tt.wantBig {
				if _, ok := got.(*BigInt); !ok {
					t.Errorf("Binom() expected BigInt, got %T", got)
				}
			} else {
				if i, ok := got.(Int); !ok || int64(i) != tt.want {
					t.Errorf("Binom() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestBinomPascal(t *testing.T) {
	// Verify Pascal's triangle identity: C(n,r) = C(n-1,r-1) + C(n-1,r)
	for n := int64(2); n <= 10; n++ {
		for r := int64(1); r < n; r++ {
			got, _ := Binom(Int(n), Int(r))
			left, _ := Binom(Int(n-1), Int(r-1))
			right, _ := Binom(Int(n-1), Int(r))
			expected := int64(left.(Int)) + int64(right.(Int))
			if int64(got.(Int)) != expected {
				t.Errorf("Pascal's identity failed: C(%d,%d)=%d, expected %d", n, r, got, expected)
			}
		}
	}
}

// --- Parity Tests ---

func TestEven(t *testing.T) {
	tests := []struct {
		name    string
		n       Value
		want    Value
		wantErr bool
	}{
		{"even(0)", Int(0), Bool(true), false},
		{"even(2)", Int(2), Bool(true), false},
		{"even(4)", Int(4), Bool(true), false},
		{"even(-2)", Int(-2), Bool(true), false},
		{"even(1)", Int(1), Bool(false), false},
		{"even(3)", Int(3), Bool(false), false},
		{"even(-1)", Int(-1), Bool(false), false},
		{"even with float error", Float(2.0), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Even(tt.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("Even() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equal(got, tt.want) {
				t.Errorf("Even() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOdd(t *testing.T) {
	tests := []struct {
		name    string
		n       Value
		want    Value
		wantErr bool
	}{
		{"odd(0)", Int(0), Bool(false), false},
		{"odd(1)", Int(1), Bool(true), false},
		{"odd(3)", Int(3), Bool(true), false},
		{"odd(-1)", Int(-1), Bool(true), false},
		{"odd(2)", Int(2), Bool(false), false},
		{"odd(-2)", Int(-2), Bool(false), false},
		{"odd with float error", Float(1.0), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Odd(tt.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("Odd() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equal(got, tt.want) {
				t.Errorf("Odd() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- Remainder Tests ---

func TestRem(t *testing.T) {
	tests := []struct {
		name    string
		a       Value
		b       Value
		want    Value
		wantErr bool
	}{
		{"rem(7, 3)", Int(7), Int(3), Int(1), false},
		{"rem(-7, 3)", Int(-7), Int(3), Int(-1), false},
		{"rem(7, -3)", Int(7), Int(-3), Int(1), false},
		{"rem(-7, -3)", Int(-7), Int(-3), Int(-1), false},
		{"rem(10, 5)", Int(10), Int(5), Int(0), false},
		{"rem(0, 5)", Int(0), Int(5), Int(0), false},
		{"rem(5, 0) error", Int(5), Int(0), nil, true},
		{"rem with float", Float(7.5), Float(2.5), Float(0.0), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Rem(tt.a, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("Rem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equal(got, tt.want) {
				t.Errorf("Rem() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- DivEuclid Tests ---

func TestDivEuclid(t *testing.T) {
	tests := []struct {
		name    string
		a       Value
		b       Value
		want    Value
		wantErr bool
	}{
		{"div_euclid(7, 3)", Int(7), Int(3), Int(2), false},
		{"div_euclid(-7, 3)", Int(-7), Int(3), Int(-3), false},
		{"div_euclid(7, -3)", Int(7), Int(-3), Int(-2), false},
		{"div_euclid(-7, -3)", Int(-7), Int(-3), Int(3), false},
		{"div_euclid(10, 5)", Int(10), Int(5), Int(2), false},
		{"div_euclid(0, 5)", Int(0), Int(5), Int(0), false},
		{"div_euclid(5, 0) error", Int(5), Int(0), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DivEuclid(tt.a, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("DivEuclid() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equal(got, tt.want) {
				t.Errorf("DivEuclid() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- RemEuclid Tests ---

func TestRemEuclid(t *testing.T) {
	tests := []struct {
		name    string
		a       Value
		b       Value
		want    Value
		wantErr bool
	}{
		{"rem_euclid(7, 3)", Int(7), Int(3), Int(1), false},
		{"rem_euclid(-7, 3)", Int(-7), Int(3), Int(2), false},
		{"rem_euclid(7, -3)", Int(7), Int(-3), Int(1), false},
		{"rem_euclid(-7, -3)", Int(-7), Int(-3), Int(2), false},
		{"rem_euclid(10, 5)", Int(10), Int(5), Int(0), false},
		{"rem_euclid(0, 5)", Int(0), Int(5), Int(0), false},
		{"rem_euclid(5, 0) error", Int(5), Int(0), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RemEuclid(tt.a, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("RemEuclid() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equal(got, tt.want) {
				t.Errorf("RemEuclid() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Verify Euclidean division identity: a = div_euclid(a,b) * b + rem_euclid(a,b)
func TestEuclideanIdentity(t *testing.T) {
	testCases := []struct{ a, b int64 }{
		{7, 3}, {-7, 3}, {7, -3}, {-7, -3},
		{10, 5}, {-10, 5}, {10, -5}, {-10, -5},
		{17, 4}, {-17, 4}, {17, -4}, {-17, -4},
	}

	for _, tc := range testCases {
		a, b := tc.a, tc.b
		q, _ := DivEuclid(Int(a), Int(b))
		r, _ := RemEuclid(Int(a), Int(b))

		// a = q * b + r
		reconstructed := int64(q.(Int))*b + int64(r.(Int))
		if reconstructed != a {
			t.Errorf("Euclidean identity failed: %d = div_euclid(%d,%d)*%d + rem_euclid(%d,%d) = %d*%d + %d = %d",
				a, a, b, b, a, b, q, b, r, reconstructed)
		}

		// rem_euclid should always be non-negative
		if int64(r.(Int)) < 0 {
			t.Errorf("rem_euclid(%d, %d) = %d should be non-negative", a, b, r)
		}
	}
}

// --- Quo Tests ---

func TestQuo(t *testing.T) {
	tests := []struct {
		name    string
		a       Value
		b       Value
		want    Value
		wantErr bool
	}{
		{"quo(7, 3)", Int(7), Int(3), Int(2), false},
		{"quo(-7, 3)", Int(-7), Int(3), Int(-2), false},
		{"quo(7, -3)", Int(7), Int(-3), Int(-2), false},
		{"quo(-7, -3)", Int(-7), Int(-3), Int(2), false},
		{"quo(10, 5)", Int(10), Int(5), Int(2), false},
		{"quo(0, 5)", Int(0), Int(5), Int(0), false},
		{"quo(5, 0) error", Int(5), Int(0), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Quo(tt.a, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("Quo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equal(got, tt.want) {
				t.Errorf("Quo() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- BigInt Tests ---

func TestBigIntString(t *testing.T) {
	bi := &BigInt{Value: big.NewInt(12345)}
	if bi.String() != "12345" {
		t.Errorf("BigInt.String() = %v, want 12345", bi.String())
	}

	if bi.Type() != "int" {
		t.Errorf("BigInt.Type() = %v, want int", bi.Type())
	}
}
