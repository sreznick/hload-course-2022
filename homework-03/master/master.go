package main

import (
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	_ "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"net/http"
	"strings"
	"urlShortener/utils"
	_ "urlShortener/utils"

	_ "github.com/confluentinc/confluent-kafka-go/kafka"
)

var clients = map[string]bool{"bmadzhuga-client-a": true, "bmadzhuga-client-b": true}

func main() {

	b, err := ioutil.ReadFile("pass.conf")
	if err != nil {
		panic(err)
	}

	key, err := ioutil.ReadFile("//users//bogdan//.ssh//id_ed25519")
	if err != nil {
		panic(err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		panic(err)
	}

	// convert bytes to string
	pass := string(b)

	server := &utils.SSH{
		Ip:     "158.160.9.8",
		User:   "bmadzhuga",
		Port:   22,
		Cert:   pass,
		Signer: signer,
	}

	err = server.Connect(utils.CERT_PUBLIC_KEY_FILE)
	if err != nil {
		panic(err)
	}

	defer server.Close()

	utils.InitConection(*server)

	client := &utils.DBConnect{
		Ip:   "localhost",
		User: "postgres",
		Name: "bmadzhuga",
		Cert: pass}

	err = client.Open()

	if err != nil {
		panic(err)
	}

	kafkaMasterClient := &utils.Kafka{
		Topic: "bmadzhuga-events",
		Type:  "master",
	}

	err = kafkaMasterClient.Connect()

	if err != nil {
		panic(err)
	}

	utils.ClientKafka = kafkaMasterClient

	kafkaMetricClient := &utils.Kafka{
		Topic: "bmadzhuga-metrics",
		Type:  "master",
	}

	err = kafkaMetricClient.Connect()

	if err != nil {
		panic(err)
	}

	redis := utils.Redis{Cluster: "158.160.19.212:26379"}
	err = redis.Connect()

	if err != nil {
		panic(err)
	}

	defer redis.Close()
	defer client.Close()
	defer kafkaMasterClient.Close()

	go listenTopic(kafkaMasterClient)
	go listenMetrics(kafkaMetricClient, client)

	utils.ClientBD = client

	utils.RegPrometheus()
	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/", utils.HandleGet)
	http.HandleFunc("/ping", utils.HandlePing)
	http.HandleFunc("/create", utils.HandlePut)

	fmt.Println("Server started")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}

}

// Прослушиваем очередь обновленя метрик
func listenMetrics(client *utils.Kafka, database *utils.DBConnect) {
	if client.Consumer == nil {
		panic(errors.New("Empty consumer"))
	}

	for {
		msg, err := client.Consumer.ReadMessage(-1)
		if err != nil {
			panic(err)
			break
		}
		request := strings.Split(string(msg.Value), "::")

		if len(request) != 3 {
			continue
		}

		val := request[1]
		key := request[0]

		database.SetCounter(key, val)
	}
}

// Прослушиваем очередь запросов от клиентов.
// Если пришел запрос на полную ссылку отправляем
func listenTopic(client *utils.Kafka) {

	if client.Consumer == nil {
		panic(errors.New("Empty consumer"))
	}

	fmt.Println("Start reading from topic: ", client.Topic)

	for {
		msg, err := client.Consumer.ReadMessage(-1)
		if err != nil {
			panic(err)
			break
		}
		request := strings.Split(string(msg.Value), "::")

		if len(request) != 3 {
			continue
		}

		key := request[1]
		topic := request[0]

		if !clients[topic] {
			continue
		}

		url, ok := utils.GetURLFromKey(key)

		kafkaClient := &utils.Kafka{
			Topic: topic,
			Type:  "client",
		}

		err = kafkaClient.Connect()

		if err != nil {
			panic(err)
		}

		err = kafkaClient.Send(key, url, ok)

		fmt.Printf("Success sent answer to %v as %v:%v\n", topic, key, url)

		if err != nil {
			panic(err)
		}

	}

}
