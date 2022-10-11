package decimal

import (
	"fmt"
	"testing"
)

func TestToDecimal(t *testing.T) {
	t.Run("int64", func(t *testing.T) {
		dec := ToDecimal(int64(123456))
		got := fmt.Sprint(dec)
		want := "123456"
		if got != want {
			t.Errorf("ToDecimal() = %v, want %v", got, want)
		}
	})
	t.Run("int64 15 precision", func(t *testing.T) {
		dec := ToDecimal(int64(123456))
		got := fmt.Sprintf("%.15f", dec)
		want := "123456.000000000000000"
		if got != want {
			t.Errorf("ToDecimal() = %v, want %v", got, want)
		}
	})
	t.Run("float64", func(t *testing.T) {
		dec := ToDecimal(1234.5678)
		got := fmt.Sprint(dec)
		want := "1234.5678"
		if got != want {
			t.Errorf("ToDecimal() = %v, want %v", got, want)
		}
	})
	t.Run("float64 15 precision", func(t *testing.T) {
		dec := ToDecimal(1234.5678)
		got := fmt.Sprintf("%.15f", dec)
		want := "1234.567800000000000"
		if got != want {
			t.Errorf("ToDecimal() = %v, want %v", got, want)
		}
	})
	t.Run("float64 15 precision round", func(t *testing.T) {
		dec := ToDecimal(1234.56789123)
		got := fmt.Sprintf("%.15f", dec)
		want := "1234.567900000000000"
		if got != want {
			t.Errorf("ToDecimal() = %v, want %v", got, want)
		}
	})
	t.Run("float64 15 precision not round", func(t *testing.T) {
		dec := ToDecimal(1234.56781123)
		got := fmt.Sprintf("%.15f", dec)
		want := "1234.567800000000000"
		if got != want {
			t.Errorf("ToDecimal() = %v, want %v", got, want)
		}
	})
	t.Run("float64 format", func(t *testing.T) {
		dec := ToDecimal(1234.56781123)
		got := fmt.Sprintf("%.4f", dec)
		want := "1234.5678"
		if got != want {
			t.Errorf("ToDecimal() = %v, want %v", got, want)
		}
		dec = ToDecimal(1234.56781123)
		got = fmt.Sprintf("%.3f", dec)
		want = "1234.567"
		if got != want {
			t.Errorf("ToDecimal() = %v, want %v", got, want)
		}
		dec = ToDecimal(1234.56781123)
		got = fmt.Sprintf("%.2f", dec)
		want = "1234.56"
		if got != want {
			t.Errorf("ToDecimal() = %v, want %v", got, want)
		}
		dec = ToDecimal(1234.56781123)
		got = fmt.Sprintf("%.1f", dec)
		want = "1234.5"
		if got != want {
			t.Errorf("ToDecimal() = %v, want %v", got, want)
		}
		dec = ToDecimal(1234.56781123)
		got = fmt.Sprintf("%.0f", dec)
		want = "1234"
		if got != want {
			t.Errorf("ToDecimal() = %v, want %v", got, want)
		}
	})
	t.Run("int64 negative", func(t *testing.T) {
		dec := ToDecimal(int64(-123456))
		got := fmt.Sprintf("%.15f", dec)
		want := "-123456.000000000000000"
		if got != want {
			t.Errorf("ToDecimal() = %v, want %v", got, want)
		}
	})
	t.Run("float64 negative", func(t *testing.T) {
		dec := ToDecimal(-1234.56781123)
		got := fmt.Sprintf("%.15f", dec)
		want := "-1234.567800000000000"
		if got != want {
			t.Errorf("ToDecimal() = %v, want %v", got, want)
		}
	})
	t.Run("addition", func(t *testing.T) {
		dec1 := ToDecimal(12.345671)
		dec2 := ToDecimal(12.345671)
		dec3 := ToDecimal(12.345671)
		got := fmt.Sprintf("%.15f", dec1+dec2+dec3)
		want := "37.037100000000000"
		if got != want {
			t.Errorf("ToDecimal() = %v, want %v", got, want)
		}
	})
	t.Run("subtraction", func(t *testing.T) {
		dec1 := ToDecimal(1234.56789)
		dec2 := ToDecimal(123.45678)
		dec3 := ToDecimal(12.34567)
		got := fmt.Sprintf("%.15f", dec1-dec2-dec3)
		want := "1098.765400000000000"
		if got != want {
			t.Errorf("ToDecimal() = %v, want %v", got, want)
		}
	})
	t.Run("multiply", func(t *testing.T) {
		dec1 := ToDecimal(12.345671)
		dec2 := ToDecimal(12.345671)
		dec3 := ToDecimal(12.345671)
		got := fmt.Sprintf("%.15f", dec1.Multiply(dec2).Multiply(dec3))
		want := "1881.685900000000000"
		if got != want {
			t.Errorf("ToDecimal() = %v, want %v", got, want)
		}
	})
	t.Run("divide", func(t *testing.T) {
		dec1 := ToDecimal(12345.67891)
		dec2 := ToDecimal(23.45678)
		dec3 := ToDecimal(9.34567)
		got := fmt.Sprintf("%.15f", dec1.Divide(dec2).Divide(dec3))
		want := "56.316300000000000"
		if got != want {
			t.Errorf("ToDecimal() = %v, want %v", got, want)
		}
	})
}
