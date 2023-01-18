package dqueue

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/go-zookeeper/zk"
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
	infoPrefix   string
	queuePrefix  string
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
		getOrCreateZkPath(c, []string{prefix, "queue", name})
		getOrCreateZkPath(c, []string{prefix, "info"})
		infoPath := fmt.Sprintf("/%s/info/%s", prefix, name)
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
	queuePrefix := prefix + "/queue"
	infoPrefix := prefix + "/info"
	return DQueue{name, nShards, 0, prefix, infoPrefix, queuePrefix}, nil
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

				err := dq.SaveToZk(strconv.Itoa(dq.currentIndex))
				return err
			} else {
				dq.currentIndex = dq.currentIndex % dq.nShards
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
		redisAddr, err := dq.getNextAddress()
		if err != nil {
			panic(err)
		}
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

/*
Check redis node
*/
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
