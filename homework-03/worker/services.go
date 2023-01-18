package worker

import (
  "context"
  "fmt"
  "strconv"

  "github.com/confluentinc/confluent-kafka-go/kafka"
)

const REDIS_PREFIX string = "yakurbatov_"

var ctx = context.Background()

func insertRecord(tinyurl string, longurl string) error {
  return redisClient.Set(ctx, REDIS_PREFIX + tinyurl, longurl, 0).Err()
}

func fetchRecord(tinyurl string) (string, error) {
  val, err := redisClient.Get(ctx, REDIS_PREFIX + tinyurl).Result()

  return val, err
}

func pingRedis() {
  if err := redisClient.Ping(ctx).Err(); err != nil {
    panic(err)
  }
}

func incrementRecordCounter(tinyurl string) (int64, error) {
  value, err := redisClient.Incr(ctx, REDIS_PREFIX + "counter-" + tinyurl).Result()

  return value, err
}

func reportCounterProgress(kafkaProducer *kafka.Producer, tinyurl string) {
  delivery_chan := make(chan kafka.Event, 10000)
  var err = kafkaProducer.Produce(&kafka.Message{
      TopicPartition: kafka.TopicPartition{Topic: &config.Topics.Clicks, Partition: kafka.PartitionAny},
      Value: []byte(tinyurl + "," + strconv.Itoa(CLICKS_REPORT_PERIOD))},
      delivery_chan,
  )
  if err != nil {
    fmt.Println(err)
  } else {
    fmt.Println("Sent message to topic " + config.Topics.Clicks)
  }
}
