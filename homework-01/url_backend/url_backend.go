package url_backend

import "fmt"

const (
	domain = 62
)

func IdToUrl(id int) (string, error) {
	bytes := make([]byte, 7)

	for i, _ := range bytes {
		var err error
		bytes[i], err = intToByte(id % domain)
		if err != nil {
			return "", err
		}

		id /= domain
	}

	return string(bytes), nil
}

func UrlToId(url string) (int, error) {
	bytes := []byte(url)
	id := 0

	for i := len(bytes) - 1; i >= 0; i-- {
		id *= domain
		byteId, err := byteToInt(bytes[i])
		if err != nil {
			return 0, err
		}

		id += byteId
	}

	return id, nil
}

func intToByte(id int) (byte, error) {
	if id > domain {
		return '0', fmt.Errorf("id out of domain")
	}

	if id < 26 {
		return byte('a' + id), nil
	}

	if id >= 26 && id < 52 {
		return byte('A' + (id - 26)), nil
	}

	return byte('0' + (id - 52)), nil
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

	return 0, fmt.Errorf("wrong character")
}
