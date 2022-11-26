package dqueue

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/go-zookeeper/zk"
	"sort"
	"strconv"
	"time"
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
	name         string
	nShards      int
	currentIndex int
	prefix       string
}

var ctx = context.Background()

/*
 * Запомнить данные и везде использовать
 */
func Config(redisAddrs []string, zkCluster []string) {
	ctx = context.WithValue(ctx, "redisAddrs", redisAddrs)
	ctx = context.WithValue(ctx, "zkCluster", zkCluster)
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
	prefix := "tamarinvs"
	zkAddr := ctx.Value("zkCluster")
	switch v := zkAddr.(type) {
	case []string:
		c, _, err := zk.Connect(v, time.Second) //*10)
		if err != nil {
			panic(err)
		}
		createZkRoot(c, prefix, name)
		infoPath := fmt.Sprintf("/%s_info_%s", prefix, name)
		info, _, err := c.Get(infoPath)
		if err != nil {
			_, err := c.Create(infoPath, []byte(strconv.Itoa(nShards)), 0, zk.WorldACL(zk.PermAll))
			if err != nil {
				panic(err)
			}
		} else {
			infoShards, err := strconv.Atoi(string(info)[:])
			if err != nil {
				panic(err)
			}
			if infoShards != nShards {
				return DQueue{}, fmt.Errorf("invalid number of shards %d. Expected %d", nShards, infoShards)
			}
		}
	}
	return DQueue{name, nShards, 0, prefix}, nil
}

/*
 * Пишем в очередь. Каждый следующий Push - в следующий шард
 *
 * Если шард упал - пропускаем шард, пишем в следующий по очереди
 */
func (dq *DQueue) Push(value string) error {
	redisAddrs := ctx.Value("redisAddrs")
	switch v := redisAddrs.(type) {
	case []string:
		for {
			alive := checkRedisAddress(dq.currentIndex)
			if alive {
				redisOptions := redis.Options{
					Addr:     v[dq.currentIndex],
					Password: "",
					DB:       0,
				}
				client := redis.NewClient(&redisOptions)
				client.RPush(client.Context(), fmt.Sprintf("%s::%s", dq.prefix, dq.name), value)

				err := SaveToZk(dq.prefix, dq.name, strconv.Itoa(dq.currentIndex))
				return err
			} else {
				dq.currentIndex = (dq.currentIndex) % dq.nShards
			}
		}
	}
	return nil
}

/*
 * Читаем из очереди
 *
 * Из того шарда, в котором самое раннее сообщение
 *
 */
func (dq *DQueue) Pull() (string, error) {
	redisAddrs := ctx.Value("redisAddrs")
	switch v := redisAddrs.(type) {
	case []string:
		redisAddr := getNextAddress(dq.prefix, dq.name)
		redisOptions := redis.Options{
			Addr:     v[redisAddr],
			Password: "",
			DB:       0,
		}
		client := redis.NewClient(&redisOptions)
		return client.LPop(client.Context(), fmt.Sprintf("%s::%s", dq.prefix, dq.name)).Val(), nil
	}
	return "", nil
}

func SaveToZk(prefix string, queueName string, redisAddr string) error {
	zkAddr := ctx.Value("zkCluster")
	switch v := zkAddr.(type) {
	case []string:
		c, _, err := zk.Connect(v, time.Second) //*10)
		if err != nil {
			panic(err)
		}

		pathToRedis := fmt.Sprintf("/%s/%s", prefix, queueName)

		children, _, _, err := c.ChildrenW(pathToRedis)
		index := -1
		if err == nil {
			for _, child := range children {
				childId, errParse := strconv.Atoi(child)
				if errParse != nil {
					panic(errParse)
				}
				if -index > -childId {
					index = childId
				}
			}
		}
		index += 1

		path := fmt.Sprintf("/%s/%s/%d", prefix, queueName, index)
		_, errCreate := c.Create(path, []byte(redisAddr), 0, zk.WorldACL(zk.PermAll))
		if errCreate != nil {
			panic(errCreate)
		}
	}
	return nil
}

func createZkRoot(c *zk.Conn, prefix string, queueName string) {
	path1 := fmt.Sprintf("/%s/%s", prefix, queueName)
	path2 := fmt.Sprintf("/%s", prefix)

	exists, _, err := c.Exists(path2)
	if err != nil {
		panic(err)
	}
	if !exists {
		_, errCreate := c.Create(path2, []byte{}, 0, zk.WorldACL(zk.PermAll))
		if errCreate != nil {
			panic(errCreate)
		}
	}
	exists, _, err = c.Exists(path1)
	if err != nil {
		panic(err)
	}
	if !exists {
		_, errCreate := c.Create(path1, []byte{}, 0, zk.WorldACL(zk.PermAll))
		if errCreate != nil {
			panic(errCreate)
		}
	}
}

func getNextAddress(prefix string, queueName string) int {
	zkAddr := ctx.Value("zkCluster")
	switch v := zkAddr.(type) {
	case []string:
		c, _, err := zk.Connect(v, time.Second) //*10)
		if err != nil {
			panic(err)
		}

		createZkRoot(c, prefix, queueName)
		pathToRedis := fmt.Sprintf("/%s/%s", prefix, queueName)

		children, _, _, err := c.ChildrenW(pathToRedis)
		if err != nil {
			panic(err)
		}
		childrenIds := make([]int, len(children))
		for i, child := range children {
			childId, errParse := strconv.Atoi(child)
			if errParse != nil {
				panic(errParse)
			}
			childrenIds[i] = childId
		}
		childrenSlice := childrenIds[:]
		sort.Ints(childrenSlice)
		for _, child := range childrenIds {
			path := fmt.Sprintf("%s/%d", pathToRedis, child)
			redisAddrBytes, _, err2 := c.Get(path)
			if err2 != nil {
				panic(err2)
			}
			err := c.Delete(path, -1)
			if err != nil {
				return 0
			}
			redisAddr, _ := strconv.Atoi(string(redisAddrBytes[:]))
			if checkRedisAddress(redisAddr) {
				return redisAddr
			}
		}
	}
	return 0
}

func checkRedisAddress(addrIndex int) bool {
	zkAddr := ctx.Value("zkCluster")
	switch v := zkAddr.(type) {
	case []string:
		client := redis.NewClient(&redis.Options{
			Addr:     v[addrIndex],
			Password: "",
			DB:       0,
		})

		_, err := client.Ping(client.Context()).Result()
		return err != nil
	}
	return false
}
