package api

import (
	"github.com/strava/go.strava"
)

type ActivityList []*strava.ActivitySummary
type ActivityCache map[int64]ActivityList

func NewActivityCache() *ActivityCache {
	cache := make(ActivityCache)
	return &cache
}

func (c *ActivityCache) Store(athleteId int64, activities ActivityList) {
	(*c)[athleteId] = activities
}

func (c *ActivityCache) Get(athleteId int64) (ActivityList, bool) {
	if activities, ok := (*c)[athleteId]; ok {
		return activities, true
	} else {
		return nil, false
	}
}
