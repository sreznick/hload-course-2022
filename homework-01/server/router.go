package server

import (
	"database/sql"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

var (
	createRequestsNumber = promauto.NewCounter(prometheus.CounterOpts{
		Name: "total_create_requests",
		Help: "The total number of processed `create` requests",
	})
	getRequestsNumber = promauto.NewCounter(prometheus.CounterOpts{
		Name: "total_get_requests",
		Help: "The total number of processed `get` requests",
	})
	createRequestTime = promauto.NewSummary(prometheus.SummaryOpts{
		Name: "put_request_time",
		Help: "Average time of `create` request processing",
	})

	getRequestTime = promauto.NewSummary(prometheus.SummaryOpts{
		Name: "get_request_time",
		Help: "Average time of `get` request processing",
	})
)

const (
	selectByUrl = "SELECT id FROM urlsStorage WHERE long_url = $1"
	selectById  = "SELECT long_url FROM urlsStorage WHERE id = $1"
	insertUrl   = "INSERT INTO urlsStorage(long_url) VALUES ($1) ON CONFLICT DO NOTHING"

	createUrl = "/create"
	getUrl    = "/:tinyurl"
)

type CreateRequest struct {
	Longurl string `json:"longurl"`
}

func Setup(db *sql.DB) *gin.Engine {
	request := gin.Default()

	request.PUT(createUrl, func(c *gin.Context) {
		createRequestsNumber.Inc()
		startHandlingTime := time.Now()

		var request CreateRequest
		err := c.BindJSON(&request)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"response": err})
		} else {
			longUrl := request.Longurl
			var id int
			selectErr := db.QueryRow(selectByUrl, longUrl).Scan(&id)
			if selectErr != nil {
				_, insertErr := db.Exec(insertUrl, longUrl)
				if insertErr != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"response": "insertion url error: " + err.Error()})
					return
				}
				_ = db.QueryRow(selectByUrl, longUrl).Scan(&id)

			}
			resultTinyUrl, _ := GenerateTinyUrl(id)
			c.JSON(http.StatusOK, gin.H{"longurl": longUrl, "tinyurl": resultTinyUrl})

		}
		elapsed := float64(time.Since(startHandlingTime).Milliseconds())
		createRequestTime.Observe(elapsed)
	})

	request.GET(getUrl, func(c *gin.Context) {
		getRequestsNumber.Inc()
		startHandlingTime := time.Now()

		tinyUrl := c.Params.ByName("tinyurl")
		id, err := GetIdByTinyUrl(tinyUrl)
		if err != nil {
			c.Writer.WriteHeader(http.StatusNotFound)
		} else {
			var longUrl string
			selectErr := db.QueryRow(selectById, id).Scan(&longUrl)
			if selectErr != nil {
				c.Writer.WriteHeader(http.StatusNotFound)
			} else {
				c.Redirect(http.StatusFound, longUrl)
			}

		}

		elapsed := float64(time.Since(startHandlingTime).Milliseconds())
		getRequestTime.Observe(elapsed)
	})

	return request
}
