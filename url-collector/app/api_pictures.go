package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

const (
	PLANETARY_APOD_URL = "https://api.nasa.gov/planetary/apod"
)

var client = resty.New()

func PicturesGet(c *gin.Context) {
	queryParams, err := validatePicturesGet(c)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	responseBody, err := processPicturesGet(queryParams)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("request couldn't be processed due to error: %v", err)})
		return
	}
	log.Println(responseBody.Urls)
	c.JSON(http.StatusOK, responseBody)
}
