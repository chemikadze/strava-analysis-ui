package cache

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

// datastore-based activity cache

const DATASTORE_PAGE_SIZE = 50

type DatastoreActivityCache struct {
	ctx context.Context
}

func NewDatastoreActivityCache(ctx context.Context) ActivityCache {
	return &DatastoreActivityCache{ctx}
}

type DatastoreJsonEntity struct {
	JsonPayload string `datastore:",noindex"`
}

type PagedEntityMetadata struct {
	PageCount int
}

func (c *DatastoreActivityCache) storeEntity(entityName string, id interface{}, entity interface{}) {
	k := datastore.NewKey(c.ctx, entityName, fmt.Sprintf("%v", id), 0, nil)

	data, err := json.Marshal(entity)
	if err != nil {
		panic(err.Error())
	}
	e := new(DatastoreJsonEntity)
	e.JsonPayload = string(data)
	if _, err := datastore.Put(c.ctx, k, e); err != nil {
		panic(err.Error())
	}
}

func (c *DatastoreActivityCache) retrieveEntity(entityName string, id interface{}, entity interface{}) (found bool) {
	k := datastore.NewKey(c.ctx, entityName, fmt.Sprintf("%v", id), 0, nil)
	e := new(DatastoreJsonEntity)
	if err := datastore.Get(c.ctx, k, e); err != nil {
		return false
	} else {
		err = json.Unmarshal([]byte(e.JsonPayload), &entity)
		if err != nil {
			panic(err.Error())
		}
		return true
	}
}

func pageId(entityId interface{}, pageId int) string {
	return fmt.Sprintf("%v_page%v", entityId, pageId)
}

func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func (c *DatastoreActivityCache) Store(athleteId int64, activities ActivityList) {
	pageCount := len(activities)/DATASTORE_PAGE_SIZE + 1
	for pageNum := 1; pageNum <= pageCount; pageNum++ {
		start := DATASTORE_PAGE_SIZE * (pageNum - 1)
		end := min(DATASTORE_PAGE_SIZE*pageNum, len(activities))
		log.Debugf(c.ctx, "Storing ActivityList page %s", pageId(athleteId, pageNum))
		c.storeEntity("ActivityList", pageId(athleteId, pageNum), activities[start:end])
	}
	c.storeEntity("ActivityList", athleteId, PagedEntityMetadata{PageCount: pageCount})
}

func (c *DatastoreActivityCache) Get(athleteId int64) (ActivityList, bool) {
	var metadata PagedEntityMetadata
	if c.retrieveEntity("ActivityList", athleteId, &metadata) {
		var activities ActivityList
		for pageNum := 1; pageNum <= metadata.PageCount; pageNum++ {
			log.Debugf(c.ctx, "Loading ActivityList page %s", pageId(athleteId, pageNum))
			var pageActivities ActivityList
			if c.retrieveEntity("ActivityList", pageId(athleteId, pageNum), &pageActivities) {
				activities = append(activities, pageActivities...)
			} else {
				log.Warningf(c.ctx, "Found broken paged ActivityList: %s did not have page %s", athleteId, pageNum)
				return nil, false
			}
		}
		return activities, true
	} else {
		return nil, false
	}
}

func (c *DatastoreActivityCache) GetActivity(activityId int64) (*ExtendedActivityInfo, bool) {
	var info ExtendedActivityInfo
	if c.retrieveEntity("Activity", activityId, &info) {
		return &info, true
	} else {
		return nil, false
	}
}

func (c *DatastoreActivityCache) StoreActivity(activityId int64, activity *ExtendedActivityInfo) {
	c.storeEntity("Activity", activityId, activity)
}
