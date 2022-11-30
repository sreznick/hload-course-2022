package dqueue

import (
//    "fmt"
    "context"
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
 type DQueue struct {
	name string
}

var ctx = context.Background()
var currentRedisOptions = redis.ClusterOptions {
    Addrs: []string{":7000", ":7001"},
}
var dQueue = make(map[string]DQueue)
var pushNumber = 0  // % (node count) gives us a shard to push


/*
 * Запомнить данные и везде использовать
 */
func Config(redisOptions *redis.ClusterOptions) {
    currentRedisOptions = redisOptions
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
    rdb := redis.NewClusterClient(currentRedisOptions)
    real_name := "aalexl:" + name + "_params"
    exists, err := rdb.Exists(ctx, real_name).Result()
    if err != nil {
        if err != redis.Nil {
            panic(err)
        }
    } else {
        if (exists == 1) { // if queue name alredy exists
            shards, err := redis.Strings(rdb.LRange(ctx, real_name, 0, -1))
            if err != nil {
                panic(err)
            }
            if (len(shards) == nShards) {
                dQueue[real_name] = DQueue{name: real_name}
                return dQueue[real_name], nil
            } else {
                panic("Name exists but shards count is different!")
            }
        }
    }
    dQueue[real_name] = DQueue{name: real_name}
    newRedisOptions := currentRedisOptions
    newRedisOptions.Addrs = newRedisOptions.Addrs[:nShards]
    rdb = redis.NewClusterClient(newRedisOptions)
    err = rdb.ForEachShard(ctx, func(ctx context.Context, shard *redis.Client) error {
        return shard.LPush(ctx, real_name, newRedisOptions.Addrs).Err()
    })
    if err != nil {
        panic(err)
    }
    dQueue[real_name] = DQueue{name: real_name}
    return dQueue[real_name], nil
}

/*
 * Пишем в очередь. Каждый следующий Push - в следующий шард
 * 
 * Если шард упал - пропускаем шард, пишем в следующий по очереди
 */
func push_internal(name string, shards []string, time_value string, try_number int) error {
    list_name := "aalexl:" + name + "_que"
    pushNumber++
    if (len(shards) * 2 < try_number) {
        panic("Tried to push in each shard 2 times. Failed all of them!")
    }
    rdb_shard_ip := shards[pushNumber % len(shards)]
    newRedisOptions := currentRedisOptions
    newRedisOptions.Addrs = []string{rdb_shard_ip}
    rdb_shard := redis.NewClusterClient(newRedisOptions)
    err := rdb_shard.RPush(ctx, list_name, time_value)
    if err != nil {
        push_internal(name, shards, time_value, try_number)
    }
    return nil
}


func (self *DQueue) Push(value string) error {
    rdb := redis.NewClusterClient(currentRedisOptions)
    real_name := "aalexl:" + self.name + "_params"
    exists, err := rdb.Exists(ctx, real_name).Result()
    if err != nil {
        if err != redis.Nil {
            panic(err)
        } else {
            panic("Que was not found!")
        }
    } else {
        if (exists == 1) { // if queue name alredy exists
            shards, err := redis.Strings(rdb.LRange(ctx, real_name, 0, -1))
            if err != nil {
                panic(err)
            }
            return push_internal(self.name, shards, string(time.Now().UnixNano()) + ' ' + value, 0)
        } else {
            panic("Que was not found or there is more then one que with this name!")
        }
    }
}

/*
 * Читаем из очереди
 *
 * Из того шарда, в котором самое раннее сообщение
 * 
 */
func (*DQueue) Pull() (string, error) {
    rdb := redis.NewClusterClient(currentRedisOptions)
    real_name := "aalexl:" + self.name + "_params"
    list_name := "aalexl:" + self.name + "_que"
    exists, err := rdb.Exists(real_name).Result()
    if err != nil {
        if err != redis.Nil {
            panic(err)
        } else {
            panic("Que was not found!")
        }
    } else {
        if (exists == 1) { // if queue name alredy exists
            shards, err := redis.Strings(rdb.LRange(key, 0, -1))
            if err != nil {
                panic(err)
            }
            best_ans = nil
            best_time = 0
            err := rdb.ForEachShard(ctx, func(ctx context.Context, shard *redis.Client) error {
                time_value, err := shard.Lpop(ctx, list_name, newRedisOptions.Addrs)
                if err != nil {
                    panic(err)
                }
                current_time_str, value = strings.Cut(time_value, " ")
                current_time := int(current_time_str)
                if best_ans == nil {
                    best_ans = value
                    best_time = current_time
                } else {
                    if best_time < current_time {
                        return shard.LPush(ctx, list_name, string(best_time) + " " + best_ans).Err()
                        best_ans = value
                        best_time = current_time
                    } else {
                        return shard.LPush(ctx, list_name, time_value).Err()
                    }
                }
            })
        } else {
            panic("Que was not found or there is more then one que with this name!")
        }
    }
}