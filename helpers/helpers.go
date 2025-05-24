package helpers

import (
	"math/rand"
	"time"
)

// Generate a random time duration between min and max
func GetRandomTime(min time.Duration, max time.Duration) time.Duration {
	timeRange := max - min

	return min + time.Duration(rand.Intn(int(timeRange)))
}

func gcd(a int, b int) int {
	for b != 0 {
		a, b = b, a%b
	}

	return a
}

// Generate a randomized order of an array of integer values from 0 to modulo-1
func RandomizeCyclicGroup(modulo int, randomness int) []int {
	var arr []int
	var k int
	for i := modulo; i >= 1; i-- {
		arr = append(arr, modulo-i)
		if gcd(i-1, modulo) == 1 {
			k = i - 1
		}
	}

	randomNum := rand.Intn(modulo-1) + 1

	for randomness > 0 {
		for i := range arr {
			arr[i] = (arr[i] + randomNum) * k % modulo
		}

		if randomness > 20 {
			randomness -= 10
		} else {
			randomness--
		}
	}

	return arr
}
