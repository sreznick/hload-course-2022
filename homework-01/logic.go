package main

import (
	"sync"
)

var createUrlLock = &sync.Mutex{}

func addUrl(longUrl string) (string, error) {
	db := getDatabaseInstance()
	{
		createUrlLock.Lock()
		defer createUrlLock.Unlock()

		tinyUrl, err := db.GetIdByUrl(longUrl)
		if err == nil {
			return tinyUrl, nil
		}

		tinyUrl, err = db.AddNewUrl(longUrl)
		if err != nil {
			return "", err
		}

		return tinyUrl, nil
	}
}

func getUrl(tinyUrl string) (string, error) {
	db := getDatabaseInstance()
	url, err := db.GetUrlById(tinyUrl)
	return url, err
}
