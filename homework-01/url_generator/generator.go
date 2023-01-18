package url_generator

import (
	"hash/fnv"
	"math"
)

var CHARS_NUM = int64(62)
var SHORT_URL_LEN = 7
var ALPHABET_LEN = int64(26)
var MAX_URL_NUMBER = int64(math.Pow(float64(CHARS_NUM), float64(SHORT_URL_LEN)))

func IntToShortUrl(n int64) string {

	if n <= 0 || n >= MAX_URL_NUMBER {
		panic("Number out of bound")
	}
	var res string
	var cur = n
	for i := 0; i < SHORT_URL_LEN; i++ {
		var d = cur % CHARS_NUM
		switch d / ALPHABET_LEN {
		case 0:
			res += string(rune('a' + d%ALPHABET_LEN))
		case 1:
			res += string(rune('A' + d%ALPHABET_LEN))
		default:
			res += string(rune('0' + d%ALPHABET_LEN))
		}
		cur /= CHARS_NUM
	}
	return res
}

func hash(s string) int64 {
	h := fnv.New64()
	h.Write([]byte(s))
	return int64(h.Sum64())
}

func GenerateShortUrl(long_url string) (int64, string) {
	url_id := ((hash(long_url) % MAX_URL_NUMBER) + MAX_URL_NUMBER) % MAX_URL_NUMBER
	return url_id, IntToShortUrl(url_id)
}
