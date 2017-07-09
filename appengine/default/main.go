package main

import (
	"fmt"
	"github.com/chemikadze/strava-analysis-ui/api"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func getEnvOrPanic(name, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if len(value) == 0 {
		if defaultValue == "" {
			panic(fmt.Sprintf("Expected %s environment variable", name))
		} else {
			value = defaultValue
		}
	}
	return value
}

func resolveUrlFetchFunc(r *http.Request) *http.Client {
	enabledVar := os.Getenv("APPENGINE_ENABLED")
	if strings.ToLower(enabledVar) == "true" || enabledVar == "1" {
		return urlfetch.Client(appengine.NewContext(r))
	} else {
		return http.DefaultClient
	}
}

func newCacheFactory() func(ctx context.Context) api.ActivityCache {
	log.Printf("Using cache impl: %s", api.DEFAULT_CACHE_IMPL)
	if api.DEFAULT_CACHE_IMPL == "memory" {
		cache := api.NewMapActivityCache()
		return func(ctx context.Context) api.ActivityCache { return cache }
	} else if api.DEFAULT_CACHE_IMPL == "file" {
		cache := api.NewDefaultFileActivityCache()
		return func(ctx context.Context) api.ActivityCache { return cache }
	} else if api.DEFAULT_CACHE_IMPL == "datastore" {
		return func(ctx context.Context) api.ActivityCache { return api.NewDatastoreActivityCache(ctx) }
	} else {
		panic("Unknown cache impl: " + api.DEFAULT_CACHE_IMPL)
	}
}

func init() {
	clientId, _ := strconv.Atoi(getEnvOrPanic("STRAVA_CLIENT_ID", ""))
	if clientId == 0 {
		panic("STRAVA_CLIENT_ID should be set to non-zero value")
	}
	clientSecret := getEnvOrPanic("STRAVA_CLIENT_SECRET", "")
	rootUrl := getEnvOrPanic("ROOT_URL", "http://localhost:8080")
	zonesEnabled := getEnvOrPanic("STRAVA_ZONES_ENABLED", "false") == "true"

	params := api.Params{
		rootUrl,
		clientId,
		clientSecret,
		resolveUrlFetchFunc,
		newCacheFactory(),
		zonesEnabled,
	}
	apiService := api.NewApi(params)
	appService := api.NewApp(params)
	apiService.AttachHandlers(http.DefaultServeMux)
	appService.AttachHandlers(http.DefaultServeMux)
}

func main() {
	log.Println("Starting server...")
	appengine.Main()
}
