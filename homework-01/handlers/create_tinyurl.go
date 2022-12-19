package handlers

import (
      "net/http"
      "github.com/gin-gonic/gin"

      services "main/services"
)

type CreateTinyURLRequestModel struct {
  Longurl string `json:"longurl" binding:"required"`
}


func CreateTinyURL(c *gin.Context) {
  var requestModel CreateTinyURLRequestModel

  if err := c.BindJSON(&requestModel); err != nil {
    c.JSON(http.StatusUnprocessableEntity, gin.H{})
    return
  }

  result, err := services.CreateTinyURL(requestModel.Longurl)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{})
    return
  }

  c.JSON(http.StatusOK, gin.H{"longurl": requestModel.Longurl, "tinyurl": result})
}
