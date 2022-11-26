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
type Configs struct {
	redisOptions *redis.Options
	ctx          context.Context
}

type DQueue struct {
	Name           string
	NShards        int
	Hosts          *[]string
	QuestionsCount int
	Lock           bool //чтобы клиенты не могли параллельно делать pull/push в одну и ту же очередь
}

var cfgs = Configs{}

/*
 * Запомнить данные и везде использовать
 */
func Config(redisOptions *redis.Options) {
	cfgs.redisOptions = redisOptions
	cfgs.ctx = context.Background()
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
func Open(name string, nShards int, addresses *[]string, c *zk.Conn) (DQueue, error) {
	nShardsPath := "/zookeeper/aisakova" + "/" + name + "/" + strconv.Itoa(nShards)
	dequeuePath := "/zookeeper/aisakova" + "/" + name
	check, _, err := c.Exists(nShardsPath)
	if check {
		var d DQueue
		rawData, _, err := c.Get(nShardsPath)
		err = json.Unmarshal(rawData, &d)
		return d, err
	}
	check, _, err = c.Exists(dequeuePath)
	if check {
		return DQueue{}, errors.New("Dequeue alredy exists on different hosts")
	}
	d := DQueue{Name: name, NShards: nShards, Hosts: addresses, QuestionsCount: 0, Lock: false}
	_, err = c.Create(dequeuePath, []byte("dequeueFolder"), 0, zk.WorldACL(zk.PermAll))
	_, err = c.Create(nShardsPath, []byte("nShardsFolder"), 0, zk.WorldACL(zk.PermAll))
	marshalDequeue, err := json.Marshal(d)
	_, err = c.Set(nShardsPath, marshalDequeue, 0)
	return d, err
}

/*
 * Пишем в очередь. Каждый следующий Push - в следующий шард
 *
 * Если шард упал - пропускаем шард, пишем в следующий по очереди
 */
func (d *DQueue) Push(value string) error {
	if d.Lock {
		return errors.New("resource is currently locked; wait for a while")
	}

	var err error
	tries := 0
	d.QuestionsCount++
	d.Lock = true

	for (err != nil || tries == 0) && tries < d.NShards {
		cfgs.redisOptions.Addr = (*d.Hosts)[(d.QuestionsCount+tries)%d.NShards]
		rdb := redis.NewClient(cfgs.redisOptions)
		err = rdb.Do(cfgs.ctx, "rpush", "aisakova", time.Now().String()+"_"+value).Err()
		tries++

	}
	d.Lock = false
	return err
}

/*
 * Читаем из очереди
 *
 * Из того шарда, в котором самое раннее сообщение
 *
 */
func (d *DQueue) Pull() (string, error) {
	if d.Lock {
		return "", errors.New("resource is currently locked; wait for a while")
	}

	var err error
	var minTimeClient *redis.Client
	minValue := ""
	minTime := time.Now().String()
	d.Lock = true

	for i := 0; i < d.NShards; i++ {
		cfgs.redisOptions.Addr = (*d.Hosts)[i]
		rdb := redis.NewClient(cfgs.redisOptions)
		value, getErr := rdb.Do(cfgs.ctx, "lindex", "aisakova", 0).Result()
		if getErr == nil {
			valueString := fmt.Sprintf("%v", value)
			timestamp := strings.Split(valueString, "_")[0]
			//Временнные метки состоят из цифр, недостающие разряды заполняются нулями. Поэтому можем сравнивать, как строки
			if timestamp < minTime {
				_, popErr := rdb.Do(cfgs.ctx, "lpop", "aisakova", 1).Result()
				if popErr == nil {
					if minTimeClient != nil {
						minTimeClient.Do(cfgs.ctx, "lpush", "aisakova", minTime+"_"+minValue)

					}
					minTime = timestamp
					minValue = strings.Split(valueString, "_")[1]
					minTimeClient = rdb

				} else {
					err = popErr
				}

			} else {
				err = getErr
			}
		}

	}
	d.Lock = false
	if minValue == "" {
		return minValue, err
	} else {
		return minValue, nil
	}

}
