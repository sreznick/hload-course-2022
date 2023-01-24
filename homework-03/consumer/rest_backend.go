package consumer

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	clicksThrsh = 10
)

func getUrl(c *gin.Context, urlVarName string) {
	tinyUrl := c.Params.ByName(urlVarName)
	longUrl, err := GetLongUrl(tinyUrl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"response": "Something went wrong with database: " + err.Error()})
		return
	}

	err = IncrementClick(tinyUrl)
	if err != nil {
		fmt.Printf("Something went wrong with Redis: %s\n", err)
	}

	cnt, err := GetClicks(tinyUrl)
	if err != nil {
		fmt.Printf("Something went wrong with Redis: %s\n", err)
	} else {
		if cnt%clicksThrsh == 0 {
			PushClicks(tinyUrl)
		}
	}

	c.Redirect(302, longUrl)
}

func SetupWorker() *gin.Engine {
	r := gin.Default()

	urlVarName := "url"
	r.GET(fmt.Sprintf("/:%s", urlVarName), func(c *gin.Context) {
		getUrl(c, urlVarName)
	})

	return r
}
