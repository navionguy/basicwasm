package decimal

import (
	"errors"
	"fmt"
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

	if len(parts) == 2 {
		dec.exp, err = strconv.Atoi(parts[1])
	}

	return dec, nil
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
