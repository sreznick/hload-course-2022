package dqueue

import (
	//    "fmt"
	//    "context"

	"context"
	"encoding/json"
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
	name    string
	nShards int
	curIdx  int
}

/*
 * Запомнить данные и везде использовать
 */
type Options struct {
	ctx           context.Context
	redis         *redis.Options
	redisAddreses []string
	prefix        string
	zk            []string
}

var opt = Options{}

func Config(redisOptions *redis.Options, zkCluster []string, redisAddr []string) {
	opt.ctx = context.Background()
	opt.redis = redisOptions
	opt.redisAddreses = redisAddr
	opt.prefix = "amozhaev"
	opt.zk = zkCluster
}

func tilt(err error) {
	if err != nil {
		panic(err)
	}
}

/*
 * Открываем очередь на nShards шардах
 *
 * Попытка создать очередь с существующим именем и другим количеством шардов
 * должна приводить к ошибке
 *
 * При попытке создать очередь с существующим именем и тем же количеством шардов
 * нужно вернуть экземпляр DQueue, позволяющий делать Push/Pull
 *
 * Предыдущее открытие может быть совершено другим клиентом, соединенным с любым узлом
 * Redis-кластера
 *
 * Отдельные узлы Redis-кластера могут выпадать. Availability очереди в целом
 * не должна от этого страдать
 *
 */
func Open(name string, nShards int) (DQueue, error) {
	c, _, err := zk.Connect(opt.zk, time.Second) //*10)
	tilt(err)

	nShardsPath := "/" + opt.prefix + "/" + name + "/" + strconv.Itoa(nShards)
	exists, _, err := c.Exists(nShardsPath)
	tilt(err)
	if exists {
		var deq DQueue
		bytes, _, err := c.Get(nShardsPath)
		tilt(err)
		err = json.Unmarshal(bytes, &deq)
		tilt(err)
		return deq, nil
	}

	deqPath := "/" + opt.prefix + "/" + name
	exists, _, err = c.Exists(deqPath)
	tilt(err)
	if exists {
		return DQueue{}, fmt.Errorf("DQueue already exists with different number of shards")
	}

	deq := DQueue{name: name, nShards: nShards, curIdx: 0}
	_, err = c.Create(deqPath, []byte{}, 0, zk.WorldACL(zk.PermAll))
	tilt(err)
	_, err = c.Create(nShardsPath, []byte{}, 0, zk.WorldACL(zk.PermAll))
	tilt(err)

	bytes, err := json.Marshal(deq)
	tilt(err)
	_, err = c.Set(nShardsPath, bytes, 0)
	tilt(err)
	return deq, nil
}

/*
 * Пишем в очередь. Каждый следующий Push - в следующий шард
 *
 * Если шард упал - пропускаем шард, пишем в следующий по очереди
 */

func (deq *DQueue) lock(c *zk.Conn) error {
	lockPath := "/" + opt.prefix + "/" + deq.name + "/" + strconv.Itoa(deq.nShards) + "/" + "lock"
	for {
		exists, _, err := c.Exists(lockPath)
		tilt(err)

		if !exists {
			_, err = c.Create(lockPath, []byte{}, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
			tilt(err)
			return nil
		}
		time.Sleep(1000)
	}
}

func (deq *DQueue) unlock(c *zk.Conn) error {
	lockPath := "/" + opt.prefix + "/" + deq.name + "/" + strconv.Itoa(deq.nShards) + "/" + "lock"
	err := c.Delete(lockPath, -1)
	return err
}

func (deq *DQueue) Push(value string) error {
	c, _, err := zk.Connect(opt.zk, time.Second) //*10)
	tilt(err)
	deq.lock(c)

	for i := 0; i < deq.nShards; i++ {
		deq.curIdx++
		deq.curIdx %= deq.nShards

		opt.redis.Addr = opt.redisAddreses[deq.curIdx]
		client := redis.NewClient(opt.redis)

		_, err := client.Ping(client.Context()).Result()
		if err != nil {
			continue
		}

		path := "/" + opt.prefix + "/" + deq.name
		val := time.Now().String() + "-" + value
		client.RPush(client.Context(), path, val)

		bytes, err := json.Marshal(deq)
		tilt(err)
		nShardsPath := path + "/" + strconv.Itoa(deq.nShards)
		_, err = c.Set(nShardsPath, bytes, 0)
		tilt(err)

		deq.unlock(c)
		return nil
	}
	return fmt.Errorf("couldn't connect")
}

/*
 * Читаем из очереди
 *
 * Из того шарда, в котором самое раннее сообщение
 *
 */
func (deq *DQueue) Pull() (string, error) {
	c, _, err := zk.Connect(opt.zk, time.Second) //*10)
	tilt(err)
	deq.lock(c)

	path := "/" + opt.prefix + "/" + deq.name
	best_time := time.Now().String()
	best_val := ""
	best_idx := 0
	best_exists := false
	for i := 0; i < deq.nShards; i++ {
		opt.redis.Addr = opt.redisAddreses[i]
		client := redis.NewClient(opt.redis)

		_, err := client.Ping(client.Context()).Result()
		if err != nil {
			continue
		}

		val, err := client.LIndex(client.Context(), path, 0).Result()
		tilt(err)

		value := strings.Split(fmt.Sprintf("%v", val), "-")
		timestamp := value[0]
		if best_time > timestamp {
			best_idx = i
			best_val = value[1]
			best_exists = true
		}
	}
	if best_exists {
		opt.redis.Addr = opt.redisAddreses[best_idx]
		client := redis.NewClient(opt.redis)

		client.LPop(client.Context(), path)
		deq.unlock(c)
		return best_val, nil
	}
	deq.unlock(c)
	return "", fmt.Errorf("couldn't connect")
}
