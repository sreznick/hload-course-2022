package dqueue

import (
	"dqueue/common"
	"github.com/go-zookeeper/zk"
	"strconv"
	"time"
)

const (
	zkSeqNodeNum = 10
)

type ZKHandler struct {
	prefix    string
	queuePath string
	queueName string

	hosts []string
}

func (zkh ZKHandler) buildQueuePath() string {
	return zkh.prefix + "/" + zkh.queuePath + "/" + zkh.queueName
}

func Lock(c *zk.Conn, path string) error {
	// Post lock children
	locknodePath := path + "/_locknode"
	path, err := c.Create(locknodePath+"/lock", []byte{}, zk.FlagEphemeral|zk.FlagSequence, zk.WorldACL(zk.PermAll))
	if err != nil {
		return err
	}

	// Get seq num
	seqNum, err := getNumFromSeqNodePath(path)
	if err != nil {
		return err
	}

	// Wait until obtain
	for {
		children, _, err := c.Children(locknodePath)
		if err != nil {
			return err
		}
		minNode, minId := common.Min(common.Fmap(children, getNumFromSeqNodePath))

		// If not minimum node -- wait with watch
		// In other way lock is obtained
		if minNode < seqNum {
			e, _, ch, err := c.ExistsW(locknodePath + "/" + children[minId])
			if err != nil {
				return err
			}

			// If not exists -- nothing to wait
			if !e {
				continue
			}

			// To prevent deadlock
			select {
			case <-ch:
				continue
			case <-time.After(5 * time.Second):
				continue
			}
		} else {
			return nil
		}
	}
}

func getNumFromSeqNodePath(path string) (int, error) {
	i, err := strconv.Atoi(path[len(path)-zkSeqNodeNum:])
	if err != nil {
		return -1, err
	}

	return i, nil
}
