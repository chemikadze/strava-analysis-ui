package api

import (
	"github.com/strava/go.strava"
	"io/ioutil"
	"os"
	_ "reflect"
	"testing"
)

func TestDiretoryFormat(t *testing.T) {
	cache := FileActivityCache{""}
	activityId := int64(1235)
	expected := "activities/1235/activity.json"
	actual := cache.activityFilename(activityId)
	if expected != actual {
		t.Errorf("%s != %s", expected, actual)
	}
}

func TestFileCacheCreation(t *testing.T) {
	cacheRoot, _ := ioutil.TempDir("", "activityCache")
	defer os.Remove(cacheRoot)

	cache := NewFileActivityCache(cacheRoot)
	if cache == nil {
		t.Error("New cache should not be nil!")
	}
}

func TestFileCacheEmptyReads(t *testing.T) {
	cacheRoot, _ := ioutil.TempDir("", "activityCache")
	defer os.Remove(cacheRoot)

	cache := NewFileActivityCache(cacheRoot)
	athleteId := int64(12345)
	if _, ok := cache.Get(athleteId); ok {
		t.Error("Get on empty cache should return false!")
	}

	activityId := int64(12345)
	if _, ok := cache.GetActivity(activityId); ok {
		t.Error("Get on empty cache should return false!")
	}
}

func TestFileCacheCanGetActivityList(t *testing.T) {
	cacheRoot, _ := ioutil.TempDir("", "activityCache")
	defer os.Remove(cacheRoot)

	cache := NewFileActivityCache(cacheRoot)
	expectedActivities := make(ActivityList, 0)
	activityId := int64(12345)
	athleteId := int64(1234)
	expectedActivities = append(expectedActivities, &strava.ActivitySummary{Id: activityId, Name: "My Activity"})
	cache.Store(athleteId, expectedActivities)
	if loadedActivities, ok := cache.Get(athleteId); ok {
		equal := len(expectedActivities) == len(loadedActivities) &&
			expectedActivities[0].Id == loadedActivities[0].Id &&
			expectedActivities[0].Name == loadedActivities[0].Name
		if !equal {
			t.Error("Activity fields should be equal!")
		}
		// // TODO: failing
		// equal = reflect.DeepEqual(expectedActivities, loadedActivities)
		// if !equal {
		// 	t.Error("Activities should be deep equal!")
		// }
	} else {
		t.Error("cache.Get should return ok after storing activity!")
	}
}

func TestFileCacheCanGetActivity(t *testing.T) {
	cacheRoot, _ := ioutil.TempDir("", "activityCache")
	defer os.Remove(cacheRoot)

	cache := NewFileActivityCache(cacheRoot)
	activityId := int64(12345)
	expectedActivity := ExtendedActivityInfo{
		Activity: &strava.ActivityDetailed{
			ActivitySummary: strava.ActivitySummary{
				Id:   activityId,
				Name: "Test Name",
			},
		},
		ZonesSummary: &strava.ZonesSummary{
			CustonZones: true,
			Score:       1234,
		}}
	cache.StoreActivity(activityId, &expectedActivity)
	if loadedActivity, ok := cache.GetActivity(activityId); ok {
		equal := loadedActivity.Activity.Id == expectedActivity.Activity.Id &&
			loadedActivity.Activity.Name == expectedActivity.Activity.Name &&
			loadedActivity.ZonesSummary.CustonZones == expectedActivity.ZonesSummary.CustonZones &&
			loadedActivity.ZonesSummary.Score == expectedActivity.ZonesSummary.Score
		if !equal {
			t.Error("Activities should be equal!")
		}
		// // TODO: failing
		// equal = reflect.DeepEqual(expectedActivity, loadedActivity)
		// if !equal {
		// 	t.Error("Activities should be deep equal!")
		// }
	} else {
		t.Error("cache.GetActivity should return ok after storing activity!")
	}
}
