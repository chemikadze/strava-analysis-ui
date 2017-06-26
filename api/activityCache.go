package api

import (
	"github.com/strava/go.strava"
)

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
type MapActivityCache struct {
	activityLists   map[int64]ActivityList
	activityDetails map[int64]*ExtendedActivityInfo
}

func NewActivityCache() ActivityCache {
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
