package random

import "math/rand"

func Random(n int) int {
	return rand.Intn(n-1) + 1
}
