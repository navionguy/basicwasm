package decimal

import (
	"testing"
)

func TestDivRound(t *testing.T) {
	tests := []struct {
		numerator   Decimal
		denomanator Decimal
		precision   int
		expect      Decimal
	}{
		{Decimal{value: 4512, exp: -2}, Decimal{value: 34, exp: -1}, 5, Decimal{value: 1327059, exp: -5}},
		{Decimal{value: 4512, exp: -2}, Decimal{value: 2, exp: 0}, 6, Decimal{value: 2256, exp: -2}},
	}

	for _, tt := range tests {
		res := tt.numerator.DivRound(tt.denomanator, tt.precision)

		if res.Cmp(tt.expect) != 0 {
			t.Errorf("expected(v:%d, e:%d) got (v:%d, e:%d)", tt.expect.value, tt.expect.exp, res.value, res.exp)
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
	}

	for _, tt := range tests {
		res := min(tt.x, tt.y)

		if res != tt.res {
			t.Errorf("expected %d got %d", tt.res, res)
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
			t.Errorf("expected %d got %d", tt.res, res)
		}
	}
}
