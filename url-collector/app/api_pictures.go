package app

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	collectorModels "github.com/nduni/S3J5c3RpYW4gR29nb0FwcHMgTkFTQQ/url-collector/models/collector"
	nasaModels "github.com/nduni/S3J5c3RpYW4gR29nb0FwcHMgTkFTQQ/url-collector/models/nasa"
)

const (
	CONCURRENT_REQUESTS = 5
	PLANETARY_APOD_URL  = "https://api.nasa.gov/planetary/apod"
	API_KEY             = "DEMO_KEY"
)

var client = resty.New()

func PicturesGet(c *gin.Context) {
	queryParams, err := validatePicturesGet(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	responseBody, err := processPicturesGet(queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "request couldn't be processed due to processing error"})
	}
	log.Println(responseBody.Urls)
	c.JSON(http.StatusOK, responseBody)
}

type NasaParams struct {
	Date      time.Time
	Copyright *string
}

func processPicturesGet(queryParams collectorModels.PicturesQueryParams) (collectorModels.Pictures, error) {
	done := make(chan struct{})
	defer close(done)

	nasaParam := sendParams(done, queryParams)

	resps := make(chan *resty.Response)
	var wg sync.WaitGroup
	wgCounter := CONCURRENT_REQUESTS
	if diff := queryParams.To.Sub(queryParams.From); int(diff.Hours()/24)+1 < CONCURRENT_REQUESTS {
		wgCounter = int(diff.Hours() / 24)
	}
	wg.Add(wgCounter)
	for i := 0; i < wgCounter; i++ {
		go func() {
			requestNasaApi(done, nasaParam, resps)
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(resps)
	}()

	return createResponseBody(resps)
}

func createResponseBody(resps chan *resty.Response) (collectorModels.Pictures, error) {
	pictures := collectorModels.Pictures{}
	urls := []string{}
	for resp := range resps {
		url, err := retrieveUrlFromResp(resp.Body())
		if err != nil {
			return pictures, err
		}
		log.Println("url ", url)
		urls = append(urls, url)
	}
	pictures.Urls = urls
	return pictures, nil
}

func retrieveUrlFromResp(body []byte) (string, error) {
	var apod nasaModels.Apod
	err := json.Unmarshal(body, &apod)
	if err != nil {
		return "", err
	}
	return apod.URL, nil
}

func sendParams(done <-chan struct{}, queryParams collectorModels.PicturesQueryParams) <-chan NasaParams {
	log.Println("processing query params")
	nasaParams := make(chan NasaParams)
	go func() {
		for date := queryParams.From; date.After(queryParams.To) == false; date = date.AddDate(0, 0, 1) {
			param := NasaParams{
				Date:      date,
				Copyright: queryParams.Copyright,
			}
			log.Printf("params: %v, %v", param.Date.String(), param.Copyright)
			select {
			case nasaParams <- param:
			case <-done:
				return
			}
		}
		close(nasaParams)
	}()
	return nasaParams
}

func requestNasaApi(done <-chan struct{}, params <-chan NasaParams, r chan<- *resty.Response) {
	log.Println("sending requests")
	queryParams := map[string]string{
		"api_key": API_KEY,
	}

	for param := range params {
		queryParams["date"] = param.Date.Format("2006-01-02")
		resp, err := client.R().
			SetQueryParams(queryParams).
			Get(PLANETARY_APOD_URL)
		log.Println(resp.Request.URL)
		if err != nil {
			log.Println("error during request: ", err.Error())
			continue
		}
		if resp.StatusCode() != http.StatusOK {
			log.Println("response status code doesn't equal 200: ", resp.StatusCode())
			continue
		}
		log.Printf("request to '%v' was succesful", PLANETARY_APOD_URL)
		select {
		case r <- resp:
		case <-done:
			return
		}
	}
}
