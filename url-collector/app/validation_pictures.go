package app

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	collectorModels "github.com/nduni/S3J5c3RpYW4gR29nb0FwcHMgTkFTQQ/url-collector/models/collector"
)

func validatePicturesGet(c *gin.Context) (collectorModels.PicturesQueryParams, error) {
	var queryParams collectorModels.PicturesQueryParams
	if err := c.BindQuery(&queryParams); err != nil {
		return queryParams, err
	}
	if c.Request.Body != http.NoBody {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			return queryParams, fmt.Errorf("request body not empty: %s", err)
		}
		return queryParams, fmt.Errorf("request body not empty: %s", bodyBytes)
	}
	return queryParams, nil
}
