package db_interactor

import (
	"fmt"

	_ "github.com/lib/pq"

	"database/sql"

	"github.com/markphelps/optional"
)

type DbInteractor struct {
	Conn *sql.DB
}

func OpenSQLConnection() DbInteractor {
	fmt.Println(sql.Drivers())

	conn, _ := sql.Open("postgres", "postgres://postgres:dsc@localhost/mydb1?sslmode=disable")

	err := conn.Ping()
	if err != nil {
		panic("Unable to connect db")
	}
	return DbInteractor{conn}
}

func (db *DbInteractor) CreateTableIfNotExists() {
	rows, err := db.Conn.Query("CREATE TABLE IF NOT EXISTS urlmap (short_url text primary key, long_url text)")
	if err != nil {
		fmt.Println("Unable to create table")
	}
	defer rows.Close()
}

func (db *DbInteractor) GetLongURL(short_url string) optional.String {
	rows, err := db.Conn.Query(fmt.Sprintf("SELECT long_url FROM urlmap WHERE short_url = '%s'", short_url))
	if err != nil {
		fmt.Println("Db error occured")
		return optional.String{}
	}
	defer rows.Close()
	for rows.Next() {
		var long_url_db string
		rows.Scan(&long_url_db)
		return optional.NewString(long_url_db)
	}
	fmt.Println("Unable to find short url in db")
	return optional.String{}
}

func (db *DbInteractor) InsertURL(short_url string, long_url string) optional.String {
	rows, err := db.Conn.Query(fmt.Sprintf("INSERT INTO urlmap values ('%s', '%s') ON CONFLICT(short_url) DO UPDATE SET long_url=urlmap.long_url  RETURNING urlmap.long_url", short_url, long_url))
	if err != nil {
		fmt.Println("Db error during insert")
		return optional.String{}
	}
	defer rows.Close()
	for rows.Next() {
		var long_url_db string
		rows.Scan(&long_url_db)
		return optional.NewString(long_url_db)
	}
	return optional.String{}
}
