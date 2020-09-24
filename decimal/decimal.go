package decimal

// borrowed only the parts I need to from shopspring.decimal
// the full package from shopspring uses the big libary
// which causes tinygo to woof his cookies
// I then had to tweak some of the logic to make it behave like the
// gwbasic interpreter.

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// DivisionPrecision how precise is my work
var DivisionPrecision = 8

// Zero value for checking
var Zero = Decimal{value: 0, exp: 1}

// Decimal represents a fixed-point decimal.
// number = value * 10 ^ exp
type Decimal struct {
	value int
	exp   int
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
		dec.value = 0
	}

	if len(parts) == 2 {
		parts[0] = parts[0] + parts[1]
		dec.exp = -len(parts[1])
	}

	dec.value, err = strconv.Atoi(parts[0])

	return dec, err
}

// NewFromInt32 converts a int32 to Decimal.
//
// Example:
//
//     NewFromInt(123).String() // output: "123"
//     NewFromInt(-10).String() // output: "-10"
func NewFromInt32(value int32) Decimal {
	return Decimal{
		value: int(value),
		exp:   0,
	}
}

func (d *Decimal) String() string {
	if d.exp == 0 {
		return fmt.Sprintf("%d", d.value)
	}

	form := "%f"

	if d.exp < 0 {
		form = fmt.Sprintf("%%.%df", abs(d.exp))
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf(form, float64(d.value)*math.Pow(10, float64(d.exp))), "0"), ".")
}

// Neg returns the negative value from that passed
func (d *Decimal) Neg() Decimal {
	return Decimal{value: -d.value, exp: d.exp}
}

// Add returns the sum of the two values
func (d Decimal) Add(d2 Decimal) Decimal {
	rd, rd2 := rescalePair(d, d2)

	//d3Value := new(big.Int).Add(rd.value, rd2.value)
	return Decimal{
		value: rd.value + rd2.value,
		exp:   rd.exp,
	}
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

	dig := prod.countDigits()

	if dig > 7 {
		return prod.Round(-6)
	}

	return prod
}

// Div divides the two numbers and then does "GWBasic Rounding"
// Which means, the final answer will have seven digits.
// ex: 14.52 / 3.4 = 13.27059
//     14.52 / 7.3 = 6.180822
//
// who thought that up?
func (d Decimal) Div(d2 Decimal) Decimal {
	precision := 7

	q, r := d.QuoRem(d2, precision)

	if r.value == 0 {
		return q.Round(-precision)
	}

	r = r.Abs()
	r.value *= 2
	r.exp += precision + q.exp

	var c = r.Cmp(d2.Abs())

	if c < 0 {
		return q.Round(-precision)
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
	var aa, bb int
	var scalerest int
	// d = a 10^ea
	// d2 = b 10^eb
	if e < 0 {
		aa = d.value
		bb = int(math.Pow10(-e)) * d2.value
		scalerest = d.exp
		// now aa = a
		//     bb = b 10^(scale + eb - ea)
	} else {
		aa = int(math.Pow10(e))
		aa *= d.value
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

// IntPart returns the integer component of the decimal.
func (d Decimal) IntPart() int64 {
	scaledD := d.rescale(0)
	return int64(scaledD.value)
}

// Float64 returns the nearest float64 value for d and a bool indicating
// whether f represents d exactly.
// In this implementation, it is exact
func (d Decimal) Float64() (f float64, exact bool) {
	return float64(d.value) * math.Pow10(d.exp), true
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

// Round rounds the decimal to places decimal places.
// If places < 0, it will round the integer part to the nearest 10^(-places).
//
// Example:
//
// 	   NewFromString("5.45").Round(-1).String() // output: "5.5"
// 	   NewFromString("545").Round(1).String() // output: "550"
//
func (d Decimal) Round(places int) Decimal {
	rc := Decimal{value: d.value, exp: places}
	dc := d.countDigits()
	ms := -d.exp
	rc.exp = max(-(7 - (dc + d.exp)), places)
	rExp := -(ms + rc.exp - 1)
	/*if dc > abs(places) {
		rc.exp += dc + places
		rExp = 1 - (dc + places)
		places = rc.exp
	}*/
	exp := math.Pow10(rExp)
	rndVal := float64(d.value) * exp
	if rndVal >= 0 {
		rndVal += 5
	} else {
		rndVal -= 5
	}
	rc.value = int(rndVal / 10)

	return rc

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

func (d Decimal) cmp(t int) int {
	val := d.value
	if d.exp != 0 {
		val = int(float64(val) * math.Pow10(d.exp))
	}

	if val < t {
		return -1
	}

	if val > t {
		return 1
	}

	return 0
}

func (d Decimal) countDigits() int {
	if d.exp >= 0 {
		return d.exp + countDigits(d.value)
	}

	return countDigitsTrimmed(d.value)
}

func countDigits(num int) int {
	ct := 0
	for num != 0 {
		num /= 10
		ct++
	}
	return ct
}

func countDigitsTrimmed(num int) int {
	ct := 0
	//tz := true

	for num != 0 {
		if num%10 != 0 {
			//tz = false
		}
		num /= 10
		//if !tz { // until I see non-zero, just trailing zeroes
		ct++
		//}
	}
	return ct
}

func max(x, y int) int {
	if x >= y {
		return x
	}

	return y
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
