package vt100

// umin finds the smallest uint
func umin(a, b uint) uint {
	if a < b {
		return a
	}
	return b
}
