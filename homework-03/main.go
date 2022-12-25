package main

import (
  "fmt"
  "os"
  "io/ioutil"

  "gopkg.in/yaml.v3"

  master "hw3/master"
  worker "hw3/worker"
  shared "hw3/shared"
)

func loadConfig(filename string) (*shared.Config, error) {
    buf, err := ioutil.ReadFile(filename)

    if err != nil {
      return nil, err
    }

    c := &shared.Config{}
    err = yaml.Unmarshal(buf, c)
    if err != nil {
        return nil, fmt.Errorf("in file %q: %w", filename, err)
    }

    return c, err
}

func print_help() {
  var help = `
    Usage:
      ROLE=<role> go run main.go
      ROLE := master | worker
    `

    fmt.Println(help)
}

func dispatch(role string, c *shared.Config) {
  if (role == "master") {
    master.Serve(c.Master)
  } else if (role == "worker") {
    worker.Serve(c.Worker)
  } else {
    print_help();
    os.Exit(1)
  }
}

func main() {
  var role = os.Getenv("ROLE")

  if (len(role) == 0) {
    print_help();
    os.Exit(1)
  }

  var config, err = loadConfig("./config.yml")
  if err != nil {
    panic(err)
  }

  dispatch(role, config)
}
