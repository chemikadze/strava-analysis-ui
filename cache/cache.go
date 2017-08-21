package cache

import (
	"github.com/strava/go.strava"
	"log"
	"os"
)

var DEFAULT_CACHE_IMPL string
var DEFAULT_CACHE_ROOT string
var DEFAULT_FILE_MODE os.FileMode = 0655
var DEFAULT_DIR_MODE os.FileMode = 0755

func init() {
	DEFAULT_CACHE_ROOT = os.Getenv("STRAVA_CACHE_ROOT")
	if len(DEFAULT_CACHE_ROOT) == 0 {
		DEFAULT_CACHE_ROOT = "."
	}
	DEFAULT_CACHE_IMPL = os.Getenv("STRAVA_CACHE_IMPL")
	if len(DEFAULT_CACHE_IMPL) == 0 {
		DEFAULT_CACHE_IMPL = "memory"
	}
}

type ActivityCache interface {
	// store activity list for user
	Store(int64, ActivityList)

	// get activity list for user, returns (nil, false) if not present
	Get(int64) (ActivityList, bool)

	// put activity into cache by id
	StoreActivity(int64, *ExtendedActivityInfo)

	// get activity by id, returns (nil, false) if not present
	GetActivity(int64) (*ExtendedActivityInfo, bool)
}

type ExtendedActivityInfo struct {
	Activity     *strava.ActivityDetailed
	ZonesSummary *strava.ZonesSummary
}

type ActivityList []*strava.ActivitySummary

func NewActivityCache() ActivityCache {
	log.Println("Using cache impl: %v", DEFAULT_CACHE_IMPL)
	if DEFAULT_CACHE_IMPL == "memory" {
		return NewMapActivityCache()
	} else if DEFAULT_CACHE_IMPL == "file" {
		return NewDefaultFileActivityCache()
	} else {
		panic("Unknown cache impl: " + DEFAULT_CACHE_IMPL)
	}
}
