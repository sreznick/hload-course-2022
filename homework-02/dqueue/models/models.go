package models

import "errors"

var NotExisingQueueErr = errors.New("Queue dose not exist in zookeeper")
var QueueIsEmptyErr = errors.New("Queue is empty")

type QueueInfo struct {
	Shards  []*Host
	Current int
}

type Host struct {
	Addr string
	Key  string
}

type RedisValue struct {
	Value     string
	Timestamp int64
}
