package server

import (
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"math"
	"strings"
)

const (
	base     int = 62
	alphabet     = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

func GenerateTinyUrl(id int) string {
	builder := strings.Builder{}
	for id > 0 {
		r := id % base
		id /= base
		builder.WriteString(string(alphabet[r]))
	}
	result := builder.String()
	url := result + strings.Repeat("0", int(math.Max(0, float64(7-len(result)))))
	return url
}

func GetIdByTinyUrl(tinyUrl string) (int, error) {
	var res int
	for pow, char := range tinyUrl {
		pos := strings.IndexRune(alphabet, char)
		if pos == -1 {
			return 0, errors.New("Unknown char: " + string(char))
		}
		res += pos * int(math.Pow(float64(base), float64(pow)))
	}
	return res, nil
}

func HandleError(err error, msg string) {
	if err != nil {
		fmt.Println(msg, err)
		panic("Exit")
	}
}
