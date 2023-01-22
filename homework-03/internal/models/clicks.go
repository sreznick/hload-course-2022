package models

type Clicks struct {
	ShortUrlID int64 `json:"short_url_id"`
	Inc        int64 `json:"inc"`
}
