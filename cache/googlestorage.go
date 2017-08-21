package cache

import (
	"cloud.google.com/go/storage"
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"io/ioutil"
	"path"
)

// datastore-based activity cache

type GoogleStorageActivityCache struct {
	ctx        context.Context
	bucketName string
	cacheRoot  string
}

func NewGoogleStorageActivityCache(ctx context.Context, bucketName string, prefix string) ActivityCache {
	return &GoogleStorageActivityCache{ctx, bucketName, prefix}
}

func (c *GoogleStorageActivityCache) activityListFilename(athleteId int64) string {
	return path.Join(c.cacheRoot, fmt.Sprintf("users/%v/activity_list.json", athleteId))
}

func (c *GoogleStorageActivityCache) activityFilename(activityId int64) string {
	return path.Join(
		c.cacheRoot,
		fmt.Sprintf("activities/%v/activity.json", activityId))
}

func (c *GoogleStorageActivityCache) storeAtPath(path string, goObject interface{}) {
	data, err := json.Marshal(goObject)
	if err != nil {
		panic(err.Error())
	}
	client, err := storage.NewClient(c.ctx)
	defer client.Close()
	if err != nil {
		panic(err.Error())
	}
	bucket := client.Bucket(c.bucketName)
	object := bucket.Object(path)
	writer := object.NewWriter(c.ctx)
	if _, err = writer.Write(data); err != nil {
		writer.Close()
		if err := object.Delete(c.ctx); err != nil {
			panic(fmt.Sprintf("Failed to delete cache object after unsuccessful write: %s", err.Error()))
		} else {
			panic(fmt.Sprintf("Failed to save cache object, it have been erased from cache. Original error: %s", err.Error()))
		}
	} else {
		writer.Close()
	}

}

func (c *GoogleStorageActivityCache) getFromPath(path string, goObject interface{}) bool {
	client, err := storage.NewClient(c.ctx)
	defer client.Close()
	if err != nil {
		panic(err.Error())
	}
	bucket := client.Bucket(c.bucketName)
	object := bucket.Object(path)
	reader, err := object.NewReader(c.ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return false
		} else {
			panic(err.Error())
		}
	}
	defer reader.Close()
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err.Error())
	}
	if err := json.Unmarshal([]byte(data), goObject); err != nil {
		panic(err.Error())
	}
	return true
}

func (c *GoogleStorageActivityCache) Store(athleteId int64, activities ActivityList) {
	path := c.activityListFilename(athleteId)
	c.storeAtPath(path, &activities)
}

func (c *GoogleStorageActivityCache) Get(athleteId int64) (ActivityList, bool) {
	path := c.activityListFilename(athleteId)
	activities := make(ActivityList, 0)
	ok := c.getFromPath(path, &activities)
	return activities, ok
}

func (c *GoogleStorageActivityCache) StoreActivity(activityId int64, activity *ExtendedActivityInfo) {
	path := c.activityFilename(activityId)
	c.storeAtPath(path, activity)
}

func (c *GoogleStorageActivityCache) GetActivity(activityId int64) (*ExtendedActivityInfo, bool) {
	path := c.activityFilename(activityId)
	var info ExtendedActivityInfo
	ok := c.getFromPath(path, &info)
	return &info, ok
}
