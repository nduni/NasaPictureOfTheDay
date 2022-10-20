package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/nduni/S3J5c3RpYW4gR29nb0FwcHMgTkFTQQ/url-collector/config"
	collectorModels "github.com/nduni/S3J5c3RpYW4gR29nb0FwcHMgTkFTQQ/url-collector/models/collector"
	nasaModels "github.com/nduni/S3J5c3RpYW4gR29nb0FwcHMgTkFTQQ/url-collector/models/nasa"
	"go.uber.org/multierr"
)

const (
	PLANETARY_APOD_URL = "https://api.nasa.gov/planetary/apod"
)

var client = resty.New()

func PicturesGet(c *gin.Context) {
	queryParams, err := validatePicturesGet(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	responseBody, err := processPicturesGet(queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("request couldn't be processed due to error: %v", err)})
		return
	}
	log.Println(responseBody.Urls)
	c.JSON(http.StatusOK, responseBody)
}

type NasaResponse struct {
	Resp *resty.Response
	Err  error
}

func processPicturesGet(queryParams collectorModels.PicturesQueryParams) (collectorModels.Pictures, error) {
	done := make(chan struct{})
	defer close(done)

	nasaParam := sendParams(done, queryParams)

	resps := make(chan NasaResponse)
	var wg sync.WaitGroup
	wgCounter := setWgCounter(queryParams)
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

func sendParams(done <-chan struct{}, queryParams collectorModels.PicturesQueryParams) <-chan nasaModels.NasaParams {
	log.Println("processing query params")
	nasaParams := make(chan nasaModels.NasaParams)
	go func() {
		for date := queryParams.From; date.After(queryParams.To) == false; date = date.AddDate(0, 0, 1) {
			param := nasaModels.NasaParams{
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

func setWgCounter(queryParams collectorModels.PicturesQueryParams) int {
	wgCounter := config.Config.ConcurrentRequests
	if diff := queryParams.To.Sub(queryParams.From); int(diff.Hours()/24)+1 < config.Config.ConcurrentRequests {
		wgCounter = int(diff.Hours() / 24)
	}
	return wgCounter
}

func requestNasaApi(done <-chan struct{}, params <-chan nasaModels.NasaParams, r chan<- NasaResponse) {
	log.Println("sending requests")
	queryParams := map[string]string{
		"api_key": config.Config.ApiKey,
	}

	for param := range params {
		var nasaResp NasaResponse
		queryParams["date"] = param.Date.Format("2006-01-02")
		resp, err := client.R().
			SetQueryParams(queryParams).
			Get(PLANETARY_APOD_URL)
		url := resp.Request.URL
		var errLog string
		if err != nil {
			errLog = fmt.Sprintf("request to '%v' failed. error during request: %v", url, err.Error())
			log.Println(errLog)
		} else if resp.StatusCode() != http.StatusOK {
			errLog = fmt.Sprintf("request to '%v' failed. response status code doesn't equal 200: %v", url, resp.StatusCode())
			log.Println(errLog)

		} else {
			log.Printf("request to '%v' was succesful", url)
		}

		if errLog != "" {
			err = errors.New(errLog)
		}
		nasaResp = NasaResponse{resp, err}

		select {
		case r <- nasaResp:
		case <-done:
			return
		}
	}
}

func createResponseBody(resps chan NasaResponse) (collectorModels.Pictures, error) {
	pictures := collectorModels.Pictures{}
	urls := []string{}
	var errs []error
	for resp := range resps {
		if resp.Err != nil {
			errs = append(errs, resp.Err)
			continue
		}

		url, err := retrieveUrlFromResp(resp.Resp.Body())
		if err != nil {
			errs = append(errs, err)
			continue
		}
		log.Println("append url ", url)
		urls = append(urls, url)
	}

	pictures.Urls = urls
	return pictures, multierr.Combine(errs...)
}

func retrieveUrlFromResp(body []byte) (string, error) {
	var apod nasaModels.Apod
	err := json.Unmarshal(body, &apod)
	if err != nil {
		return "", err
	}
	return apod.URL, nil
}
