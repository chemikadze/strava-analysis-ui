package api

import (
	"github.com/strava/go.strava"
)

type ActivityCache interface {
	Store(int64, ActivityList)
	Get(int64) (ActivityList, bool)
}

type ActivityList []*strava.ActivitySummary
type MapActivityCache map[int64]ActivityList

func NewActivityCache() ActivityCache {
	cache := make(MapActivityCache)
	return &cache
}

func (c *MapActivityCache) Store(athleteId int64, activities ActivityList) {
	(*c)[athleteId] = activities
}

func (c *MapActivityCache) Get(athleteId int64) (ActivityList, bool) {
	if activities, ok := (*c)[athleteId]; ok {
		return activities, true
	} else {
		return nil, false
	}
}
