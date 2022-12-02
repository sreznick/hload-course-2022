package dqueue

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/go-zookeeper/zk"
	"sort"
	"strconv"
	"time"
)

const prefix = "bmadzhuga"
const limit = 1000

type DQueue struct {
	name    string
	nShards int
}

var config map[string][]string

func Config(redisCluster []string, zkCluster []string) {
	config = make(map[string][]string, 2)
	config["redisCluster"] = redisCluster
	config["zkCluster"] = zkCluster
}

func createZNode(connect *zk.Conn, infoPath string, data []byte) error {
	_, err := connect.Create(
		infoPath,
		data,
		0,
		zk.WorldACL(zk.PermAll),
	)

	if err != nil {
		return fmt.Errorf("Zookeeper node creation failed: %w", err)
	}

	return nil
}

func getIndex(connect *zk.Conn, queueName string) (int, error) {
	pathToRedis := fmt.Sprintf("/%s/%s", prefix, queueName)
	children, _, _, err := connect.ChildrenW(pathToRedis)
	if err != nil {
		return 0, nil
	}

	var (
		index   = -1
		childID int
	)
	for _, child := range children {
		childID, err = strconv.Atoi(child)
		if err != nil {
			return 0, fmt.Errorf("Can't parse childID to int: %w", err)
		}

		if index <= childID {
			index = childID
		}
	}

	return index + 1, nil
}

func createZNodeWithIndex(queueName string, redisAddr string) error {
	zkCluster, ok := config["zkCluster"]
	if !ok {
		return fmt.Errorf("Zookeeper cluster isn't found")
	}

	connect, _, err := zk.Connect(zkCluster, time.Second)
	if err != nil {
		return fmt.Errorf("Zookeeper connection failed: %w", err)
	}

	index, err := getIndex(connect, queueName)
	if err != nil {
		return fmt.Errorf("Getting index failed: %w", err)
	}

	path := fmt.Sprintf("/%s/%s/%d", prefix, queueName, index)
	return createZNode(connect, path, []byte(redisAddr))
}

func createZNodeRoot(connect *zk.Conn, queueName string) error {
	err := checkExistsAndCreateZNode(connect, fmt.Sprintf("/%s", prefix))
	if err != nil {
		return fmt.Errorf("Checking node failed: %w", err)
	}

	err = checkExistsAndCreateZNode(connect, fmt.Sprintf("/%s/%s", prefix, queueName))
	if err != nil {
		return fmt.Errorf("Checking and creating node failed: %w", err)
	}

	return nil
}

func checkExistsAndCreateZNode(connect *zk.Conn, path string) error {
	exists, _, err := connect.Exists(path)
	if err != nil {
		return fmt.Errorf("Path %v doesn't exist: %w", path, err)
	}

	if exists {
		return nil
	}

	return createZNode(connect, path, []byte{})
}

func checkAddrToRedis(addrIndex int) error {
	redisAddr, ok := config["redisCluster"]
	if !ok {
		return errors.New("Can't read Redis config")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr[addrIndex],
		Password: "",
		DB:       0,
	})

	_, err := client.Ping(client.Context()).Result()

	return err
}

func nextAddress(queueName string) (int, error) {
	zkAddr, ok := config["zkCluster"]
	if !ok {
		return 0, fmt.Errorf("Zookeeper cluster isn't found")
	}

	connect, _, err := zk.Connect(zkAddr, time.Second)
	if err != nil {
		return 0, fmt.Errorf("Zookeeper connect failed: %w", err)
	}

	err = createZNodeRoot(connect, queueName)
	if err != nil {
		return 0, fmt.Errorf("Creation Zookeeper node failed: %w", err)
	}

	pathToRedis := fmt.Sprintf("/%s/%s", prefix, queueName)
	children, _, _, err := connect.ChildrenW(pathToRedis)
	if err != nil {
		return 0, fmt.Errorf("Connection children failed: %w", err)
	}

	var (
		childID     int
		childrenIDs = make([]int, 0, len(children))
	)
	for _, child := range children {
		childID, err = strconv.Atoi(child)
		if err != nil {
			return 0, fmt.Errorf("Can't parse childID to id: %w", err)
		}

		childrenIDs = append(childrenIDs, childID)
	}
	sort.Ints(childrenIDs)

	var (
		redisBytes  []byte
		addrToRedis int
	)
	for _, child := range childrenIDs {
		path := fmt.Sprintf("%s/%d", pathToRedis, child)

		redisBytes, _, err = connect.Get(path)
		if err != nil {
			return 0, fmt.Errorf("Connection get failed: %w", err)
		}

		err = connect.Delete(path, -1)
		if err != nil {
			return 0, fmt.Errorf("Connection delete failed: %w", err)
		}

		addrToRedis, err = strconv.Atoi(string(redisBytes))
		if err != nil {
			return 0, fmt.Errorf("Can't parse redisBytes to int: %w", err)
		}

		err := checkAddrToRedis(addrToRedis)
		if err == nil {
			return addrToRedis, nil
		}
	}

	return 0, nil

}

func Open(name string, nShards int) (DQueue, error) {

	dQueue := DQueue{
		name:    name,
		nShards: nShards,
	}

	zkCluster, ok := config["zkCluster"]
	if !ok {
		return DQueue{}, fmt.Errorf("Can't find Zookeeper cluster in config")
	}

	connect, _, err := zk.Connect(zkCluster, time.Second)
	if err != nil {
		return DQueue{}, fmt.Errorf("Zookeeper connection failed: %w", err)
	}

	if err = createZNodeRoot(connect, name); err != nil {
		return DQueue{}, fmt.Errorf("Zookeeper root node creation failed: %w", err)
	}

	infoPath := fmt.Sprintf("/%v_info_%v", prefix, name)
	infoBytes, _, err := connect.Get(infoPath)

	if err != nil {
		return dQueue, createZNode(connect, infoPath, []byte(fmt.Sprintf("%v", nShards)))
	}

	infoShards, err := strconv.Atoi(string(infoBytes))
	if err != nil {
		return DQueue{}, fmt.Errorf("Convertation bytes to int failed: %w", err)
	}

	if infoShards != nShards {
		return DQueue{}, fmt.Errorf("Info shards amount (%d) doesn't equal to shards amount (%d)", nShards, infoShards)
	}

	return dQueue, nil
}

func (d *DQueue) Push(value string) error {
	redisCluster, ok := config["redisCluster"]
	if !ok {
		return fmt.Errorf("Can't find redis cluster")
	}

	counter := 0

	zkKey := fmt.Sprintf("%s/%s", d.name, "push")

	clusterIndex, err := nextAddress(zkKey)
	clusterIndex = clusterIndex % d.nShards

	if err != nil {
		return fmt.Errorf("Can't get current index: %w", err)
	}

	for {
		err := checkAddrToRedis(clusterIndex)
		if err != nil {
			counter++

			if counter > limit {
				return err
			}

			continue
		}

		addr := redisCluster[clusterIndex]

		redisOptions := redis.Options{
			Addr:     addr,
			Password: "",
			DB:       0,
		}

		key, err := d.findKeyForAddr(addr)

		if err != nil {
			return fmt.Errorf("Can't find key for slot: %w", err)
		}

		fmt.Println(addr)
		client := redis.NewClient(&redisOptions)
		client.RPush(client.Context(), key, value)

		return createZNodeWithIndex(zkKey, strconv.Itoa((clusterIndex+1)%d.nShards))
	}
}

func (d *DQueue) Pull() (string, error) {
	redisCluster, ok := config["redisCluster"]
	if !ok {
		return "", fmt.Errorf("Redis cluster isn't found")
	}

	zkKey := fmt.Sprintf("%s/%s", d.name, "pull")

	clusterIndex, err := nextAddress(zkKey)
	clusterIndex = clusterIndex % d.nShards

	if err != nil {
		return "", fmt.Errorf("Getting next adress failed: %w", err)
	}

	addr := redisCluster[clusterIndex]

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	key, err := d.findKeyForAddr(addr)

	if err != nil {
		return "", fmt.Errorf("Can't find key for slot: %w", err)
	}

	res, err := client.LPop(client.Context(), key).Result()

	if err != nil {
		return "", fmt.Errorf("Can't pop value: %w", err)
	}

	return res, createZNodeWithIndex(zkKey, strconv.Itoa((clusterIndex+1)%d.nShards))
}

func (d *DQueue) findKeyForAddr(addr string) (string, error) {

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	left := 0
	right := 16000

	switch addr {
	case "158.160.9.8:6379":
		{
			left = 11
			right = 5460
			break
		}
	case "158.160.19.212:6379":
		{
			left = 10923
			right = 15911
			break
		}
	case "158.160.19.2:6379":
		{
			left = 15912
			right = 16383
			break
		}
	case "51.250.106.140:6379":
		{
			left = 5461
			right = 10922
			break
		}
	}

	counter := 0

	for {
		key := fmt.Sprintf("%s::%s::%v", prefix, d.name, counter)
		randSlot := int(client.ClusterKeySlot(client.Context(), key).Val())

		if left <= randSlot && randSlot <= right {
			return key, nil
		} else {
			counter += 1
		}
	}
}
