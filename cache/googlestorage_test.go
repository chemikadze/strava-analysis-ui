package cache

import (
	"testing"
)

func TestObjectKeyFormat(t *testing.T) {
	cache := GoogleStorageActivityCache{}
	activityId := int64(1235)
	expected := "activities/1235/activity.json"
	actual := cache.activityFilename(activityId)
	if expected != actual {
		t.Errorf("%s != %s", expected, actual)
	}
}

func TestPrefixedObjectKeyFormat(t *testing.T) {
	cache := GoogleStorageActivityCache{
		cacheRoot: "my/object/prefix",
	}
	activityId := int64(1235)
	expected := "my/object/prefix/activities/1235/activity.json"
	actual := cache.activityFilename(activityId)
	if expected != actual {
		t.Errorf("%s != %s", expected, actual)
	}
}

func TestPrefixedWithSlashObjectKeyFormat(t *testing.T) {
	cache := GoogleStorageActivityCache{
		cacheRoot: "my/object/prefix/",
	}
	activityId := int64(1235)
	expected := "my/object/prefix/activities/1235/activity.json"
	actual := cache.activityFilename(activityId)
	if expected != actual {
		t.Errorf("%s != %s", expected, actual)
	}
}
