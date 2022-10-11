package decimal

import (
	"fmt"
	"math"
	"strings"
)

const (
	factor int64 = 10000
	count  int   = 4
)

type Decimal int64

func (d Decimal) Original() int64 {
	return int64(d)
}

func (d Decimal) Multiply(d2 Decimal) Decimal {
	return Decimal(int64(d) * int64(d2) / factor)
}

func (d Decimal) Divide(d2 Decimal) Decimal {
	return Decimal(int64(d) * factor / int64(d2))
}

func (d Decimal) Format(state fmt.State, verb rune) {
	i := int64(d)
	isNegative := i < 0
	l := i
	negativeSym := ""
	if isNegative {
		l = i * -1
		negativeSym = "-"
	}
	r := make([]rune, count)
	rm := int64(0)
	for j := 0; j < count; j++ {
		cj := l % 10
		r[count-j-1] = rune(cj + 48)
		l /= 10
		rm += cj
	}

	var err error

	switch verb {
	case 'f':
		if precision, ok := state.Precision(); ok {
			if precision > count {
				precision -= count
				_, err = fmt.Fprintf(state, "%s%d.%s%s", negativeSym, l, string(r), strings.Repeat("0", precision))
			} else if precision == 0 {
				_, err = fmt.Fprintf(state, "%s%d", negativeSym, l)
			} else {
				_, err = fmt.Fprintf(state, "%s%d.%s", negativeSym, l, string(r[0:precision]))
			}
			return
		}
	}

	if rm > 0 {
		_, err = fmt.Fprintf(state, "%s%d.%s", negativeSym, l, string(r))
	} else {
		_, err = fmt.Fprintf(state, "%s%d", negativeSym, l)
	}
	if err != nil {
		return
	}
}

type types interface {
	~int64 | ~int32 | ~int | ~float64 | ~float32
}

func ToDecimal[T types](f T) Decimal {
	return Decimal(math.Round(float64(f * T(factor))))
}
