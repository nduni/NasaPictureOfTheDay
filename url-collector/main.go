package main

import "github.com/nduni/S3J5c3RpYW4gR29nb0FwcHMgTkFTQQ/url-collector/app"

func main() {
	routes := app.GenerateRoutes()
	router := app.NewRouter(routes)
	router.Run(":8080")
}
