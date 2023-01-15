package producer

import (
	"context"
	"database/sql"
	"fmt"
	"main/common"
	"strconv"
)

func getUpdateQuery() string {
	return "update urls set clicks = clicks + " + strconv.Itoa(common.ClicksThrsh) + " WHERE ID = $1;"
}

// Read kafka url topic
func GetClicks(db *sql.DB) {
	ctx := context.Background()
	for {
		v, err := clicksReader.FetchMessage(ctx)

		if err != nil {
			fmt.Printf("Kafka error: %s\n", err)
			continue
		}

		t, err := UrlToId(string(v.Key))
		if err != nil {
			fmt.Printf("Something went wrong with Clicks parsing: %s\n", err)
			continue
		}

		_, err = db.Exec(getUpdateQuery(), t)
		if err != nil {
			fmt.Printf("Something went wrong with Clicks in DB: %s\n", err)
			continue
		}

		err = clicksReader.CommitMessages(ctx, v)
		if err != nil {
			fmt.Printf("Kafka error: %s\n", err)
		}

		fmt.Printf("New clicks recieved -- %s, %d\n", v.Key, common.ClicksThrsh)
	}
}
