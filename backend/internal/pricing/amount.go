package pricing

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func ParseMajorAmount(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, ErrInvalidInput
	}
	parts := strings.Split(value, ".")
	if len(parts) > 2 || parts[0] == "" {
		return 0, ErrInvalidInput
	}
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || whole < 0 {
		return 0, ErrInvalidInput
	}
	fraction := int64(0)
	if len(parts) == 2 {
		if len(parts[1]) == 0 || len(parts[1]) > 2 {
			return 0, ErrInvalidInput
		}
		padded := parts[1] + strings.Repeat("0", 2-len(parts[1]))
		fraction, err = strconv.ParseInt(padded, 10, 64)
		if err != nil || fraction < 0 {
			return 0, ErrInvalidInput
		}
	}
	if whole > (math.MaxInt64-fraction)/100 {
		return 0, ErrOverflow
	}
	return whole*100 + fraction, nil
}

func FormatMajorAmount(amountMinor int64) string {
	return fmt.Sprintf("%d.%02d", amountMinor/100, amountMinor%100)
}
