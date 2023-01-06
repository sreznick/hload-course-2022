package main

import (
	"database/sql"
	"fmt"
	"sync"
)

const (
	createTableSQL = "CREATE TABLE IF NOT EXISTS tinyurls(id SERIAL, url VARCHAR UNIQUE);"
	addUrlSQL      = "INSERT INTO tinyurls(url) VALUES ($1) RETURNING id;"
	getIdSQL       = "SELECT id FROM tinyurls WHERE url = $1;"
	getUrlSQL      = "SELECT url FROM tinyurls WHERE id = $1;"
)

var lock = &sync.Mutex{}

type Database struct {
	db *sql.DB
}

var singleDatabase *Database

func getDatabaseInstance() *Database {
	if singleDatabase == nil {
		lock.Lock()
		defer lock.Unlock()
		if singleDatabase == nil {
			singleDatabase = &Database{}
		}
	}

	return singleDatabase
}

func (database *Database) InitDB(sql_driver string, config string) error {
	conn, err := sql.Open(sql_driver, config)
	if err != nil {
		return fmt.Errorf("Failed to open: %e", err)
	}

	err = conn.Ping()
	if err != nil {
		return fmt.Errorf("Failed to ping Database: %e", err)
	}

	database.db = conn

	err = database.createTable()
	if err != nil {
		return fmt.Errorf("Failed to create table: %e", err)
	}

	return nil
}

func (database *Database) createTable() error {
	_, err := database.db.Exec(createTableSQL)
	return err
}

func (database *Database) AddNewUrl(longUrl string) (string, error) {
	db := database.db
	var id uint64
	err := db.QueryRow(addUrlSQL, longUrl).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("Error with url creation: %e", err)
	}
	return toBase62(id), nil
}

func (database *Database) GetIdByUrl(longUrl string) (string, error) {
	db := database.db
	var id uint64
	scanErr := db.QueryRow(getIdSQL, longUrl).Scan(&id)
	if scanErr == sql.ErrNoRows {
		return "", fmt.Errorf("URL %s does not exists. Please save url before using.", longUrl)
	}
	if scanErr != nil {
		return "", scanErr
	}

	return toBase62(id), nil
}

func (database *Database) GetUrlById(tinyUrl string) (string, error) {
	db := database.db
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
