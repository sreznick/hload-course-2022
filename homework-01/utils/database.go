package utils

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"golang.org/x/crypto/ssh"
	"net"
	"time"
)

var sshcon SSH

type ViaSSHDialer struct {
	client *ssh.Client
}

func (self *ViaSSHDialer) Open(s string) (_ driver.Conn, err error) {
	return pq.DialOpen(self, s)
}

func (self *ViaSSHDialer) Dial(network, address string) (net.Conn, error) {
	return self.client.Dial(network, address)
}

func (self *ViaSSHDialer) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	return self.client.Dial(network, address)
}

type DBConnect struct {
	Ip   string
	User string
	Cert string
	Name string

	db *sqlx.DB
}

type UrlDB struct {
	Key string `db:"tinyurl"`
	URL string `db:"longurl"`
}

func InitConection(con SSH) {
	sshcon = con
}

func (dbClient *DBConnect) Open() error {
	sql.Register("postgres+ssh", &ViaSSHDialer{sshcon.client})

	db, err := sqlx.Open("postgres+ssh", fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbClient.User, dbClient.Cert, dbClient.Ip, dbClient.Name))

	if err != nil {
		return err
	}

	dbClient.db = db
	return nil
}

func (dbClient *DBConnect) Close() {
	dbClient.db.Close()
}

func (dbClient *DBConnect) GetURLS() []UrlDB {
	res := []UrlDB{}
	rows, err := dbClient.db.Query("SELECT tinyurl, longurl FROM urls")
	if err != nil {
		return res
	}

	for rows.Next() {
		var key, url string
		err = rows.Scan(&key, &url)

		if err != nil {
			fmt.Println(err)
		}

		el := UrlDB{Key: key, URL: url}

		res = append(res, el)
	}

	return res
}

func (dbClient *DBConnect) store(key string, url string) {
	dbClient.db.MustExec("INSERT INTO urls (tinyurl, longurl) VALUES ($1, $2)", key, url)
}

func (dbClient *DBConnect) loadKey(url string) (key string, ok bool) {
	el := UrlDB{}
	dbClient.db.Get(&el, "SELECT tinyurl, longurl FROM urls WHERE longurl=$1", url)
	key, ok = el.Key, el.Key != ""
	return
}

func (dbClient *DBConnect) loadURL(key string) (url string, ok bool) {
	el := UrlDB{}
	dbClient.db.Get(&el, "SELECT tinyurl, longurl FROM urls WHERE tinyurl=$1", key)
	url, ok = el.URL, el.URL != ""
	return
}
