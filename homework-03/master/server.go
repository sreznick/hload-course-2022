package master

import (
  "fmt"
  "net/http"
  "database/sql"
  "os"
  "strconv"
  "strings"

  shared "hw3/shared"

  "github.com/gin-gonic/gin"
  "github.com/confluentinc/confluent-kafka-go/kafka"
)

const SQL_DRIVER = "postgres"

var config shared.MasterConfig
var PGConnection *sql.DB
var kafkaProducer *kafka.Producer
var kafkaConsumer *kafka.Consumer

type CreateTinyURLRequestModel struct {
  Longurl string `json:"longurl" binding:"required"`
}

func sendToKafka(longurl string, tinyurl string) {
  delivery_chan := make(chan kafka.Event, 10000)
  var err = kafkaProducer.Produce(&kafka.Message{
      TopicPartition: kafka.TopicPartition{Topic: &config.Topics.Tinurls, Partition: kafka.PartitionAny},
      Value: []byte(tinyurl + "," + longurl)},
      delivery_chan,
  )
  if err != nil {
    fmt.Println(err)
  } else {
    fmt.Println("Sent message to topic " + config.Topics.Tinurls)
  }
}

func CreateTinyURL(c *gin.Context) {
  var requestModel CreateTinyURLRequestModel

  if err := c.BindJSON(&requestModel); err != nil {
    c.JSON(http.StatusUnprocessableEntity, gin.H{})
    return
  }

  result, err := CreateTinyURLService(requestModel.Longurl)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{})
    return
  }

  sendToKafka(requestModel.Longurl, result)

  c.JSON(http.StatusOK, gin.H{"longurl": requestModel.Longurl, "tinyurl": result})
}

func setupRouter() *gin.Engine {
  gin.SetMode(gin.ReleaseMode)
	router := gin.New()

  router.PUT("/create", CreateTinyURL)

  fmt.Println("[master] gin: OK")

	return router
}

func setupPGConnection() {
  var err error
  PGConnection, err = sql.Open(SQL_DRIVER, config.Postgres)
  if err != nil {
      fmt.Println("Failed to open", err)
      panic("exit")
  }

  fmt.Println("[master] pg: OK")
}

// enable.auto.commit: false for at-least-once processing guarantee
func setupKafkaConsumer() {
  var err error
  kafkaConsumer, err = kafka.NewConsumer(&kafka.ConfigMap{
    "bootstrap.servers":    config.Kafka,
    "group.id":             "yakurbatov-consumer-group",
    "enable.auto.commit":   false,
  })
  if err != nil {
    panic(err)
  }

  fmt.Println("[server] kafka-consumer: OK")
}

// acks: 1 for at-least-once delivery
func setupKafkaProducer() {
  var err error
  kafkaProducer, err = kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": config.Kafka, "acks": 1})

  if err != nil {
      fmt.Printf("Failed to create producer: %s\n", err)
      os.Exit(1)
  }

  fmt.Println("[master] kafka-producer: OK")
}

func subscribeToKafkaClicksTopic() {
  err := kafkaConsumer.Subscribe(config.Topics.Clicks, nil)
  if err != nil {
    panic(err)
  } else {
    fmt.Println("[master] subscribed to kafka topic " + config.Topics.Clicks)
  }

  for {
      ev := kafkaConsumer.Poll(100)
      switch e := ev.(type) {
      case *kafka.Message:
          fmt.Println("[server] Received message with data", string(e.Value))
          var tokens = strings.Split(string(e.Value), ",")
          var delta, err = strconv.Atoi(tokens[1])
          if err != nil {
            panic(err)
          }

          fmt.Println("Updating clicks counter by " + strconv.Itoa(delta))

          UpdateClicksCounter(string(tokens[0]), delta)
          kafkaConsumer.Commit() // send ACK


      case kafka.Error:
          fmt.Fprintf(os.Stderr, "%% Error: %v\n", e)
          // do nothing club
      }
  }
}

func Serve(c shared.MasterConfig) {
  config = c

  setupPGConnection()
  setupKafkaProducer()
  setupKafkaConsumer()
  go subscribeToKafkaClicksTopic()
  setupRouter().Run(config.Socket)
}
