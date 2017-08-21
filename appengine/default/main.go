package main

import (
	"fmt"
	"github.com/chemikadze/strava-analysis-ui/api"
	"github.com/chemikadze/strava-analysis-ui/cache"
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

func newCacheFactory() func(ctx context.Context) cache.ActivityCache {
	log.Printf("Using cache impl: %s", cache.DEFAULT_CACHE_IMPL)
	if cache.DEFAULT_CACHE_IMPL == "memory" {
		instance := cache.NewMapActivityCache()
		return func(ctx context.Context) cache.ActivityCache { return instance }
	} else if cache.DEFAULT_CACHE_IMPL == "file" {
		instance := cache.NewDefaultFileActivityCache()
		return func(ctx context.Context) cache.ActivityCache { return instance }
	} else if cache.DEFAULT_CACHE_IMPL == "datastore" {
		return func(ctx context.Context) cache.ActivityCache { return cache.NewDatastoreActivityCache(ctx) }
	} else if cache.DEFAULT_CACHE_IMPL == "googlestorage" {
		bucket := getEnvOrPanic("STRAVA_CACHE_BUCKET", "")
		prefix := os.Getenv("STRAVA_CACHE_PREFIX")
		return func(ctx context.Context) cache.ActivityCache {
			return cache.NewGoogleStorageActivityCache(ctx, bucket, prefix)
		}
	} else {
		panic("Unknown cache impl: " + cache.DEFAULT_CACHE_IMPL)
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
	staticServerType := getEnvOrPanic("STATIC_SERVER_TYPE", api.RESOURCE_STATIC)

	params := api.Params{
		rootUrl,
		clientId,
		clientSecret,
		resolveUrlFetchFunc,
		newCacheFactory(),
		zonesEnabled,
		staticServerType,
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
