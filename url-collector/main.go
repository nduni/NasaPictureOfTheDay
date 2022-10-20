package main

import (
	"github.com/nduni/S3J5c3RpYW4gR29nb0FwcHMgTkFTQQ/url-collector/app"
	"github.com/nduni/S3J5c3RpYW4gR29nb0FwcHMgTkFTQQ/url-collector/config"
)

func main() {
	if err := config.LoadConfig(); err != nil {
		panic(err)
	}
	routes := app.GenerateRoutes()
	router := app.NewRouter(routes)
	router.Run(":" + config.Config.Port)
}
