package main

import (
	"errors"
	"math"
	"strings"
)

const (
	base         uint64 = 62
	characterSet        = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

func toBase62(num uint64) string {
	builder := strings.Builder{}
	for num > 0 {
		r := num % base
		num /= base
		builder.WriteString(string(characterSet[r]))
	}
	result := builder.String()
	url := result + strings.Repeat("0", int(math.Max(0, float64(7-len(result)))))
	if len(url) > 7 {
		panic("All tinyurl are busy!")
	}
	return url
}

func formBase62(encoded string) (uint64, error) {
	var res uint64
	for pow, char := range encoded {
		pos := strings.IndexRune(characterSet, char)
		if pos == -1 {
			return 0, errors.New("Invalic char: " + string(char))
		}
		res += uint64(pos) * uint64(math.Pow(float64(base), float64(pow)))
	}
	return res, nil
}
