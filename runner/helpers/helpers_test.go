package helpers

import (
	"fmt"
	"slices"
	"testing"
	"time"
)

func TestGetRandomTime(t *testing.T) {
	tests := [][2]time.Duration{
		{time.Millisecond * 100, time.Millisecond * 500},
		{time.Millisecond * 200, time.Millisecond * 500},
		{time.Second * 100, time.Second * 300},
	}

	for _, test := range tests {
		randomTime := GetRandomTime(test[0], test[1])
		if randomTime > test[1] || randomTime < test[0] {
			t.Errorf(
				"Invalid random time, time must be between %v and %v, instead get %v\n",
				test[0],
				test[1],
				randomTime,
			)
		}
	}
}

func checkEqualArr(a []int, b []int) bool {
	if len(a) != len(b) {
		return false
	}

	for _, v := range a {
		if !slices.Contains(b, v) {
			return false
		}
	}

	return true
}

func TestRandomizeCyclicGroup(t *testing.T) {
	modulos := []int{4, 5, 7, 12}

	for _, modulo := range modulos {
		var orderredArr []int
		for i := 0; i < modulo; i++ {
			orderredArr = append(orderredArr, i)
		}

		randomizedArr := RandomizeCyclicGroup(modulo)
		if !checkEqualArr(orderredArr, randomizedArr) {
			t.Errorf(
				"2 arrays are not the same: orderred:%v, randomized:%v\n",
				orderredArr,
				randomizedArr,
			)
		}

		fmt.Printf("Orderred array:%v\nRandomized Array:%v\n",
			orderredArr,
			randomizedArr,
		)
	}
}
