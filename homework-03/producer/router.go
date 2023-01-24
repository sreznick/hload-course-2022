package producer

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"math"
	"math/rand"
	"net/http"
	"time"
)

const (
	insertNewUrl = "insert into urls(id, url) values ($1, $2)"
	selectByUrl  = "select id from urls where url = $1"

	uniqueViolationErrorCode = "23505"
)

type CreateRequestJsonBody struct {
	Longurl string `json:"longurl"`
}

func getLongUrlFromJson(c *gin.Context) (string, error) {
	body := CreateRequestJsonBody{}
	err := c.BindJSON(&body)
	if err != nil {
		errorMessage := "wrong json format"
		return "", fmt.Errorf(errorMessage)
	}

	return body.Longurl, nil
}

func createNewId(db *sql.DB, longUrl string) (*int64, error) {
	var tinyUrlId int64
	for {
		s1 := rand.NewSource(time.Now().UnixNano())
		r1 := rand.New(s1)
		tinyUrlId = r1.Int63n(int64(math.Pow(62, 7)))
		err := db.QueryRow(insertNewUrl, tinyUrlId, longUrl).Err() //TODO is queryrow?

		if err == nil {
			return &tinyUrlId, nil
		}

		pqerr, ok := err.(*pq.Error)
		if !ok {
			return nil, fmt.Errorf("Internal error: `QueryRow` returned not *pq.Error")
		}

		if pqerr.Code == uniqueViolationErrorCode {
			continue
		}

		return nil, err
	}
}

func hasUrl(db *sql.DB, longUrl string) (*int64, bool, error) {
	var tinyUrlId int64
	err := db.QueryRow(selectByUrl, longUrl).Scan(&tinyUrlId)
	if err == sql.ErrNoRows {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	} else {
		return &tinyUrlId, true, nil
	}
}

func create(c *gin.Context, db *sql.DB) {
	longUrl, err := getLongUrlFromJson(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"response": err})
		return
	}

	tinyUrlId, has, err := hasUrl(db, longUrl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"response": "Something went wrong with database (router:80): " + err.Error()})
		return
	}

	if !has {
		tinyUrlId, err = createNewId(db, longUrl)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"response": "Something went wrong with database (router:87): " + err.Error()})
			return
		}
	}

	tinyUrl, err := IdToUrl(*tinyUrlId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"response": "Impossible to return tiny url: no more space for unique urls (router:94)"})
		return
	}

	if !has {
		PushCreatedUrl(tinyUrl, longUrl)
	}

	c.JSON(http.StatusOK, gin.H{"longurl": longUrl, "tinyurl": tinyUrl})
}

func getUrl(c *gin.Context, db *sql.DB, urlVarName string) {
	shortUrl := c.Params.ByName(urlVarName)
	shortUrlId, err := UrlToId(shortUrl)
	if err != nil {
		c.Writer.WriteHeader(404)
		return
	}
	var longUrl string

	err = db.QueryRow("select url from urls where id = $1", shortUrlId).Scan(&longUrl)
	if err != nil {
		c.Writer.WriteHeader(404)
		return
	}

	c.Redirect(302, longUrl)
}

func SetupRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()

	r.PUT("/create", func(c *gin.Context) {
		create(c, db)
	})

	urlVarName := "url"
	r.GET(fmt.Sprintf("/:%s", urlVarName), func(c *gin.Context) {
		getUrl(c, db, urlVarName)
	})

	return r
}
