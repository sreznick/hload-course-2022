package server

import (
	"fmt"
	_ "github.com/lib/pq"
)

const (
	tinyUrlLength = 7
	domain        = 62
	alphabetSize  = 26
)

func GenerateTinyUrl(id int) (string, error) {
	tinyUrlBytes := make([]byte, tinyUrlLength)

	for i := range tinyUrlBytes {
		var err error
		tinyUrlBytes[i], err = getByte(id % domain)
		if err != nil {
			return "", err
		}
		id /= domain
	}

	return string(tinyUrlBytes), nil
}

func GetIdByTinyUrl(tinyUrl string) (int, error) {
	tinyUrlBytes := []byte(tinyUrl)
	id := 0
	for i := len(tinyUrlBytes) - 1; i >= 0; i-- {
		id *= domain
		byteId, err := byteToInt(tinyUrlBytes[i])
		if err != nil {
			return 0, err
		}
		id += byteId
	}
	return id, nil
}

func HandleError(err error, msg string) {
	if err != nil {
		fmt.Println(msg, err)
		panic("Exit")
	}
}

func getByte(id int) (byte, error) {
	if id < alphabetSize {
		return byte('a' + id), nil
	}

	if id >= alphabetSize && id < alphabetSize*2 {
		return byte('A' + (id - alphabetSize)), nil
	}

	return byte('0' + (id - alphabetSize*2)), nil
}

func byteToInt(b byte) (int, error) {
	if b >= 'a' && b <= 'z' {
		return int(b - 'a'), nil
	}

	if b >= 'A' && b <= 'Z' {
		return int(b - 'A' + alphabetSize), nil
	}

	if b >= '0' && b <= '9' {
		return int(b - '0' + alphabetSize*2), nil
	}

	return 0, fmt.Errorf("incorrect character")
}
