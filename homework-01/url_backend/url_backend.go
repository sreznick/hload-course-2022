package url_backend

import "fmt"

const (
	domain = 62
)

func IdToUrl(id int64) (string, error) {
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

func UrlToId(url string) (int64, error) {
	bytes := []byte(url)
	var id int64 = 0

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

func intToByte(id int64) (byte, error) {
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

	return 0, fmt.Errorf("wrong character")
}
