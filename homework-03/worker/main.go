package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"log"
	worker "main/worker/src"
)

func main() {
	var (
		r   = worker.SetupRouter()
		ctx = context.Background()
	)
	go worker.ReplicaReadNewDataFromMaster(ctx, "mdiagilev-test-master")

	r.GET("/:tiny", func(ctx *gin.Context) {
		if err := worker.Get(ctx); err != nil {
			log.Fatalf("worker.Get: %v", err)
		}
	})

	if err := r.Run("0.0.0.0:8081"); err != nil {
		log.Fatalf("r.Runn: %v", err)
	}
}
