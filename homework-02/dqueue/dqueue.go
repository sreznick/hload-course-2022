package dqueue

import (
	"dqueue/models"
	"dqueue/redis"
	"dqueue/zookeeper"
	"errors"
	"time"

	"github.com/go-zookeeper/zk"

	redisLib "github.com/go-redis/redis/v8"
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
	name      string
	zkName    string
	redisName string
}

var (
	zooPrefix   string = "/mpereskokova/queues/"
	redisPrefix string = "mpereskokova:"
	ZkClient           = zookeeper.NewZookeeper()
	RedisClient        = redis.NewRedisClient()
)

/*
 * Запомнить данные и везде использовать
 */
func Config(redisOptions *redisLib.ClusterOptions, zkCluster []string) {
	ZkClient.Config(zkCluster)
	RedisClient.Configure(redisOptions)
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
func Open(name string, nShards int) (resQ DQueue, resErr error) {
	conn, err := ZkClient.Connect()
	if err != nil {
		return DQueue{}, err
	}

	zooPath := zooPrefix + name
	redisPath := redisPrefix + name
	lock, err := ZkClient.Lock(conn, zooPrefix, name)
	if err != nil {
		return DQueue{}, err
	}
	defer func() {
		err = ZkClient.UnLock(conn, lock)
		if err != nil {
			resErr = err
		}
	}()

	if queueInfo, err := ZkClient.Get(conn, zooPath); err != nil {
		if !errors.Is(err, zk.ErrNoNode) {
			return DQueue{}, err
		}
	} else {
		if len(queueInfo.Shards) != nShards {
			return DQueue{}, errors.New("Queue has a different number of shards")
		}
		return DQueue{
			zkName:    zooPath,
			redisName: redisPath,
		}, nil
	}

	shards, err := RedisClient.Create(redisPath, nShards)
	if err != nil {
		return DQueue{}, err
	}

	info := models.QueueInfo{
		Shards:  shards,
		Current: 0,
	}
	err = ZkClient.Create(conn, zooPath, info)
	if err != nil {
		return DQueue{}, err
	}

	return DQueue{
		name:      name,
		zkName:    zooPath,
		redisName: redisPath,
	}, nil
}

/*
 * Пишем в очередь. Каждый следующий Push - в следующий шард
 *
 * Если шард упал - пропускаем шард, пишем в следующий по очереди
 */
func (q *DQueue) Push(value string) (resErr error) {
	conn, err := ZkClient.Connect()
	if err != nil {
		return err
	}

	lock, err := ZkClient.Lock(conn, zooPrefix, q.name)
	if err != nil {
		return err
	}
	defer func() {
		err = ZkClient.UnLock(conn, lock)
		if err != nil {
			resErr = err
		}
	}()

	queueInfo, err := ZkClient.Get(conn, q.zkName)
	if err != nil {
		return err
	}

	attempts := 0
	for {
		attempts++

		err = RedisClient.Push(*queueInfo.Shards[queueInfo.Current], models.RedisValue{Value: value, Timestamp: time.Now().UnixNano()})
		queueInfo.Current = (queueInfo.Current + 1) % len(queueInfo.Shards)
		if err == nil {
			break
		} else if attempts >= len(queueInfo.Shards) {
			return errors.New("All shards unavailable")
		}
	}

	err = ZkClient.Update(conn, q.zkName, *queueInfo)
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
func (q *DQueue) Pull() (resAns string, resErr error) {
	conn, err := ZkClient.Connect()
	if err != nil {
		return "", err
	}

	lock, err := ZkClient.Lock(conn, zooPrefix, q.name)
	if err != nil {
		return "", err
	}
	defer func() {
		err = ZkClient.UnLock(conn, lock)
		if err != nil {
			resErr = err
		}
	}()

	queueInfo, err := ZkClient.Get(conn, q.zkName)
	if err != nil {
		return "", err
	}

	for {
		var bestShard *models.Host
		bestValue := &models.RedisValue{
			Timestamp: time.Now().UnixNano(),
		}

		for _, shard := range queueInfo.Shards {
			value, err := RedisClient.Front(shard)
			if err == nil {
				if value.Timestamp < bestValue.Timestamp {
					bestValue = value
					bestShard = shard
				}
			}
		}
		if bestShard == nil {
			return "", models.QueueIsEmptyErr
		}

		err = RedisClient.Pop(*bestShard)
		if err != nil {
			continue
		}

		return bestValue.Value, nil
	}
}
