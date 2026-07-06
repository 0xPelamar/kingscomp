package randhelper

import (
	"math/rand/v2"
)

// GenerateDistinctNumbers returns a slice of N distinct int64 numbers between min and max (inclusive).
func GenerateDistinctNumbers(n int, min, max int64) []int64 {
	if min > max {
		return nil
	}
	// Calculate the total available numbers in the range
	// We use uint64 to prevent potential overflow issues if min is deeply negative and max is deeply positive
	rangeSize := uint64(max - min + 1)

	if uint64(n) > rangeSize {
		return nil
	}

	// Use a map to ensure uniqueness
	seen := make(map[int64]struct{}, n)
	result := make([]int64, 0, n)

	for len(result) < n {
		// rand.Uint64N handles the full uint64 range correctly
		val := int64(rand.Uint64N(rangeSize)) + min

		if _, exists := seen[val]; !exists {
			seen[val] = struct{}{}
			result = append(result, val)
		}
	}

	return result
}
