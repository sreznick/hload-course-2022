package dqueue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-zookeeper/zk"
)

/*
 * Можно создавать несколько очередей
 *
 * Для клиента они различаются именами
 *
 * В реализации они могут потребовать вспомогательных данных
 * Для них - эта структура. Можете определить в ней любые поля
 */
type DQueue struct {
	Name           string
	ShardCount     int
	Hosts          *[]string
	LastShardIndex int
}

type ConfigData struct {
	redisOptions *redis.Options
}

var configData = ConfigData{}

/*
 * Запомнить данные и везде использовать
 */
func Config(redisOptions *redis.Options) {
	configData.redisOptions = redisOptions
}

func getAddr(name string) string {
	return "/zookeeper/epostnikov/" + name
}

func createPath(path string, c *zk.Conn) (string, error) {
	return c.Create(path, []byte{}, 0, zk.WorldACL(zk.PermAll))
}

/*
 * Отдельные узлы Redis-кластера могут выпадать. Availability очереди в целом
 * не должна от этого страдать
 */
func Open(name string, nShards int, addresses *[]string, c *zk.Conn) (DQueue, error) {
	shardsPath := getAddr(name) + "/" + strconv.Itoa(nShards)
	doesExist, _, _ := c.Exists(shardsPath)
	if doesExist {
		var d DQueue
		rawData, _, err := c.Get(shardsPath)
		err = json.Unmarshal(rawData, &d)
		return d, err
	}

	dQuePath := getAddr(name)
	doesExist2, _, err := c.Exists(dQuePath)
	if doesExist2 {
		return DQueue{}, errors.New("already exists")
	}
	d := DQueue{Name: name, ShardCount: nShards, Hosts: addresses, LastShardIndex: 0}
	_, err = createPath(dQuePath, c)
	_, err = createPath(shardsPath, c)
	marshal, err := json.Marshal(d)
	_, err = c.Set(shardsPath, marshal, 0)
	return d, err
}

/*
 * Пишем в очередь. Каждый следующий Push - в следующий шард
 * Если шард упал - пропускаем шард, пишем в следующий по очереди
 */
func (queue *DQueue) Push(value string) error {
	var err error
	for {
		indexOfShard := queue.LastShardIndex % queue.ShardCount
		configData.redisOptions.Addr = (*queue.Hosts)[indexOfShard]
		err = redis.
			NewClient(configData.redisOptions).
			RPush(context.Background(), queue.Name, time.Now().String()+"_"+value).
			Err()
		queue.LastShardIndex++
		if err == nil {
			break
		}
	}
	return err
}

func (queue *DQueue) Clear() {
	for {
		v, e := queue.Pull()
		if e != nil {
			fmt.Printf("End clearing with error: %s\n", e.Error())
			return
		} else {
			fmt.Printf("value %s\n", v)
		}
	}
}

/*
 * Читаем из очереди
 *
 * Из того шарда, в котором самое раннее сообщение
 *
 */
func (queue *DQueue) Pull() (string, error) {

	ctx := context.Background()
	var minClient *redis.Client
	minVal := ""
	minTime := time.Now().String()

	for i := 0; i < queue.ShardCount; i++ {
		configData.redisOptions.Addr = (*queue.Hosts)[i]
		client := redis.NewClient(configData.redisOptions)
		_, errPing := client.Ping(ctx).Result()
		if errPing != nil {
			continue
		}
		value, errLIndex := client.LIndex(ctx, queue.Name, 0).Result()
		if errLIndex != nil {
			if strings.HasPrefix(errLIndex.Error(), "MOVED") {
				continue
			}
			return "", errLIndex
		}
		nodeInfo := strings.Split(fmt.Sprintf("%v", value), "_")
		if nodeInfo[0] < minTime {
			if client.LPop(ctx, queue.Name).Err() == nil {
				if minClient != nil {
					minClient.LPush(ctx, queue.Name, minTime+"_"+minVal)
				}
				minTime = nodeInfo[0]
				minVal = nodeInfo[1]
				minClient = client
			}
		}
	}
	if minVal == "" {
		return "", fmt.Errorf("cannot connect")
	} else {
		return minVal, nil
	}
}
