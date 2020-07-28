package xstrconv

import "strconv"

// FormatFloat64 float64转string
func FormatFloat64(f float64) string {
	return strconv.FormatFloat(f, 'f', 6, 64)
}
