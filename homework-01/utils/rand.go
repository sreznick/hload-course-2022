package utils

import (
	"math/rand"
	"strings"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func RandKey() string {
	var build strings.Builder
	build.Grow(7)
	for i := 0; i < 7; i++ {
		build.WriteRune(letters[rand.Intn(len(letters))])
	}
	return build.String()
}
