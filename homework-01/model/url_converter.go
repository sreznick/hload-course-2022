package model

import (
	"fmt"
)

const (
	alphabetSize      = 62
	latinAlphabetSize = 26
	messageLength     = 7
)

func TinyUrlIdToTinyUrl(id int64) (string, error) {
	bytes := make([]byte, messageLength)

	for i := range bytes {
		var err error
		bytes[i], err = intToByte(id % alphabetSize)
		if err != nil {
			return "", err
		}

		id /= alphabetSize
	}

	return string(bytes), nil
}

func TinyUrlToTinyUrlId(url string) (int64, error) {
	bytes := []byte(url)
	id := int64(0)

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

func intToByte(number int64) (byte, error) {
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

func byteToInt(b byte) (int64, error) {
	if b >= 'a' && b <= 'z' {
		return int64(b - 'a'), nil
	}

	if b >= 'A' && b <= 'Z' {
		return int64(b - 'A' + 26), nil
	}

	if b >= '0' && b <= '9' {
		return int64(b - '0' + 52), nil
	}

	return 0, fmt.Errorf("incorrect character in url %c", b)
}
