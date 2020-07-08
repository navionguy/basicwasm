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

var divisionPrecision = 8

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

// New creaets a new decimal
func New(v int, exp int) Decimal {
	return Decimal{value: v, exp: exp}
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

	if len(parts) == 2 {
		parts[0] = parts[0] + parts[1]
		dec.exp = -len(parts[1])
	}

	dec.value, err = strconv.Atoi(parts[0])

	return dec, err
}

func (d *Decimal) String() string {
	form := "%d"
	if d.exp < 0 {
		form = fmt.Sprintf("%%.%df", abs(d.exp))
	}
	return fmt.Sprintf(form, float64(d.value)*math.Pow(10, float64(d.exp)))
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

// Div returns d / d2. If it doesn't divide exactly, the result will have
// DivisionPrecision digits after the decimal point.
func (d Decimal) Div(d2 Decimal) Decimal {
	return d.DivRound(d2, divisionPrecision)
}

// DivRound divides and rounds to a given precision
// i.e. to an integer multiple of 10^(-precision)
//   for a positive quotient digit 5 is rounded up, away from 0
//   if the quotient is negative then digit 5 is rounded down, away from 0
// Note that precision<0 is allowed as input.
func (d Decimal) DivRound(d2 Decimal, precision int) Decimal {
	// QuoRem already checks initialization
	q, r := d.QuoRem(d2, precision)
	// the actual rounding decision is based on comparing r*10^precision and d2/2
	// instead compare 2 r 10 ^precision and d2

	rv2 := abs(r.value)
	rv2 <<= 1
	// now rv2 = abs(r.value) * 2
	r2 := Decimal{value: rv2, exp: r.exp + precision}
	// r2 is now 2 * r * 10 ^ precision
	var c = r2.Cmp(d2.Abs())

	if c < 0 {
		return q
	}

	if d.sign()*d2.sign() < 0 {
		return q.Sub(New(1, -precision))
	}

	return q.Add(New(1, -precision))
}

// QuoRem does divsion with remainder
// d.QuoRem(d2,precision) returns quotient q and remainder r such that
//   d = d2 * q + r, q an integer multiple of 10^(-precision)
//   0 <= r < abs(d2) * 10 ^(-precision) if d>=0
//   0 >= r > -abs(d2) * 10 ^(-precision) if d<0
// Note that precision<0 is allowed as input.
func (d Decimal) QuoRem(d2 Decimal, precision int) (Decimal, Decimal) {
	if d2.sign() == 0 {
		panic("decimal division by 0")
	}
	scale := -precision
	e := d.exp - d2.exp - int(scale)
	if e > math.MaxInt32 || e < math.MinInt32 {
		panic("overflow in decimal QuoRem")
	}
	var aa, bb, expo int
	var scalerest int
	// d = a 10^ea
	// d2 = b 10^eb
	if e < 0 {
		aa = d.value
		expo = -e
		bb = (10 ^ expo) * d2.value
		scalerest = d.exp
		// now aa = a
		//     bb = b 10^(scale + eb - ea)
	} else {
		expo = e
		aa = 10 ^ expo*d.value
		bb = d2.value
		scalerest = int(scale) + d2.exp
		// now aa = a ^ (ea - eb - scale)
		//     bb = b
	}
	q := aa / bb
	r := aa % bb
	dq := Decimal{value: q, exp: scale}
	dr := Decimal{value: r, exp: scalerest}
	return dq, dr
}

// Cmp compares two decimals and returns
// -1 if d < d2
//  0 if d == d2
//  1 if d > d2
func (d Decimal) Cmp(d2 Decimal) int {
	if d.exp != d2.exp {
		rd, rd2 := rescalePair(d, d2)
		return rd.Cmp(rd2)
	}

	if d.value < d2.value {
		return -1
	}

	if d.value == d2.value {
		return 0
	}

	return 1
}

// Abs calculates and returns the absolute value
func (d *Decimal) Abs() Decimal {
	return Decimal{value: abs(d.value), exp: d.exp}
}

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

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
