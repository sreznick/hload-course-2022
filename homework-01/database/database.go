package database

import (
	"database/sql"
	"fmt"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "228"
	dbname   = "urls"
)

const (
	insertNewUrl = "insert into url_mappings(url) values ($1) on conflict do nothing"
	selectByUrl  = "select id from url_mappings where url = $1"
	selectById   = "select url from url_mappings where id = $1"
)

const SqlDriver = "postgres"

type UrlTable struct {
	Connection *sql.DB
}

func (db *UrlTable) SetupTable() {
	_, err := db.Connection.Exec("create table if not exists url_mappings(id  serial, url text)")
	if err != nil {
		fmt.Println("Failed to create table", err)
		panic("exit")
	}
}

func (db *UrlTable) InsertNewUrl(url string) error {
	_, err := db.Connection.Exec(insertNewUrl, url)
	return err
}

func (db *UrlTable) GetTinyUrlIdByLongUrl(url string) (int64, error) {
	var tinyUrlId int64
	err := db.Connection.QueryRow(selectByUrl, url).Scan(&tinyUrlId)
	return tinyUrlId, err
}

func (db *UrlTable) GetLongUrlByTinyUrlId(id int64) (string, error) {
	var longUrl string
	err := db.Connection.QueryRow(selectById, id).Scan(&longUrl)
	return longUrl, err
}

func CreateConnection() *UrlTable {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	conn, err := sql.Open(SqlDriver, psqlInfo)
	if err != nil {
		fmt.Println("Failed to open", err)
		panic("exit")
	}

	err = conn.Ping()
	if err != nil {
		fmt.Println("Failed to ping database: ", err)
		panic("exit")
	}
	return &UrlTable{conn}
}
