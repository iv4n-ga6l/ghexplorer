package helper

// StringOrNA handle potentially nil strings
func StringOrNA(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}

// Min returns the smaller of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns the larger of two integers
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
