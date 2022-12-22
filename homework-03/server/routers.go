package server

import (
	"container/list"
	"localKafka"
	"localRedis"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
)

type CreateRequest struct {
	Longurl string `json:"longurl"`
}

func SetupRouter(cluster *localRedis.RedisCluster, urlWriter *kafka.Writer, urlReaders *list.List) *gin.Engine {
	r := gin.Default()

	r.PUT("/create", func(c *gin.Context) {
		var request CreateRequest
		err := c.BindJSON(&request)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"response": err})
		} else {
			longUrl := request.Longurl
			tinyUrl, isNew := localRedis.GetTinyUrl(cluster, longUrl)
			if isNew {
				go localKafka.UrlProduce(urlWriter, (*cluster).Ctx, longUrl, tinyUrl)
				id := 0
				for e := urlReaders.Front(); e != nil; e = e.Next() {
					reader, ok := e.Value.(*kafka.Reader)
					if ok {
						go localKafka.UrlConsume(reader, (*cluster).Ctx, cluster, id)
					}
					id++
				}
				time.Sleep(time.Second * 10)
			}
			c.JSON(http.StatusOK, gin.H{"longurl": longUrl, "tinyurl": tinyUrl})

		}

	})

	r.GET("/:tinyurl", func(c *gin.Context) {

		tinyUrl := c.Params.ByName("tinyurl")
		longUrl, err := localRedis.CheckTinyUrl(cluster, tinyUrl)
		if err != nil {
			c.Writer.WriteHeader(http.StatusNotFound)
		} else {
			c.Redirect(http.StatusFound, longUrl)
		}

	})

	return r

}
