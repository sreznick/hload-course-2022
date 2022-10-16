package main

import (
	"database/sql"
)

const (
	createTableSQL = "CREATE TABLE IF NOT EXISTS tinyurls(id SERIAL, url VARCHAR UNIQUE);"
	addUrlSQL      = "INSERT INTO tinyurls(url) VALUES ($1) RETURNING id;"
	getIdSQL       = "SELECT id FROM tinyurls WHERE url = $1;"
	getUrlSQL      = "SELECT url FROM tinyurls WHERE id = $1;"
)

func createTable(db *sql.DB) error {
	_, err := db.Exec(createTableSQL)
	return err
}

func addUrl(db *sql.DB, longUrl string) (string, error) {
	var id uint64
	err := db.QueryRow(getIdSQL, longUrl).Scan(&id)
	if err != nil {
		err := db.QueryRow(addUrlSQL, longUrl).Scan(&id)
		if err != nil {
			return "", err
		}
	}

	return toBase62(id), nil
}

func getUrl(db *sql.DB, tinyUrl string) (string, error) {
	id, errUrl := formBase62(tinyUrl)
	if errUrl != nil {
		return "", errUrl
	}
	var url string
	err := db.QueryRow(getUrlSQL, id).Scan(&url)
	if err != nil {
		return "", err
	}
	return url, nil
}
