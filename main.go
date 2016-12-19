package main

import (
	"api"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
	"net/http"
	"os"
	"strconv"
)

func init() {
	clientId, _ := strconv.Atoi(os.Getenv("STRAVA_CLIENT_ID"))
	rootUrl := "http://localhost:8080"
	if envRootUrl := os.Getenv("ROOT_URL"); len(envRootUrl) != 0 {
		rootUrl = envRootUrl
	}
	params := api.Params{
		rootUrl,
		clientId,
		os.Getenv("STRAVA_CLIENT_SECRET"),
		func(r *http.Request) *http.Client { return urlfetch.Client(appengine.NewContext(r)) },
	}
	apiService := api.NewApi(params)
	appService := api.NewApp(params)
	apiService.AttachHandlers(http.DefaultServeMux)
	appService.AttachHandlers(http.DefaultServeMux)
}

func main() {
	appengine.Main()
}
