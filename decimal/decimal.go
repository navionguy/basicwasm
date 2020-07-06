package decimal

// borrowed only the parts I need to from github.com/shopspring/decimal
// the full package from shopspring causes tinygo to woof his cookies

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Decimal represents a fixed-point decimal. It is immutable.
// number = value * 10 ^ exp
type Decimal struct {
	value int

	// NOTE(vadim): this must be an int32, because we cast it to float64 during
	// calculations. If exp is 64 bit, we might lose precision.
	// If we cared about being able to represent every possible decimal, we
	// could make exp a *big.Int but it would hurt performance and numbers
	// like that are unrealistic.
	exp int
}

// NewFromString returns a Decimal object created from the string provided
func NewFromString(src string) (Decimal, error) {
	var dec Decimal
	var err error

	parts := strings.Split(src, ".")

	if len(parts) > 2 {
		return dec, errors.New("invalid decimal")
	}

	dec.value, err = strconv.Atoi(parts[0])

	if err != nil {
		return dec, err
	}

	if len(parts) == 1 {
		return dec, nil
	}

	if (len(parts) == 2) && (len(parts[1]) > 0) {
		dec.exp, err = strconv.Atoi(parts[1])
	}

	return dec, err
}

func (d *Decimal) String() string {
	var rc string

	if d.exp != 0 {
		rc = fmt.Sprintf("%d.%d", d.value, d.exp)
	} else {
		rc = fmt.Sprintf("%d", d.value)
	}

	return rc
}

// Neg returns the negative value from that passed
func (d *Decimal) Neg() Decimal {
	return Decimal{value: -d.value, exp: d.exp}
}

// Add returns the sum of the two values
func (d *Decimal) Add(d2 Decimal) Decimal {
	return Decimal{value: 0}
}

// Sub returns d - d2.
func (d Decimal) Sub(d2 Decimal) Decimal {
	rd, rd2 := rescalePair(d, d2)

	d3Value := rd.value - rd2.value
	return Decimal{
		value: d3Value,
		exp:   rd.exp,
	}
}

// Mul returns the product of the two values
func (d *Decimal) Mul(d2 Decimal) Decimal {
	// todo: figure out how to catch/report overflow
	prod := Decimal{value: d.value * d2.value, exp: d.exp + d2.exp}

	return prod
}

/*
// Div returns d / d2. If it doesn't divide exactly, the result will have
// DivisionPrecision digits after the decimal point.
func (d Decimal) Div(d2 Decimal) Decimal {
	return d.DivRound(d2, int32(DivisionPrecision))
}

// DivRound divides and rounds to a given precision
// i.e. to an integer multiple of 10^(-precision)
//   for a positive quotient digit 5 is rounded up, away from 0
//   if the quotient is negative then digit 5 is rounded down, away from 0
// Note that precision<0 is allowed as input.
func (d Decimal) DivRound(d2 Decimal, precision int32) Decimal {
	// QuoRem already checks initialization
	q, r := d.QuoRem(d2, precision)
	// the actual rounding decision is based on comparing r*10^precision and d2/2
	// instead compare 2 r 10 ^precision and d2
	var rv2 big.Int
	rv2.Abs(r.value)
	rv2.Lsh(&rv2, 1)
	// now rv2 = abs(r.value) * 2
	r2 := Decimal{value: &rv2, exp: r.exp + precision}
	// r2 is now 2 * r * 10 ^ precision
	var c = r2.Cmp(d2.Abs())

	if c < 0 {
		return q
	}

	if d.value.Sign()*d2.value.Sign() < 0 {
		return q.Sub(New(1, -precision))
	}

	return q.Add(New(1, -precision))
}
*/
func (d *Decimal) sign() int {
	if d.value == 0 {
		return 0
	}

	if d.value < 0 {
		return -1
	}

	return 1
}
func (d Decimal) rescale(exp int) Decimal {
	if d.exp == exp {
		return Decimal{d.value, d.exp}
	}

	// NOTE(vadim): must convert exps to float64 before - to prevent overflow
	diff := math.Abs(float64(exp) - float64(d.exp))
	value := d.value

	expScale := math.Pow(10, diff)
	if exp > d.exp {
		value = int(float64(value) / expScale)
	} else if exp < d.exp {
		value = int(float64(value) * expScale)
	}

	return Decimal{
		value: value,
		exp:   exp,
	}
}

func rescalePair(d1, d2 Decimal) (Decimal, Decimal) {
	if d1.exp == d2.exp {
		return d1, d2
	}

	baseScale := min(d1.exp, d2.exp)

	if baseScale != d1.exp {
		return d1.rescale(baseScale), d2
	}

	return d1, d2.rescale(baseScale)
}

func min(x, y int) int {
	if x >= y {
		return y
	}

	return x
}
