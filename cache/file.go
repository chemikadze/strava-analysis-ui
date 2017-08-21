package cache

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
)

// file-based activity cache

type FileActivityCache struct {
	cacheRoot string
}

func NewDefaultFileActivityCache() ActivityCache {
	log.Println("Using default cache root: %v", DEFAULT_CACHE_ROOT)
	return NewFileActivityCache(DEFAULT_CACHE_ROOT)
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
	log.Printf("Storing activity list: %v", filename)
	data, err := json.Marshal(activities)
	if err != nil {
		panic(err.Error())
	}
	if err := os.MkdirAll(path.Dir(filename), DEFAULT_DIR_MODE); err != nil {
		panic(err.Error())
	}
	if err := ioutil.WriteFile(filename, data, DEFAULT_FILE_MODE); err != nil {
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
	if err := os.MkdirAll(path.Dir(filename), DEFAULT_DIR_MODE); err != nil {
		panic(err.Error())
	}
	if err := ioutil.WriteFile(filename, data, DEFAULT_FILE_MODE); err != nil {
		panic(err.Error())
	}
}
