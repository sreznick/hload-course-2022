package worker

import (
  "fmt"
  "net/http"
  "os"
  "strings"

  shared "hw3/shared"

  "github.com/gin-gonic/gin"
  "github.com/confluentinc/confluent-kafka-go/kafka"
  "github.com/go-redis/redis/v8"
)

const CLICKS_REPORT_PERIOD = 100

var config shared.WorkerConfig
var redisClient *redis.Client
var kafkaConsumer *kafka.Consumer
var kafkaProducer *kafka.Producer

var hostname string

func assignHostname() {
  var err error
  if hostname, err = os.Hostname(); err != nil {
    panic(err)
  }

  fmt.Printf("[worker] assigned hostname %s\n", hostname)
}

func FetchLongURL(c *gin.Context) {
  tinyurl := c.Params.ByName("tinyurl")
  result, err := fetchRecord(tinyurl)

  if err != nil {
    fmt.Println("Failed to fetch record")
    c.JSON(http.StatusInternalServerError, gin.H{})
    return
  }

  counterValue, err := incrementRecordCounter(tinyurl)
  if err != nil {
    fmt.Println("Failed to increment counter (ignoring the error)")
  } else {
    if counterValue % CLICKS_REPORT_PERIOD == 0 {
        reportCounterProgress(kafkaProducer, tinyurl)
    }
  }

  c.Redirect(http.StatusFound, result)
}

// acks: 1 for at-least-once delivery
func setupKafkaProducer() {
  var err error
  kafkaProducer, err = kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": config.Kafka, "acks": 1})

  if err != nil {
      fmt.Printf("Failed to create producer: %s\n", err)
      os.Exit(1)
  }

  fmt.Println("[worker] kafka-producer: OK")
}

func setupRedis() {
  redisClient = redis.NewClient(&redis.Options{
        Addr:     config.Redis,
        Password: "",
        DB:       0,
  })

  pingRedis()

  fmt.Println("[worker] redis: OK")
}

func subscribeToTopic() {
  err := kafkaConsumer.Subscribe(config.Topics.Tinurls, nil)
  if err != nil {
    panic(err)
  } else {
    fmt.Println("[worker] subscribed to kafka topic " + config.Topics.Tinurls)
  }

  for {
      ev := kafkaConsumer.Poll(100)
      switch e := ev.(type) {
      case *kafka.Message:
          var tokens = strings.Split(string(e.Value), ",")

          err = insertRecord(tokens[0], tokens[1])
          fmt.Printf("Received record with long %s and tiny %s", tokens[1], tokens[0])
          if err != nil {
            fmt.Printf("An error occured while pushing to redis")
          }
          kafkaConsumer.Commit() // send ACK

      case kafka.Error:
          fmt.Fprintf(os.Stderr, "%% Error: %v\n", e)
          // do nothing
      }
  }

  defer kafkaConsumer.Close()
}

// NOTE:
// Each worker instance creates a kafka consumer with unique (we assume hostname is unique) Consumer Group ID.
// That means every worker instance will read every message pushed to kafka topic.

// enable.auto.commit: false for at-least-once processing guarantee
func setupKafkaConsumer() {
  var err error
  var groupID = fmt.Sprintf("yakurbatov-consumer-group-%s", hostname)

  kafkaConsumer, err = kafka.NewConsumer(&kafka.ConfigMap{
    "bootstrap.servers":    config.Kafka,
    "group.id":             groupID,
    "enable.auto.commit":   false,
  })
  if err != nil {
    panic(err)
  }

  fmt.Printf("[worker] kafka-consumer: OK, consumer group ID: %s\n", groupID)
}

func setupRouter() *gin.Engine {
  gin.SetMode(gin.ReleaseMode)
	router := gin.New()

  router.GET("/:tinyurl", FetchLongURL)

  fmt.Println("[worker] gin: OK")

	return router
}


func Serve(c shared.WorkerConfig) {
  config = c

  assignHostname()
  setupRedis()
  setupKafkaConsumer()
  setupKafkaProducer()
  go subscribeToTopic()

  setupRouter().Run(config.Socket)
}
