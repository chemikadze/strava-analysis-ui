package api

import (
	"encoding/json"
	"fmt"
	"github.com/strava/go.strava"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"net/http"
	"strconv"
)

var pageSize = 200

type Params struct {
	RootUrl                string
	ClientId               int
	ClientSecret           string
	RequestClientGenerator func(r *http.Request) *http.Client
}

type AnalysisApi struct {
	ActivityCache ActivityCache
	Params        Params
}

type ZoneInfoResponse struct {
	activities []ActivityZoneInfo
}

type ActivityZoneInfo struct {
	activityInfo *strava.ActivitySummary
	zoneInfo     *strava.ZonesSummary
}

func NewApi(params Params) *AnalysisApi {
	cache := NewActivityCache()
	return &AnalysisApi{
		cache,
		params,
	}
}

func (api *AnalysisApi) AttachHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/activities", api.getActivities)
	// mux.HandleFunc("/zones", api.getZonesData)
}

func (api *AnalysisApi) getStravaClient(r *http.Request) (client *strava.Client) {
	// TODO: YOLO error handling
	tokenCookie, _ := r.Cookie(cookieStravaToken)
	token := tokenCookie.Value
	if api.Params.RequestClientGenerator != nil {
		httpClient := api.Params.RequestClientGenerator(r)
		client = strava.NewClient(token, httpClient)
	} else {
		client = strava.NewClient(token)
	}
	return client
}

func (api *AnalysisApi) getAthleteId(r *http.Request) int64 {
	athleteCookie, _ := r.Cookie(cookieAthleteId)
	athleteIdInt, _ := strconv.Atoi(athleteCookie.Value)
	athleteId := int64(athleteIdInt)
	return athleteId
}

func (api *AnalysisApi) retrieveActivities(ctx context.Context, client *strava.Client, athleteId int64) (fullActivities ActivityList) {
	athletes := strava.NewAthletesService(client)
	fullActivities = make(ActivityList, 0)
	if cached, ok := api.ActivityCache.Get(athleteId); ok {
		fullActivities = cached
	} else {
		for page := 1; ; page++ {
			call := athletes.ListActivities(athleteId)
			call.PerPage(pageSize)
			call.Page(page)
			log.Debugf(ctx, "Loading athlete %s page %s", athleteId, page)
			activities, err := call.Do()
			if err != nil {
				log.Criticalf(ctx, err.Error())
				panic(err.Error())
			}
			if len(activities) == 0 {
				break
			}
			fullActivities = append(fullActivities, activities...)
		}
		api.ActivityCache.Store(athleteId, fullActivities)
	}
	return fullActivities
}

func (api *AnalysisApi) retrieveActivity(ctx context.Context, client *strava.Client, activityId int64) *ExtendedActivityInfo {
	if activity, ok := api.ActivityCache.GetActivity(activityId); ok {
		return activity
	} else {
		log.Debugf(ctx, "did not find activity %s in cache, downloading", activityId)
		activitiesService := strava.NewActivitiesService(client)
		activityCall := activitiesService.Get(activityId)
		activity, err := activityCall.Do()
		if err != nil {
			log.Criticalf(ctx, err.Error())
			panic(err.Error())
		}

		zonesCall := activitiesService.ListZones(activityId)
		zones, err := zonesCall.Do()
		if err != nil {
			log.Criticalf(ctx, err.Error())
			panic(err.Error())
		}

		var hrZone *strava.ZonesSummary

		for _, zone := range zones {
			if zone.Type == "heartrate" {
				hrZone = zone
				break
			}
		}

		return &ExtendedActivityInfo{
			Activity:     activity,
			ZonesSummary: hrZone,
		}
	}
}

func (api *AnalysisApi) getActivities(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(w, r) // TODO proper json response
		}
	}()

	// TODO: YOLO error handling
	athleteId := api.getAthleteId(r)
	client := api.getStravaClient(r)
	fullActivities := api.retrieveActivities(ctx, client, athleteId)
	content, _ := json.MarshalIndent(fullActivities, "", " ")
	fmt.Fprint(w, string(content))
}

func (api *AnalysisApi) getZonesData(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(w, r) // TODO proper json response
		}
	}()

	athleteId := api.getAthleteId(r)
	client := api.getStravaClient(r)
	fullActivities := api.retrieveActivities(ctx, client, athleteId)
	histogramData := make([]ActivityZoneInfo, 0)
	for _, activity := range fullActivities {
		activityExtended := api.retrieveActivity(ctx, client, activity.Id)
		zoneInfo := ActivityZoneInfo{
			activityInfo: activity,
			zoneInfo:     activityExtended.ZonesSummary,
		}
		histogramData = append(histogramData, zoneInfo)
	}
	response := ZoneInfoResponse{
		activities: histogramData,
	}
	content, _ := json.MarshalIndent(response, "", " ")
	fmt.Fprint(w, string(content))
}
