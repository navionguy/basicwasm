package decimal

import (
	"testing"
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
	}{
		{Decimal{value: 4512, exp: -2}, Decimal{value: 73, exp: -1}, Decimal{value: 6180822, exp: -6}},
		{Decimal{value: 4512, exp: -2}, Decimal{value: 34, exp: -1}, Decimal{value: 1327059, exp: -5}},
		{Decimal{value: 4512, exp: -2}, Decimal{value: 44, exp: -1}, Decimal{value: 1025455, exp: -5}},
		{Decimal{value: 4512, exp: -2}, Decimal{value: 2, exp: 0}, Decimal{value: 2256, exp: -2}},
		{Decimal{value: 4512, exp: -2}, Decimal{value: -2, exp: 0}, Decimal{value: -2256, exp: -2}},
		{Decimal{value: 451268, exp: -4}, Decimal{value: 34, exp: 5}, Decimal{value: 132, exp: -7}},
		{Decimal{value: -451268, exp: -4}, Decimal{value: 3422, exp: -2}, Decimal{value: -1318725, exp: -6}},
	}

	for _, tt := range tests {
		res := tt.numerator.Div(tt.denomanator)

		if res.Cmp(tt.expect) != 0 {
			t.Errorf("Div expected(v:%d, e:%d) got (v:%d, e:%d)", tt.expect.value, tt.expect.exp, res.value, res.exp)
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
		val int
		res int
	}{
		{5, 5},
		{-4, 4},
		{0, 0},
	}

	for _, tt := range tests {
		res := abs(tt.val)

		if res != tt.res {
			t.Errorf("abs() expected %d got %d", tt.res, res)
		}
	}
}
