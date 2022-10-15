package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/ssh"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

const (
	CERT_PASSWORD        = 1
	CERT_PUBLIC_KEY_FILE = 2
	DEFAULT_TIMEOUT      = 3 // second
)

type SSH struct {
	Ip      string
	User    string
	Cert    string //password or key file path
	Port    int
	session *ssh.Session
	client  *ssh.Client
}

type DBConnect struct {
	Ip   string
	User string
	Cert string
	Name string

	db *sqlx.DB
}

type urlDB struct {
	Key string `db:"short_url"`
	URL string `db:"long_url"`
}

func randKey() string {
	var build strings.Builder
	build.Grow(10)
	for i := 0; i < 10; i++ {
		build.WriteRune(letters[rand.Intn(len(letters))])
	}
	return build.String()
}

func (sshClient *SSH) readPublicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

func (sshClient *SSH) Connect(mode int) {

	var ssh_config *ssh.ClientConfig
	var auth []ssh.AuthMethod
	if mode == CERT_PASSWORD {
		auth = []ssh.AuthMethod{ssh.Password(sshClient.Cert)}
	} else if mode == CERT_PUBLIC_KEY_FILE {
		auth = []ssh.AuthMethod{sshClient.readPublicKeyFile(sshClient.Cert)}
	} else {
		log.Println("does not support mode: ", mode)
		return
	}

	ssh_config = &ssh.ClientConfig{
		User: sshClient.User,
		Auth: auth,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: time.Second * DEFAULT_TIMEOUT,
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", sshClient.Ip, sshClient.Port), ssh_config)
	if err != nil {
		fmt.Println(err)
		return
	}

	session, err := client.NewSession()
	if err != nil {
		fmt.Println(err)
		client.Close()
		return
	}

	sshClient.session = session
	sshClient.client = client
}

func (sshClient *SSH) RunCmd(cmd string) {
	out, err := sshClient.session.CombinedOutput(cmd)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(out))
}

func (sshClient *SSH) Close() {
	sshClient.session.Close()
	sshClient.client.Close()
}

func (dbClient *DBConnect) Open() {
	db, err := sqlx.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbClient.User, dbClient.Cert, dbClient.Ip, dbClient.Name))

	if err != nil {
		fmt.Println(err)
	}

	dbClient.db = db
}

func (dbClient *DBConnect) Close() {
	dbClient.db.Close()
}

func (dbClient *DBConnect) store(key string, url string) {
	dbClient.db.MustExec("INSERT INTO urls (short_url, long_url) VALUES ($1, $2)", key, url)
}

func (dbClient *DBConnect) loadKey(url string) (key string, ok bool) {
	el := urlDB{}
	dbClient.db.Get(&el, "SELECT * FROM urls WHERE long_url=$1", url)
	key, ok = el.Key, el.Key != ""
	return
}

func (dbClient *DBConnect) loadURL(key string) (url string, ok bool) {
	el := urlDB{}
	dbClient.db.Get(&el, "SELECT * FROM urls WHERE short_url=$1", key)
	url, ok = el.URL, el.URL != ""
	return
}

// demo
func main() {
	server := &SSH{
		Ip:   "217.25.88.166",
		User: "root",
		Port: 22,
		Cert: "Rhs:jgjkm166",
	}

	server.Connect(CERT_PASSWORD)
	defer server.Close()

	client := &DBConnect{
		Ip:   "127.0.0.1",
		User: "postgres",
		Name: "url_shortener",
		Cert: "Rhs:jgjkm166"}

	client.Open()
	defer client.Close()

}
