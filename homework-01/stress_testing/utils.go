package main

import (
	"math/rand"
)

func min(a int, b int) int {
	if a <= b {
		return a
	} else {
		return b
	}
}

func Pick(list []string) string {
	return list[rand.Intn(len(list))]
}

func RandomString(size int, symbolCount int) string {
	var letters = []rune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")

	s := make([]rune, size)
	for i := range s {
		s[i] = letters[rand.Intn(min(len(letters), symbolCount))]
	}
	return string(s)
}
