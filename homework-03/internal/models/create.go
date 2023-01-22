package models

type Create struct {
	LongUrl    string `json:"long_url"`
	ShortUrlID int64  `json:"short_url_id"`
}
