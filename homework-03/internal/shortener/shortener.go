package shortener

import (
	"strconv"
)

func IdToUrl(id int64) string {
	return strconv.FormatInt(id, 36)
}

func UrlToId(id string) (int64, error) {
	return strconv.ParseInt(id, 36, 64)
}
