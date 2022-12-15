package dqueue

import (
	"errors"
	"fmt"
	"github.com/go-zookeeper/zk"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (dq *DQueue) SaveToZk(redisAddr string) error {
	zkAddr := ctx.Value("zkCluster")
	switch v := zkAddr.(type) {
	case []string:
		c, _, err := zk.Connect(v, time.Second) //*10)
		if err != nil {
			panic(err)
		}

		err = lockZk(c, dq.infoPrefix, dq.name)
		for {
			if err == nil {
				break
			}
			time.Sleep(100)
			err = lockZk(c, dq.infoPrefix, dq.name)
		}
		pathToRedis := fmt.Sprintf("/%s/%s", dq.queuePrefix, dq.name)

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

		path := fmt.Sprintf("/%s/%s/%d", dq.queuePrefix, dq.name, index)
		_, errCreate := c.Create(path, []byte(redisAddr), 0, zk.WorldACL(zk.PermAll))
		if errCreate != nil {
			panic(errCreate)
		}

		err = unlockZk(c, dq.infoPrefix, dq.name)
		if err != nil {
			return err
		}
	}
	return nil
}

func getOrCreateZkPath(c *zk.Conn, path []string) {
	var currentPath = ""
	for _, subPath := range path {
		currentPath += "/" + subPath
		exists, _, err := c.Exists(currentPath)
		if err != nil {
			panic(err)
		}
		if !exists {
			_, errCreate := c.Create(currentPath, []byte{}, 0, zk.WorldACL(zk.PermAll))
			if errCreate != nil {
				panic(errCreate)
			}
		}
	}
}

func (dq *DQueue) getNextAddress() (int, error) {
	zkAddr := ctx.Value("zkCluster")
	switch v := zkAddr.(type) {
	case []string:
		c, _, err := zk.Connect(v, time.Second) //*10)
		if err != nil {
			panic(err)
		}
		err = lockZk(c, dq.infoPrefix, dq.name)
		for {
			if err == nil {
				break
			}
			time.Sleep(100)
			err = lockZk(c, dq.infoPrefix, dq.name)
		}
		getOrCreateZkPath(c, []string{dq.queuePrefix, dq.name})
		pathToRedis := fmt.Sprintf("/%s/%s", dq.queuePrefix, dq.name)

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
		res := 0
		err = nil
		for _, child := range childrenIds {
			path := fmt.Sprintf("%s/%d", pathToRedis, child)
			redisAddrBytes, _, err := c.Get(path)
			if err != nil {
				break
			}
			err = c.Delete(path, -1)
			if err != nil {
				break
			}
			redisAddr, _ := strconv.Atoi(string(redisAddrBytes[:]))
			if checkRedisAddress(redisAddr) {
				res = redisAddr
				break
			}
		}
		err = unlockZk(c, dq.infoPrefix, dq.name)
		if err != nil {
			return 0, err
		}
		return res, err
	}
	return 0, nil
}

func lockZk(c *zk.Conn, prefix string, name string) error {
	path := []string{prefix, name}
	strPath := "/" + strings.Join(path, "/")
	getOrCreateZkPath(c, path)
	strLock := strPath + "/lock"
	exists, _, err := c.Exists(strLock)
	if err != nil {
		return err
	}
	if !exists {
		_, _ = c.Create(strPath+"/lock", []byte{}, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
		return nil
	}

	return errors.New("queue was locked")
}

func unlockZk(c *zk.Conn, prefix string, name string) error {
	path := []string{prefix, name}
	strPath := "/" + strings.Join(path, "/") + "/lock"
	err := c.Delete(strPath, -1)
	return err
}
