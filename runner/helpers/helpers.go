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

// Generate a randomized order of an array of integer values from 0 to modulo-1
func RandomizeCyclicGroup(modulo int) []int {
	var arr []int

	for i := 0; i < modulo; i++ {
		arr = append(arr, i)
	}

	for i := 0; i < len(arr); i++ {
		newRandomIndex := rand.Intn(len(arr))
		arr[i], arr[newRandomIndex] = arr[newRandomIndex], arr[i]
	}

	return arr
}
