package api

import (
	"encoding/json"
	"fmt"
	"github.com/strava/go.strava"
	"io/ioutil"
	"log"
	"os"
	"path"
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
	if DEFAULT_CACHE_IMPL == "memory" {
		return NewMapActivityCache()
	} else if DEFAULT_CACHE_IMPL == "file" {
		return NewFileActivityCache("")
	} else {
		panic("Unknown cache impl: " + DEFAULT_CACHE_IMPL)
	}
}

// in-memory map activity cache

type MapActivityCache struct {
	activityLists   map[int64]ActivityList
	activityDetails map[int64]*ExtendedActivityInfo
}

func NewMapActivityCache() ActivityCache {
	var cache MapActivityCache
	cache.activityLists = make(map[int64]ActivityList)
	cache.activityDetails = make(map[int64]*ExtendedActivityInfo)
	return &cache
}

func (c *MapActivityCache) Store(athleteId int64, activities ActivityList) {
	c.activityLists[athleteId] = activities
}

func (c *MapActivityCache) Get(athleteId int64) (ActivityList, bool) {
	if activities, ok := c.activityLists[athleteId]; ok {
		return activities, true
	} else {
		return nil, false
	}
}

func (c *MapActivityCache) GetActivity(activityId int64) (*ExtendedActivityInfo, bool) {
	if activity, ok := c.activityDetails[activityId]; ok {
		return activity, true
	} else {
		return nil, false
	}
}

func (c *MapActivityCache) StoreActivity(activityId int64, activity *ExtendedActivityInfo) {
	c.activityDetails[activityId] = activity
}

// file-based activity cache

type FileActivityCache struct {
	cacheRoot string
}

func NewFileActivityCache(cacheRoot string) ActivityCache {
	return &FileActivityCache{cacheRoot}
}

func (c *FileActivityCache) activityListFilename(athleteId int64) string {
	return path.Join(c.cacheRoot, fmt.Sprintf("users/%v/activity_list.json", athleteId))
}

func (c *FileActivityCache) activityFilename(activityId int64) string {
	return path.Join(
		c.cacheRoot,
		fmt.Sprintf("activities/%v/activity.json", activityId))
}

func (c *FileActivityCache) Store(athleteId int64, activities ActivityList) {
	filename := c.activityListFilename(athleteId)
	log.Printf("Storing activity list: %s", filename)
	data, err := json.Marshal(activities)
	if err != nil {
		panic(err.Error())
	}
	if os.MkdirAll(path.Dir(filename), DEFAULT_DIR_MODE) != nil {
		panic(err.Error())
	}
	if ioutil.WriteFile(filename, data, DEFAULT_FILE_MODE) != nil {
		panic(err.Error())
	}
}

func (c *FileActivityCache) Get(athleteId int64) (ActivityList, bool) {
	data, err := ioutil.ReadFile(c.activityListFilename(athleteId))
	if err != nil {
		return nil, false
	}

	var activities ActivityList
	err = json.Unmarshal(data, &activities)
	if err != nil {
		panic(err.Error())
	}
	return activities, true
}

func (c *FileActivityCache) GetActivity(activityId int64) (*ExtendedActivityInfo, bool) {
	data, err := ioutil.ReadFile(c.activityFilename(activityId))
	if err != nil {
		return nil, false
	}

	var activity ExtendedActivityInfo
	if json.Unmarshal(data, &activity) != nil {
		panic(err.Error())
	}
	return &activity, true
}

func (c *FileActivityCache) StoreActivity(activityId int64, activity *ExtendedActivityInfo) {
	filename := c.activityFilename(activityId)
	data, err := json.Marshal(activity)
	if err != nil {
		panic(err.Error())
	}
	if os.MkdirAll(path.Dir(filename), DEFAULT_DIR_MODE) != nil {
		panic(err.Error())
	}
	if ioutil.WriteFile(filename, data, DEFAULT_FILE_MODE) != nil {
		panic(err.Error())
	}
}
