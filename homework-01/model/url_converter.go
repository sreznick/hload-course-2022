package model

import (
	"fmt"
)

const (
	alphabetSize      = 62
	latinAlphabetSize = 26
	messageLength     = 7
)

func TinyUrlIdToTinyUrl(id int) (string, error) {
	bytes := make([]byte, messageLength)

	for i, _ := range bytes {
		var err error
		bytes[i], err = intToByte(id % alphabetSize)
		if err != nil {
			return "", err
		}

		id /= alphabetSize
	}

	return string(bytes), nil
}

func TinyUrlToTinyUrlId(url string) (int, error) {
	bytes := []byte(url)
	id := 0

	for i := len(bytes) - 1; i >= 0; i-- {
		id *= alphabetSize
		byteId, err := byteToInt(bytes[i])
		if err != nil {
			return 0, err
		}

		id = id*alphabetSize + byteId
	}

	return id, nil
}

func intToByte(number int) (byte, error) {
	if number > alphabetSize {
		return '0', fmt.Errorf("incorrect argument %d should be less or equal %d", number, alphabetSize)
	}

	if number < latinAlphabetSize {
		return byte('a' + number), nil
	} else if number < 2*latinAlphabetSize {
		return byte('A' + (number - latinAlphabetSize)), nil
	} else {
		return byte('0' + (number - 2*latinAlphabetSize)), nil
	}
}

func byteToInt(b byte) (int, error) {
	if b >= 'a' && b <= 'z' {
		return int(b - 'a'), nil
	}

	if b >= 'A' && b <= 'Z' {
		return int(b - 'A' + 26), nil
	}

	if b >= '0' && b <= '9' {
		return int(b - '0' + 52), nil
	}

	return 0, fmt.Errorf("incorrect character in url %c", b)
}
