package dtos

import (
	"math"
	"strconv"
	"strings"
)

// FormatIDR formats an int64 amount into Indonesian Rupiah representation (e.g. Rp 1.250.000.000).
func FormatIDR(amount int64) string {
	if amount == math.MinInt64 {
		return "-Rp 9.223.372.036.854.775.808"
	}
	isNeg := false
	if amount < 0 {
		isNeg = true
		amount = -amount
	}
	s := strconv.FormatInt(amount, 10)
	n := len(s)
	if n <= 3 {
		if isNeg {
			return "-Rp " + s
		}
		return "Rp " + s
	}

	var sb strings.Builder
	if isNeg {
		sb.WriteString("-Rp ")
	} else {
		sb.WriteString("Rp ")
	}
	rem := n % 3
	if rem > 0 {
		sb.WriteString(s[:rem])
		if rem < n {
			sb.WriteString(".")
		}
	}
	for i := rem; i < n; i += 3 {
		sb.WriteString(s[i : i+3])
		if i+3 < n {
			sb.WriteString(".")
		}
	}
	return sb.String()
}
