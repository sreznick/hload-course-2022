package main

import (
    "fmt"
    "dqueue"
    "io/ioutil"
    "gopkg.in/yaml.v2"
)

type Config struct {
  RedisHosts []string `yaml:"redis"`
  ZookeeperHosts []string `yaml:"zookeeper"`
}

func main() {
  var config Config

  configRaw, err := ioutil.ReadFile("./config/config.yml")
  if err != nil {
      panic(err)
  }

  err = yaml.Unmarshal([]byte(configRaw), &config)
  if err != nil {
    panic(err)
  }

  dqueue.Config(config.RedisHosts, config.ZookeeperHosts)

  q1, err := dqueue.Open("test-10", 4)
  if err != nil  {
    panic(err)
  }

  q2, err := dqueue.Open("test-11", 3)
  if err != nil  {
    panic(err)
  }

  q1.Push("value-0")
  q2.Push("value-0")
  q1.Push("value-1")
  q2.Push("value-1")
  q1.Push("value-2")
  q2.Push("value-2")
  q1.Push("value-3")
  q2.Push("value-3")
  q1.Push("foobar")
  q2.Push("foobar")
  q1.Push("asdasdasd")
  q2.Push("asdasdasd")

  for i := 0; i < 10; i++ {
    pullQueue(q1)
    pullQueue(q2)
  }
}

func pullQueue(q dqueue.DQueue) {
  val, err := q.Pull()

  if err != nil && err.Error() == "No value to pop" {
    fmt.Printf("(nil)\n")
    return
  }

  if err != nil {
    panic(err)
  } else {
    fmt.Printf("%s\n", val)
  }
}
