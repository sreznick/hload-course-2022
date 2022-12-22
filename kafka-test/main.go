package main

import (
	"context"
	"time"

	"localKafka"

	"github.com/segmentio/kafka-go"
)

func main() {
	ctx := context.Background()

	urlWriter := localKafka.CreateUrlWriter()
	urlReaders := localKafka.CreateUrlReaders(2)
	go localKafka.UrlProduce(urlWriter, ctx, "alala", "000000A")
	for e := urlReaders.Front(); e != nil; e = e.Next() {
		reader, ok := e.Value.(*kafka.Reader)
		if ok {
			go localKafka.UrlConsume(reader, ctx)

		}

	}

	//readMessages := make(map[string]string)
	//go produce(ctx)
	//time.Sleep(time.Second)

	//go consume(ctx, &readMessages)
	//go consume2(ctx)

	time.Sleep(time.Second * 10)
	/*for k := range readMessages { // Loop
		fmt.Println(k)
	}*/
	/*for i := 0; i < 2; i++ {
		fmt.Println(i)
	}*/
}
