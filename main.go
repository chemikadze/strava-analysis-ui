package main

import (
	"github.com/chemikadze/strava-analysis-ui/api"
	"net/http"
	"os"
	"strconv"
)

func main() {
	clientId, _ := strconv.Atoi(os.Getenv("STRAVA_CLIENT_ID"))
	params := api.Params{
		"http://localhost:8080",
		clientId,
		os.Getenv("STRAVA_CLIENT_SECRET"),
	}
	apiService := api.NewApi()
	appService := api.NewApp(params)
	mux := http.NewServeMux()
	apiService.AttachHandlers(mux)
	appService.AttachHandlers(mux)
	http.ListenAndServe(":8080", mux)
	return
}
