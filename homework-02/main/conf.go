package main

import (
	"dqueue"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"os"
)

type RedisYaml struct {
	Options struct {
		Addr     string `json:"addr"`
		Password string `json:"password"`
		DB       int    `json:"db"`
	} `json:"options"`

	LowerKey int64 `json:"lower_key"`
	UpperKey int64 `json:"upper_key"`
}

type RedisClusterYaml struct {
	Cluster []RedisYaml `json:"cluster"`
}

type ZKHostsYaml struct {
	Zks []string `json:"zks"`
}

type Config struct {
	Zk    ZKHostsYaml      `json:"zk"`
	Redis RedisClusterYaml `json:"redis"`
}

func ReadConfig() (dqueue.RedisCluster, []string, error) {
	file, err := os.ReadFile("config/conf.json")
	if err != nil {
		return dqueue.RedisCluster{}, []string{}, err
	}

	var c Config
	err = json.Unmarshal(file, &c)
	if err != nil {
		return dqueue.RedisCluster{}, []string{}, err
	}

	return translateRedis(c.Redis), translateZK(c.Zk), nil
}

func translateRedis(c RedisClusterYaml) dqueue.RedisCluster {
	var rs dqueue.RedisCluster
	for _, r := range c.Cluster {
		dqr := dqueue.RedisHost{
			Conf:     redis.Options{Addr: r.Options.Addr, Password: r.Options.Password, DB: r.Options.DB},
			LowerKey: r.LowerKey,
			UpperKey: r.UpperKey,
		}

		rs.RedisHosts = append(rs.RedisHosts, dqr)
	}

	return rs
}

func translateZK(z ZKHostsYaml) []string {
	var zks []string
	for _, h := range z.Zks {
		zks = append(zks, h)
	}

	return zks
}
