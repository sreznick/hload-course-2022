package dqueue

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
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
	name           string
	nShards        int
	hosts          *[]string
	questionsCount int
	lock           bool //чтобы клиенты не могли параллельно делать pull/push в одну и ту же очередь
}

var cfgs = Configs{}
var dqueues = make(map[string]*DQueue)

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
func Open(name string, nShards int, addresses *[]string) (DQueue, error) {
	key := name
	value, found := dqueues[key]
	if found {
		if (*value).nShards != nShards {
			panic("Dequeue alredy exists on different hosts")
		} else {
			return *value, nil
		}

	}
	d := DQueue{name: name, nShards: nShards, hosts: addresses, questionsCount: 0, lock: false}
	fmt.Println(*addresses)
	dqueues[key] = &d
	return d, nil

}

/*
 * Пишем в очередь. Каждый следующий Push - в следующий шард
 *
 * Если шард упал - пропускаем шард, пишем в следующий по очереди
 */
func (d *DQueue) Push(value string) error {
	if d.lock {
		return errors.New("resource is currently locked; wait for a while")
	}

	var err error
	tries := 0
	d.questionsCount++
	d.lock = true

	for (err != nil || tries == 0) && tries < d.nShards {
		cfgs.redisOptions.Addr = (*d.hosts)[(d.questionsCount+tries)%d.nShards]
		//fmt.Println(cfgs.redisOptions)
		rdb := redis.NewClient(cfgs.redisOptions)
		err = rdb.Do(cfgs.ctx, "rpush", "aisakova", time.Now().String()+"_"+value).Err()
		tries++

	}
	d.lock = false
	return err
}

/*
 * Читаем из очереди
 *
 * Из того шарда, в котором самое раннее сообщение
 *
 */
func (d *DQueue) Pull() (string, error) {
	if d.lock {
		return "", errors.New("resource is currently locked; wait for a while")
	}

	var err error
	var minTimeClient *redis.Client
	minValue := ""
	minTime := time.Now().String()
	d.lock = true

	for i := 0; i < d.nShards; i++ {
		cfgs.redisOptions.Addr = (*d.hosts)[i]
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
	d.lock = false
	if minValue == "" {
		return minValue, err
	} else {
		return minValue, nil
	}

}
