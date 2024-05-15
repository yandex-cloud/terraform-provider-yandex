package math

// TODO (miksmolin) will have been replaced with function from std package after go 1.21 update
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
