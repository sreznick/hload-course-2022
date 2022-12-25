package urlHandler

import (
	"fmt"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

const TINY_URL_LENGTH = 7

var ID_TO_URL = map[uint8]string{'0': "f", '1': "A", '2': "Z", '3': "q", '4': "9", '5': "5", '6': "M", '7': "o", '8': "J", '9': "w"}
var URL_TO_ID = map[uint8]string{'f': "0", 'A': "1", 'Z': "2", 'q': "3", '9': "4", '5': "5", 'M': "6", 'o': "7", 'J': "8", 'w': "9"}

func GenerateTinyUrl(id int) string {
	strId := strconv.Itoa(id)
	length := len(strId)

	if length > TINY_URL_LENGTH {
		panic("Id out of domain. No tiny url generated")
	}

	str := strings.Repeat("0", TINY_URL_LENGTH-length)

	for i := 0; i < length; i++ {
		symbol, return_code := ID_TO_URL[strId[i]]
		if !return_code {
			panic("Invalid id")
		}
		str += symbol
	}
	return str
}

func GetIdByTinyUrl(tinyUrl string) (int, error) {
	length := len(tinyUrl)

	if length > TINY_URL_LENGTH {
		return -1, fmt.Errorf("invalid tiny url")
	}

	str := ""
	for i := 0; i < length; i++ {
		if tinyUrl[i] == '0' {
			continue
		}
		symbol, return_code := URL_TO_ID[tinyUrl[i]]
		if !return_code {
			return -1, fmt.Errorf("invalid tiny url")
		}
		str += symbol
	}

	id, _ := strconv.Atoi(str)
	return id, nil

}
