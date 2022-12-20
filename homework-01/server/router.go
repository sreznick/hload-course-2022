package server

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

const (
	selectByUrl = "select id from urlsStorage where long_url = $1"
	selectById  = "select long_url from urlsStorage where id = $1"
	insertUrl   = "insert into urlsStorage(long_url) values ($1) on conflict do nothing"
)

type CreateRequest struct {
	Longurl string `json:"longurl"`
}

func SetupRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()

	r.PUT("/create", func(c *gin.Context) {
		putRequestsNumber.Inc()
		start := time.Now()

		var request CreateRequest
		err := c.BindJSON(&request)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"response": err})
		} else {
			longUrl := request.Longurl
			var id int

			select_err := db.QueryRow(selectByUrl, longUrl).Scan(&id)
			if select_err != nil {
				_, insert_err := db.Exec(insertUrl, longUrl)
				if insert_err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"response": "insertion url error: " + err.Error()})
					return
				}
				_ = db.QueryRow(selectByUrl, longUrl).Scan(&id)

			}
			tinyUrl := GenerateTinyUrl(id)
			c.JSON(http.StatusOK, gin.H{"longurl": longUrl, "tinyurl": tinyUrl})

		}

		elapsed := float64(time.Since(start).Milliseconds())
		putRequestTime.Observe(elapsed)

	})

	r.GET("/:tinyurl", func(c *gin.Context) {
		getRequestsNumber.Inc()
		start := time.Now()

		tinyUrl := c.Params.ByName("tinyurl")
		id, err := GetIdByTinyUrl(tinyUrl)
		if err != nil {
			c.Writer.WriteHeader(http.StatusNotFound)
		} else {
			var longUrl string
			select_err := db.QueryRow(selectById, id).Scan(&longUrl)
			if select_err != nil {
				c.Writer.WriteHeader(http.StatusNotFound)
			} else {
				c.Redirect(http.StatusFound, longUrl)
			}

		}

		elapsed := float64(time.Since(start).Milliseconds())
		getRequestTime.Observe(elapsed)
	})

	return r
}
