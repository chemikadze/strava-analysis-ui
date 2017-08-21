package cache

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
