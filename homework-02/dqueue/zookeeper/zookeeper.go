package zookeeper

import (
	"dqueue/models"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/pkg/errors"
)

type zookeeper struct {
	zkHosts []string
	conn    *zk.Conn
}

func NewZookeeper() *zookeeper {
	return &zookeeper{}
}

func (z *zookeeper) Config(zkHosts []string) {
	z.zkHosts = zkHosts
}

func (z *zookeeper) Connect() (*zk.Conn, error) {
	c, _, err := zk.Connect(z.zkHosts, time.Second)
	if err != nil {
		return nil, errors.Wrap(err, "Problems with ZooKeeper")
	}

	return c, nil
}

func (z *zookeeper) Lock(c *zk.Conn, path, queue string) (string, error) {
	name, err := c.Create(path+queue+"-lock-", nil, zk.FlagSequence|zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	if err != nil {
		return "", errors.Wrap(err, "Problems with node creation")
	}

	for {
		children, _, err := c.Children(path[:len(path)-1])
		if err != nil {
			return "", errors.Wrap(err, "Internal zookeeper error, get children")
		}

		withLowest := name
		for _, child := range children {
			if strings.HasPrefix(child, path+queue+"-lock-") {
				lowest, err := strconv.Atoi(strings.TrimPrefix(child, path+queue+"-lock-"))
				if err != nil {
					return "", errors.Wrap(err, "Internal error")
				}
				suffix, err := strconv.Atoi(strings.TrimPrefix(child, path+queue+"-lock-"))
				if err != nil {
					return "", errors.Wrap(err, "Internal error")
				}

				if suffix < lowest {
					withLowest = child
					break
				}
			}
		}
		if withLowest == name {
			return name, nil
		}

		exists, _, events, err := c.ExistsW(withLowest)
		if err != nil {
			return "", errors.Wrap(err, "Internal zookeeper error, exists + watch")
		}

		if exists {
			_ = <-events
		}
	}
}

func (z *zookeeper) UnLock(c *zk.Conn, lock string) error {
	_, stat, err := c.Get(lock)
	if err != nil {
		return errors.Wrap(err, "Internal zookeeper error, get lock node stat")
	}

	err = c.Delete(lock, stat.Version)
	if err != nil {
		return errors.Wrap(err, "Internal zookeeper error, delete lock")
	}

	return nil
}

func (z *zookeeper) Get(c *zk.Conn, node string) (*models.QueueInfo, error) {
	data, _, err := c.Get(node)
	if err != nil {
		return nil, errors.Wrap(err, "Internal zookeeper error, get node stat to get")
	}

	var result models.QueueInfo
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, errors.Wrap(err, "Internal error")
	}

	return &result, nil
}

func (z *zookeeper) Create(c *zk.Conn, node string, q models.QueueInfo) error {
	data, err := json.Marshal(q)
	if err != nil {
		return errors.Wrap(err, "Internal error")
	}

	_, err = c.Create(node, data, 0, zk.WorldACL(zk.PermAll))
	if err != nil {
		return errors.Wrap(err, "Problems with node creation")
	}

	return nil
}

func (z *zookeeper) Update(c *zk.Conn, node string, q models.QueueInfo) error {
	_, stat, err := c.Get(node)
	if err != nil {
		return errors.Wrap(err, "Internal zookeeper error, get node info to update")
	}

	data, err := json.Marshal(q)
	if err != nil {
		return errors.Wrap(err, "Internal error")
	}

	_, err = c.Set(node, data, stat.Version)
	if err != nil {
		return errors.Wrap(err, "Internal zookeeper error, set data")
	}

	return nil
}
