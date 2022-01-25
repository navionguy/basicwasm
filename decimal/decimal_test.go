package decimal

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		val    int
		exp    int
		expect Decimal
	}{
		{3276712, -2, Decimal{value: 3276712, exp: -2}},
		{4512, -2, Decimal{value: 4512, exp: -2}},
	}

	for _, tt := range tests {
		res := New(tt.val, tt.exp)

		if res.Cmp(tt.expect) != 0 {
			t.Errorf("New expected(v:%d, e:%d) got (v:%d, e:%d)", tt.expect.value, tt.expect.exp, res.value, res.exp)
		}
	}
}

func TestNewFromString(t *testing.T) {
	tests := []struct {
		inp    string
		expect Decimal
		fail   bool
	}{
		{"32767.12", Decimal{value: 3276712, exp: -2}, false},
		{"45.12", Decimal{value: 4512, exp: -2}, false},
		{"45.12.34", Decimal{}, true},
		{"bob.12", Decimal{}, true},
	}

	for _, tt := range tests {
		res, fail := NewFromString(tt.inp)

		if (fail != nil) && !tt.fail {
			t.Errorf("NewFromString unexpectedly returned %s", fail.Error())
		}

		if (fail == nil) && tt.fail {
			t.Errorf("NewFromString accepted %s", tt.inp)
		}

		if res.Cmp(tt.expect) != 0 {
			t.Errorf("NewFromString expected(v:%d, e:%d) got (v:%d, e:%d)", tt.expect.value, tt.expect.exp, res.value, res.exp)
		}
	}
}

func TestNewFromInt32(t *testing.T) {
	tests := []struct {
		inp    int32
		expect Decimal
	}{
		{32767, Decimal{value: 32767, exp: 0}},
		{45, Decimal{value: 45, exp: 0}},
	}

	for _, tt := range tests {
		res := NewFromInt32(tt.inp)

		if res.Cmp(tt.expect) != 0 {
			t.Errorf("NewFromInt32 expected(v:%d, e:%d) got (v:%d, e:%d)", tt.expect.value, tt.expect.exp, res.value, res.exp)
		}
	}
}

func TestString_Value_IntPart_Float64(t *testing.T) {
	tests := []struct {
		inp    Decimal
		expect string
		intp   int64
		flt    float64
	}{
		{Decimal{value: 3276712, exp: -2}, "32767.12", 32767, 32767.12},
		{Decimal{value: 32767, exp: 0}, "32767", 32767, 32767},
	}

	for _, tt := range tests {
		res := tt.inp.String()
		rInt := tt.inp.IntPart()
		rFlt, _ := tt.inp.Float64()

		if tt.expect != res {
			t.Errorf("d.String() expected %s got %s", tt.expect, res)
		}

		if rInt != tt.intp {
			t.Errorf("d.IntPart() returned %d expected %d", rInt, tt.intp)
		}

		if rFlt != tt.flt {
			t.Errorf("d.Float64() returned %f expected %f", rFlt, tt.flt)
		}
	}
}

func TestNeg(t *testing.T) {
	tests := []struct {
		inp    Decimal
		expect Decimal
	}{
		{Decimal{value: 3276712, exp: -2}, Decimal{value: -3276712, exp: -2}},
		{Decimal{value: -3276712, exp: -2}, Decimal{value: 3276712, exp: -2}},
	}

	for _, tt := range tests {
		res := tt.inp.Neg()

		if res.Cmp(tt.expect) != 0 {
			t.Errorf("d.Neg() expected(v:%d, e:%d) got (v:%d, e:%d)", tt.expect.value, tt.expect.exp, res.value, res.exp)
		}
	}
}

/*
func TestAdd(t *testing.T) {
	tests := []struct {
		inp: int
	}{
		{inp: 2},
	}

	tests := []struct {
		inp: Decimal
		exp: Decimal
	}{
		{inp: Decimal{value: 45, exp: -1}, exp: Decimal{value: 45, exp -1} },
	}

	for _, tt := range tests {
		res := tt.inp.abs()

		assert.Equalf(t, inp.v)
	}
}*/

func TestAdd(t *testing.T) {
	tests := []struct {
		f1     Decimal
		f2     Decimal
		expect Decimal
	}{
		{Decimal{value: 3276712, exp: -2}, Decimal{value: 9, exp: 0}, Decimal{value: 3277612, exp: -2}},
		{Decimal{value: 4512, exp: -2}, Decimal{value: 2, exp: -1}, Decimal{value: 4532, exp: -2}},
	}

	for _, tt := range tests {
		res := tt.f1.Add(tt.f2)

		if res.Cmp(tt.expect) != 0 {
			t.Errorf("d.Add() expected(v:%d, e:%d) got (v:%d, e:%d)", tt.expect.value, tt.expect.exp, res.value, res.exp)
		}
	}
}

func Test_IsZero(t *testing.T) {
	tests := []struct {
		flt  string
		zero bool
	}{
		{flt: "45", zero: false},
		{flt: "0", zero: true},
	}

	for _, tt := range tests {
		dc, err := NewFromString(tt.flt)

		assert.Nilf(t, err, "NewFromString(%s) returned an error", tt.flt)
		assert.Equalf(t, dc.IsZero(), tt.zero, "IsZero(%s) expected %t, got %t", tt.flt, tt.zero, dc.IsZero())
	}
}

func Test_RescalePair(t *testing.T) {
	tests := []struct {
		flt1 string
		flt2 string
		exp1 string
		exp2 string
	}{
		{flt1: "45.25", flt2: "45.25", exp1: "45.25", exp2: "45.25"},
		{flt1: "45.25", flt2: "45.253", exp1: "45.25", exp2: "45.253"},
	}

	for _, tt := range tests {
		dc1, _ := NewFromString(tt.flt1)
		dc2, _ := NewFromString(tt.flt2)
		exp1, _ := NewFromString(tt.exp1)
		exp2, _ := NewFromString(tt.exp2)

		r1, r2 := rescalePair(dc1, dc2)

		assert.Zerof(t, r1.Cmp(exp1), "%s rescaled to %s", tt.flt1, r1.String())
		assert.Zerof(t, r2.Cmp(exp2), "%s rescaled to %s", tt.flt2, r2.String())
	}
}

func Test_Sign(t *testing.T) {
	tests := []struct {
		val string
		exp int
	}{
		{val: "38", exp: 1},
		{val: "0", exp: 0},
		{val: "-12", exp: -1},
	}

	for _, tt := range tests {
		flt, _ := NewFromString(tt.val)

		assert.Equalf(t, flt.sign(), tt.exp, "%s.sign() expected %d got %d", tt.val, tt.exp, flt.sign())
	}
}

func TestSub(t *testing.T) {
	tests := []struct {
		f1     Decimal
		f2     Decimal
		expect Decimal
	}{
		{Decimal{value: 3276712, exp: -2}, Decimal{value: 9, exp: 0}, Decimal{value: 3275812, exp: -2}},
		{Decimal{value: 4512, exp: -2}, Decimal{value: 2, exp: -3}, Decimal{value: 45118, exp: -3}},
	}

	for _, tt := range tests {
		res := tt.f1.Sub(tt.f2)

		if res.Cmp(tt.expect) != 0 {
			t.Errorf("d.Sub() expected(v:%d, e:%d) got (v:%d, e:%d)", tt.expect.value, tt.expect.exp, res.value, res.exp)
		}
	}
}

func TestMul(t *testing.T) {
	tests := []struct {
		f1     Decimal
		f2     Decimal
		expect Decimal
	}{
		{Decimal{value: 3276712, exp: -2}, Decimal{value: 9, exp: 0}, Decimal{value: 2949041, exp: -1}},
		{Decimal{value: 4512, exp: -2}, Decimal{value: 2, exp: 0}, Decimal{value: 9024, exp: -2}},
	}

	for _, tt := range tests {
		res := tt.f1.Mul(tt.f2)

		if res.Cmp(tt.expect) != 0 {
			t.Errorf("d.Mul() expected(v:%d, e:%d) got (v:%d, e:%d)", tt.expect.value, tt.expect.exp, res.value, res.exp)
		}
	}
}

func TestDiv(t *testing.T) {
	tests := []struct {
		numerator   Decimal
		denomanator Decimal
		expect      Decimal
		err         int
	}{
		{numerator: Decimal{value: 4513, exp: -2}, denomanator: Decimal{value: 0, exp: 0}, err: 11},
		{numerator: Decimal{value: 4512, exp: -2}, denomanator: Decimal{value: 73, exp: math.MinInt32}, err: 6},
		{numerator: Decimal{value: 4512, exp: -2}, denomanator: Decimal{value: 73, exp: -1}, expect: Decimal{value: 6180822, exp: -6}},
		{numerator: Decimal{value: 4512, exp: -2}, denomanator: Decimal{value: 34, exp: -1}, expect: Decimal{value: 1327059, exp: -5}},
		{numerator: Decimal{value: 4512, exp: -2}, denomanator: Decimal{value: 44, exp: -1}, expect: Decimal{value: 1025455, exp: -5}},
		{numerator: Decimal{value: 4512, exp: -2}, denomanator: Decimal{value: 2, exp: 0}, expect: Decimal{value: 2256, exp: -2}},
		{numerator: Decimal{value: 4512, exp: -2}, denomanator: Decimal{value: -2, exp: 0}, expect: Decimal{value: -2256, exp: -2}},
		{numerator: Decimal{value: 451268, exp: -4}, denomanator: Decimal{value: 34, exp: 5}, expect: Decimal{value: 132, exp: -7}},
		{numerator: Decimal{value: -451268, exp: -4}, denomanator: Decimal{value: 3422, exp: -2}, expect: Decimal{value: -1318726, exp: -6}},
	}

	for _, tt := range tests {
		res, err := tt.numerator.Div(tt.denomanator)

		assert.Equal(t, tt.err, err)

		if tt.expect.value != 0 {
			assert.Zerof(t, res.Cmp(tt.expect), "Div expected(v:%d, e:%d) got (v:%d, e:%d)", tt.expect.value, tt.expect.exp, res.value, res.exp)
		}
	}
}

func TestCmp(t *testing.T) {
	tests := []struct {
		d1  Decimal
		d2  Decimal
		exp int
	}{
		{Decimal{value: 4512, exp: -2}, Decimal{value: 4512, exp: -2}, 0},
		{Decimal{value: 4512, exp: -3}, Decimal{value: 4542, exp: -2}, -1},
		{Decimal{value: 4512, exp: -1}, Decimal{value: 451268, exp: -4}, 1},
	}

	for _, tt := range tests {
		res := tt.d1.Cmp(tt.d2)

		if res != tt.exp {
			t.Errorf("Cmp() expected %d, got %d", tt.exp, res)
		}
	}
}

func TestCmp2(t *testing.T) {
	tests := []struct {
		d1  Decimal
		d2  int
		exp int
	}{
		{Decimal{value: 4512, exp: -2}, 45, 0},
		{Decimal{value: 4512, exp: -3}, 45, -1},
		{Decimal{value: 4512, exp: -1}, 45, 1},
		{Decimal{value: 4512, exp: 0}, 45, 1},
	}

	for _, tt := range tests {
		res := tt.d1.cmp(tt.d2)

		if res != tt.exp {
			t.Errorf("cmp() expected %d, got %d", tt.exp, res)
		}
	}
}

func TestCountDigits(t *testing.T) {
	tests := []struct {
		val Decimal
		exp int
	}{
		{Decimal{value: 6180822, exp: -6}, 7},
		{Decimal{value: 6180822, exp: 0}, 7},
	}

	for _, tt := range tests {
		res := tt.val.countDigits()

		if res != tt.exp {
			t.Errorf("countDigits() expected %d, got %d", tt.exp, res)
		}
	}

}

func TestMax(t *testing.T) {
	tests := []struct {
		x   int
		y   int
		res int
	}{
		{5, 6, 6},
		{-4, 4, 4},
		{0, 3, 3},
		{5, 5, 5},
		{7, 5, 7},
	}

	for _, tt := range tests {
		res := max(tt.x, tt.y)

		if res != tt.res {
			t.Errorf("max() expected %d got %d", tt.res, res)
		}
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		x   int
		y   int
		res int
	}{
		{5, 6, 5},
		{-4, 4, -4},
		{0, 3, 0},
		{5, 5, 5},
		{7, 5, 5},
	}

	for _, tt := range tests {
		res := min(tt.x, tt.y)

		if res != tt.res {
			t.Errorf("min() expected %d got %d", tt.res, res)
		}
	}
}

func TestAbs(t *testing.T) {
	tests := []struct {
		val Decimal
		res Decimal
	}{
		{val: Decimal{value: 45, exp: -1}, res: Decimal{value: 45, exp: -1}},
		//{vaL: Decimal{value: 45, exp: -1}, res: Decimal{value: 45, exp: -1}},
		//{-4, 4},
		//{0, 0},
	}

	for _, tt := range tests {
		res := (&tt.val).Abs()

		if res != tt.res {
			t.Errorf("abs() expected %d got %d", tt.res, res)
		}
	}
}

func TestRound(t *testing.T) {
	tests := []struct {
		d1  Decimal
		r   int
		exp Decimal
	}{
		{Decimal{value: 4512, exp: -2}, 0, Decimal{value: 45, exp: 0}},
		{Decimal{value: 4512, exp: -3}, -1, Decimal{value: 45, exp: -1}},
		{Decimal{value: 4512, exp: -1}, 1, Decimal{value: 45, exp: 1}},
		{Decimal{value: 4512, exp: 0}, 1, Decimal{value: 451, exp: 1}},
	}

	for _, tt := range tests {
		res := tt.d1.Round(tt.r)

		if res != tt.exp {
			t.Errorf("cmp() expected %d, got %d", tt.exp, res)
		}
	}

}
