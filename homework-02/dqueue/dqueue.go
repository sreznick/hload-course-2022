package dqueue

import (
    "fmt"
    "errors"
    "context"
    "encoding/json"
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
  // immutable
  Name string
  RedisHosts []string

  // mutable
  NextPushHostIndex int
  NextPullHostIndex int

  lockSequenceNumber int // distributed lock sequence number, should NOT be stored in znode
}

var root = "/yakurbatov"
var ctx = context.Background()
var zkConnection *zk.Conn

/*
 * Запомнить данные и везде использовать
 */
func Config(redisHosts []string, zkCluster []string) {
  ctx = context.WithValue(ctx, "redis-hosts", redisHosts)
	ctx = context.WithValue(ctx, "zk-hosts", zkCluster)
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
    var nodeData = getQueueNodeData(name)
    var dqueue = DQueue{}

    redisHostsRaw := ctx.Value("redis-hosts")

    redisHosts, ok := redisHostsRaw.([]string)
    if !ok {
      panic("Invalid type")
    }

    if (nShards > len(redisHosts)) {
      return DQueue{}, errors.New(fmt.Sprintf("Limit of shards count is %d, provided value is %d", len(redisHosts), nShards))
    }

    if nodeData != nil {
      json.Unmarshal(nodeData, &dqueue)

      if len(dqueue.RedisHosts) != nShards {
        return DQueue{}, errors.New(fmt.Sprintf("An instance of queue with name %s already exists, but shards count differs", name))
      }
    } else {
      dqueue = DQueue{name, chooseRedisShards(nShards), 0, 0, -1}
      createDQueueNode(dqueue)
    }

    return dqueue, nil
}

// It would be better to select shards randomly to get better distribution,
// however current approach is OK too
func chooseRedisShards(count int) []string {
  redisHostsRaw := ctx.Value("redis-hosts")

  redisHosts, ok := redisHostsRaw.([]string)
  if !ok {
    panic("Invalid type")
  }

  return redisHosts[0:count]
}

/*
 * Пишем в очередь. Каждый следующий Push - в следующий шард
 *
 * Если шард упал - пропускаем шард, пишем в следующий по очереди
 */
func (q *DQueue) Push(value string) error {
    lock(q)
    err := tryPushValueInRedis(q, value)
    unlock(q)

    return err
}

// iterate over shards untill one of them allows successful push
func tryPushValueInRedis(queue *DQueue, value string) error {
  var index = queue.NextPushHostIndex
  var epoch = 0

  for {
    epoch += 1
    if epoch > 1000 { return errors.New("RetriesLimitExceeded") }

    redisOptions := redis.Options{
      Addr:     queue.RedisHosts[index],
      Password: "",
      DB:       0,
    }
    client := redis.NewClient(&redisOptions)
    _, err := client.RPush(ctx, "yakurbatov" + ":" + queue.Name, value).Result()

    if err == nil {
      queue.NextPushHostIndex = (index + 1) % len(queue.RedisHosts)
      updateDQueueNode(queue)
      return nil
    } else {
      if (isMovedError(err.Error())) {
        var redirectedHost = extractRedisHostFromMovedError(err.Error())
        redisOptions := redis.Options{
          Addr:     redirectedHost,
          Password: "",
          DB:       0,
        }
        client := redis.NewClient(&redisOptions)
        _, err := client.RPush(ctx, "yakurbatov" + ":" + queue.Name, value).Result()
        if err != nil {
          continue
        }
        queue.NextPushHostIndex = (index + 1) % len(queue.RedisHosts)
        updateDQueueNode(queue)

        return nil
      } else {
        index = (index + 1) % len(queue.RedisHosts)
      }
    }
  }
}

/*
 * Читаем из очереди
 *
 * Из того шарда, в котором самое раннее сообщение
 *
 */
func (q *DQueue) Pull() (string, error) {
    lock(q)
    result, err := tryPullValueFromRedis(q)
    unlock(q)

    return result, err
}

// iterate over shards untill one of them allows successful pull
func tryPullValueFromRedis(queue *DQueue) (string, error) {
  var index = queue.NextPushHostIndex
  var epoch = -1;

  for {
    epoch += 1
    if epoch > 1000 { return "", errors.New("RetriesLimitExceeded") }

    redisOptions := redis.Options{
      Addr:     queue.RedisHosts[index],
      Password: "",
      DB:       0,
    }
    client := redis.NewClient(&redisOptions)
    val, err := client.LPop(ctx, "yakurbatov" + ":" + queue.Name).Result()

    if err == nil {
      queue.NextPushHostIndex = (index + 1) % len(queue.RedisHosts)
      updateDQueueNode(queue)
      return val, nil
    } else {
      if (isMovedError(err.Error())) {
        var redirectedHost = extractRedisHostFromMovedError(err.Error())
        redisOptions := redis.Options{
          Addr:     redirectedHost,
          Password: "",
          DB:       0,
        }
        client := redis.NewClient(&redisOptions)
        val, err := client.LPop(ctx, "yakurbatov" + ":" + queue.Name).Result()
        if err != nil && err.Error() == "redis: nil" {
          return "", errors.New("No value to pop")
        }
        if err != nil {
          continue
        }
        queue.NextPushHostIndex = (index + 1) % len(queue.RedisHosts)
        updateDQueueNode(queue)
        return val, nil
      } else {
        index = (index + 1) % len(queue.RedisHosts)
      }
    }
  }
}
