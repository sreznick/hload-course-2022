package model

import (
	"fmt"
	"main/database"
)

func GetTinyUrlByLongUrl(db *database.UrlTable, longUrl string) (string, error) {
	var tinyUrl string
	var err error

	err = db.InsertNewUrl(longUrl)
	if err != nil {
		return tinyUrl, fmt.Errorf("DB query failed: " + err.Error())
	}

	tinyUrlId, err := db.GetTinyUrlIdByLongUrl(longUrl)
	if err != nil {
		return tinyUrl, fmt.Errorf("DB query failed: " + err.Error())
	}

	tinyUrl, err = TinyUrlIdToTinyUrl(tinyUrlId)
	if err != nil {
		return tinyUrl, fmt.Errorf("Can't decode urls " + err.Error())
	}

	return tinyUrl, nil
}

func GetLongUrlByTinyUrl(db *database.UrlTable, tinyUrl string) (string, error) {
	id, err := TinyUrlToTinyUrlId(tinyUrl)
	if err != nil {
		return "", err
	}
	longUrl, err := db.GetLongUrlByTinyUrlId(id)
	if err != nil {
		return "", err
	}
	return longUrl, nil
}
