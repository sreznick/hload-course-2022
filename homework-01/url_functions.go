package main

import "fmt"

const alphabet = 62
const maxLen = 7

func IdToTinyUrl(id int) (string, error) {
	if id < 0 || id >= alphabet^maxLen {
		return "", fmt.Errorf("wrong id")
	}

	bytes := make([]byte, 7)

	for i := 0; i < maxLen; i++ {
		bytes[i] = intToByte(id % alphabet)
		id /= alphabet
	}

	return string(bytes), nil
}

func intToByte(id int) byte {
	if id < 10 {
		return byte('0' + id)
	}

	if id >= 10 && id < 36 {
		return byte('a' + (id - 10))
	}

	return byte('A' + (id - 36))
}

func TinyUrlToId(url string) (int, error) {
	if len(url) != maxLen {
		return 0, fmt.Errorf("wrong url size")
	}

	bytes := []byte(url)
	id := 0

	for i := 0; i < maxLen; i++ {
		id *= alphabet
		char, err := byteToInt(bytes[maxLen-i-1])
		if err != nil {
			return 0, err
		}

		id += char
	}

	return id, nil
}

func byteToInt(b byte) (int, error) {
	if b >= '0' && b <= '9' {
		return int(b - '0'), nil
	}

	if b >= 'a' && b <= 'z' {
		return int(b - 'a' + 10), nil
	}

	if b >= 'A' && b <= 'Z' {
		return int(b - 'A' + 36), nil
	}

	return 0, fmt.Errorf("wrong symbol")
}
