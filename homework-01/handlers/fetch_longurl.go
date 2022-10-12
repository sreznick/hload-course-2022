package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"

    services "main/services"
)

func FetchLongURL(c *gin.Context) {
  tinyurl := c.Params.ByName("tinyurl")

  if result, err := services.FetchLongURL(tinyurl); err != nil {
    if err == services.ErrNotFound {
      c.Writer.WriteHeader(http.StatusNotFound)
    } else {
      c.Writer.WriteHeader(http.StatusInternalServerError)
    }
  } else {
    c.Redirect(http.StatusFound, result)
  }
}
