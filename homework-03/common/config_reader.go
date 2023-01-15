package common

import (
	"encoding/json"
	"os"
)

type KafkaConfig struct {
	UrlTopicName    string   `json:"urlTopicName"`
	ClicksTopicName string   `json:"clicksTopicName"`
	MessageDelim    string   `json:"messageDelim"`
	ClicksThrsh     int      `json:"clicksThrsh"`
	ClicksBrokers   []string `json:"clicksBrokers"`
	UrlsBrokers     []string `json:"urlsBrokers"`

	ClicksProducing string `json:"clicksProducing"`
	UrlsProducing   string `json:"urlsProducing"`

	ClicksGroup string `json:"clicksGroup"`
	UrlsGroup   string `json:"urlsGroup"`
}

type RedisConfig struct {
	Ip string
}

func GetKafkaConfig() KafkaConfig {
	buf, err := os.ReadFile("configuration/conf.json")
	if err != nil {
		panic(err)
	}

	c := &KafkaConfig{}
	err = json.Unmarshal(buf, c)
	if err != nil {
		panic(err)
	}

	return *c
}

func GetRedisConfig() RedisConfig {
	buf, err := os.ReadFile("configuration/redis_conf.json")
	if err != nil {
		panic(err)
	}

	c := &RedisConfig{}
	err = json.Unmarshal(buf, c)
	if err != nil {
		panic(err)
	}

	return *c
}
