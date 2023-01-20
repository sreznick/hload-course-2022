package dqueue

import (
  "fmt"
  "time"
  "strings"
  "strconv"
  "encoding/json"
  "github.com/go-zookeeper/zk"
)

func getQueueNodeData(name string) []byte {
  var connection = getZKConnection()
  var path = root + "/" + name

  nodeExists, _, err := connection.Exists(path)
  if err != nil {
    panic(err)
  }

  if !nodeExists {
    return nil
  }

  result, _, err := connection.Get(path)
  if err != nil {
    panic(err)
  }

  return result
}

func CreateZKRoot() {
  zkHostsRaw := ctx.Value("zk-hosts")
  zkHosts, ok := zkHostsRaw.([]string)
  if !ok {
    panic("Invalid type")
  }

  connection, _, err := zk.Connect(zkHosts, time.Second)
  if err != nil {
    panic(err)
  }

  rootExists, _, err := connection.Exists(root)
  if err != nil {
    panic(err)
  }

  if !rootExists {
    _, err := connection.Create(root, make([]byte, 0), 0, zk.WorldACL(zk.PermAll))
    if err != nil {
      panic(err)
    }
  }
}

func getZKConnection() *zk.Conn {
  if zkConnection != nil {
    return zkConnection
  }

  zkHostsRaw := ctx.Value("zk-hosts")

  zkHosts, ok := zkHostsRaw.([]string)
  if !ok {
    panic("Invalid type")
  }

  connection, _, err := zk.Connect(zkHosts, time.Second * 5)
  if err != nil {
    panic(err)
  }

  zkConnection = connection

  return connection
}

func createDQueueNode(dqueue DQueue) {
  var connection = getZKConnection()

  serializedDqueue, err := json.Marshal(dqueue)
  if err != nil {
    panic(err)
  }

  _, err = connection.Create(root + "/" + dqueue.Name, serializedDqueue, 0, zk.WorldACL(zk.PermAll))
  if err != nil {
    panic(err)
  }
}

func updateDQueueNode(dqueue *DQueue) {
  var connection = getZKConnection()

  serializedDqueue, err := json.Marshal(dqueue)
  if err != nil {
    panic(err)
  }

  _, err = connection.Set(root + "/" + dqueue.Name, serializedDqueue, -1)
  if err != nil {
    panic(err)
  }
}

// Ackuire distributed queue lock
// See https://zookeeper.apache.org/doc/r3.1.2/recipes.html ("Locks" section)
func lock(dqueue *DQueue) {
  var connection = getZKConnection()
  var path = fmt.Sprintf("%s/%s/__locknode-", root, dqueue.Name)

  val, err := connection.Create(path, make([]byte, 0), zk.FlagEphemeral | zk.FlagSequence, zk.WorldACL(zk.PermAll))
  if err != nil {
    panic(err)
  }

  for {
    childs, _, err := connection.Children(fmt.Sprintf("%s/%s", root, dqueue.Name))
    if err != nil {
      panic(err)
    }

    var nodeNumber = extractNumberFromLocknode(val)
    var canGetLock = true
    var nextMinNumber = -1

    for _, child := range(childs) {
      var childNumber = extractNumberFromLocknode(child)
      if childNumber < nodeNumber {
        if childNumber > nextMinNumber {
          nextMinNumber = childNumber
        }
        canGetLock = false
      }
    }

    if canGetLock {
      dqueue.lockSequenceNumber = nodeNumber
      return
    }

    var data, _, ch, errExists = connection.ExistsW(fmt.Sprintf("%s/%s/__locknode-%010d", root, dqueue.Name, nextMinNumber), )
    if errExists != nil {
      panic(errExists)
    }
    if !data {
      continue
    }

    <-ch
    continue
  }
}

// Release distributed queue lock
func unlock(dqueue *DQueue) {
  if dqueue.lockSequenceNumber == -1 {
    panic("assert failed")
  }

  var connection = getZKConnection()
  var path = fmt.Sprintf("%s/%s/__locknode-%010d", root, dqueue.Name, dqueue.lockSequenceNumber)

  connection.Delete(path, -1)
  dqueue.lockSequenceNumber = -1
}

// extract node sequence number from ZK path
func extractNumberFromLocknode(path string) int {
  var s = strings.Split(path, "/")
  var result, err = strconv.Atoi(s[len(s) - 1][11:])
  if err != nil {
    panic(nil)
  }

  return result
}
