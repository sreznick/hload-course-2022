package dqueue

import (
	//    "fmt"
	//    "context"

	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/go-zookeeper/zk"
	"strconv"
	"strings"
	"time"
	//    "github.com/go-zookeeper/zk"
)

var cluster RedisCluster
var _zkHosts []string

const (
	queuesPath = "/dkulikov/queues"
	letters    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type LocalQueue struct {
	Name string
	Host string
}

/*
 * Можно создавать несколько очередей
 *
 * Для клиента они различаются именами
 *
 * В реализации они могут потребовать вспомогательных данных
 * Для них - эта структура. Можете определить в ней любые поля
 */
type DQueue struct {
	Name       string
	NShards    int
	RedisHosts []LocalQueue

	CurIdPush int
	CurIdPop  int
	Version   int32
}

/*
 * Запомнить данные и везде использовать
 */
func Config(redisData RedisCluster, zkCluster []string) {
	cluster = redisData
	_zkHosts = zkCluster
}

/*
 * Открываем очередь на NShards шардах
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
	c, _, err := zk.Connect(_zkHosts, time.Second*3)
	if err != nil {
		return DQueue{}, err
	}
	defer c.Close()

	s, q, err := handleExisting(c, name, nShards)
	if err != nil {
		return DQueue{}, err
	}
	if s {
		return q, nil
	}

	// Locks zk
	err = Lock(c, queuesPath)
	if err != nil {
		return DQueue{}, err
	}

	// If created after lock
	s, q, err = handleExisting(c, name, nShards)
	if err != nil {
		return DQueue{}, err
	}
	if s {
		return q, nil
	}

	queuePath := queuesPath + "/" + name
	return createQueueZKNode(c, queuePath, nShards, name)
}

func handleExisting(c *zk.Conn, name string, nShards int) (bool, DQueue, error) {
	// Getting info about queue if exists
	queuePath := queuesPath + "/" + name
	ex, _, err := c.Exists(queuePath)
	if err != nil {
		return false, DQueue{}, err
	}
	if ex {
		q, err := getQueueInfo(c, queuePath)
		if err != nil {
			return true, DQueue{}, err
		}

		if q.NShards != nShards {
			return true, DQueue{}, fmt.Errorf("Conflict nshards with existing queue")
		}

		return true, q, nil
	}

	return false, DQueue{}, nil
}

func getQueueInfo(c *zk.Conn, queuePath string) (DQueue, error) {
	s, _, err := c.Get(queuePath)
	if err != nil {
		return DQueue{}, err
	}
	q := DQueue{}
	err = json.Unmarshal(s, &q)
	if err != nil {
		return DQueue{}, err
	}

	return q, nil
}

func getAllStringsLengthKBack(k int, s string) []string {
	strings := []string{}
	if k == 0 {
		return []string{s}
	}

	for i := 0; i < len(letters); i++ {
		strings = append(strings, getAllStringsLengthKBack(k-1, s+string(letters[i]))...)
	}

	return strings
}

func getAllStringsLengthK(k int) []string {
	return getAllStringsLengthKBack(k, "")
}

func contains(rs []LocalQueue, r redis.Options) bool {
	for _, rr := range rs {
		if rr.Host == r.Addr {
			return true
		}
	}

	return false
}

func selectHosts(n int, outerName string) ([]LocalQueue, error) {
	length := 1
	var hosts []LocalQueue
	for {
		strings := getAllStringsLengthK(length)
		for _, s := range strings {
			s = "dkulikov:" + outerName + ":" + s
			host, err := cluster.get_key_host(s)
			if err != nil {
				continue
			}

			if contains(hosts, host) {
				continue
			}
			hosts = append(hosts, LocalQueue{Host: host.Addr, Name: s})
			if len(hosts) == n {
				return hosts, nil
			}
		}
		length++
		if length >= 10000 {
			return hosts, fmt.Errorf("No suitable hosts")
		}
	}
}

func createQueueZKNode(c *zk.Conn, queuePath string, nShards int, name string) (DQueue, error) {
	redisHosts, err := selectHosts(nShards, name)
	if err != nil {
		return DQueue{}, nil
	}
	q := DQueue{
		NShards:    nShards,
		Name:       name,
		RedisHosts: redisHosts,

		CurIdPush: 0,
		CurIdPop:  0,
		Version:   0,
	}

	js, err := json.Marshal(q)
	if err != nil {
		return DQueue{}, err
	}
	_, err = c.Create(queuePath, js, 0, zk.WorldACL(zk.PermAll))
	if err != nil {
		return DQueue{}, err
	}

	_, err = c.Create(queuePath+"/_locknode", []byte{}, 0, zk.WorldACL(zk.PermAll))

	return q, nil
}

/*
 * Пишем в очередь. Каждый следующий Push - в следующий шард
 *
 * Если шард упал - пропускаем шард, пишем в следующий по очереди
 */
func (q *DQueue) Push(value string) error {
	c, _, err := zk.Connect(_zkHosts, time.Second*3)
	if err != nil {
		return err
	}
	defer c.Close()

	queuePath := queuesPath + "/" + q.Name
	err = Lock(c, queuePath)
	if err != nil {
		return err
	}

	s, st, err := c.Get(queuePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(s, q)
	if err != nil {
		return err
	}
	q.Version = st.Version

	for {
		redisQueue := q.RedisHosts[q.CurIdPush%q.NShards]
		ro := redis.Options{Addr: redisQueue.Host, Password: "", DB: 0}
		q.CurIdPush++

		ctx := context.Background()
		r := redis.NewClient(&ro)

		_, err = r.RPush(ctx, redisQueue.Name, fmt.Sprintf("%s|%d", value, q.CurIdPush)).Result()
		if err != nil {
			continue
		}
		break
	}

	self, err := json.Marshal(q)
	if err != nil {
		return err
	}

	_, err = c.Set(queuePath, self, q.Version)
	if err != nil {
		return err
	}

	return nil
}

/*
 * Читаем из очереди
 *
 * Из того шарда, в котором самое раннее сообщение
 *
 */
func (q *DQueue) Pull() (string, error) {
	c, _, err := zk.Connect(_zkHosts, time.Second*3)
	if err != nil {
		return "", err
	}
	defer c.Close()

	queuePath := queuesPath + "/" + q.Name
	err = Lock(c, queuePath)
	if err != nil {
		return "", err
	}

	s, st, err := c.Get(queuePath)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(s, q)
	if err != nil {
		return "", err
	}
	q.Version = st.Version

	for {
		redisQueue := q.RedisHosts[q.CurIdPop%q.NShards]
		ro := redis.Options{Addr: redisQueue.Host, Password: "", DB: 0}

		ctx := context.Background()
		r := redis.NewClient(&ro)

		v, err := r.LIndex(ctx, redisQueue.Name, 0).Result()
		if err != nil {
			q.CurIdPop++
			continue
		}

		strs := strings.Split(v, "|")
		id, err := strconv.Atoi(strs[1])
		v = strs[0]
		if err != nil {
			q.CurIdPop++
			continue
		}

		if id > q.CurIdPop {
			q.CurIdPop++
			continue
		}

		_, err = r.LPop(ctx, redisQueue.Name).Result()
		if err != nil {
			q.CurIdPop++
			continue
		}

		self, err := json.Marshal(q)
		if err != nil {
			return "", err
		}

		_, err = c.Set(queuePath, self, q.Version)
		if err != nil {
			return "", err
		}

		return v, nil
	}
}
