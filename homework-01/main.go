package main

import (
	"fmt"
	_ "github.com/lib/pq"
	_ "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io/ioutil"
	"net/http"
	"urlShortener/utils"
	_ "urlShortener/utils"
)

func main() {

	b, err := ioutil.ReadFile("pass.conf")
	if err != nil {
		fmt.Print(err)
		return
	}

	// convert bytes to string
	pass := string(b)

	server := &utils.SSH{
		Ip:   "217.25.88.166",
		User: "root",
		Port: 22,
		Cert: pass,
	}

	err = server.Connect(utils.CERT_PASSWORD)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer server.Close()

	utils.InitConection(*server)

	client := &utils.DBConnect{
		Ip:   "localhost",
		User: "postgres",
		Name: "url_shortener",
		Cert: pass}

	err = client.Open()

	if err != nil {
		fmt.Println(err)
		return
	}

	defer client.Close()

	utils.InitData(client)

	utils.RegPrometheus()
	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/", utils.HandleGet)
	http.HandleFunc("/ping", utils.HandlePing)
	http.HandleFunc("/create", utils.HandlePut)

	fmt.Println("Server started")

	if err := http.ListenAndServe(":8000", nil); err != nil {
		panic("error!")
	}

}
