module main

go 1.17

replace deque => ../greetings

require (
	dqueue v0.0.0-00010101000000-000000000000
	github.com/go-redis/redis/v8 v8.11.5
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-zookeeper/zk v1.0.3 // indirect
	github.com/pkg/errors v0.9.1 // indirect
)

replace dqueue => ../dqueue
