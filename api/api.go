package api

import (
	"encoding/json"
	"fmt"
	"github.com/chemikadze/strava-analysis-ui/cache"
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
	ActivityCacheAccessor  func(ctx context.Context) cache.ActivityCache
	ZonesEnabled           bool
	StaticServerType       string
}

const (
	RESOURCE_STATIC = "bindata"
	FILE_STATIC     = "file"
)

type AnalysisApi struct {
	Params Params
}

type ZoneInfoResponse struct {
	Activities []ActivityZoneInfo
}

type ActivityZoneInfo struct {
	ActivityInfo *strava.ActivitySummary
	ZoneInfo     *strava.ZonesSummary
}

func NewApi(params Params) *AnalysisApi {
	return &AnalysisApi{
		params,
	}
}

func (api *AnalysisApi) AttachHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/activities", api.getActivities)
	if api.Params.ZonesEnabled {
		mux.HandleFunc("/zones", api.getZonesData)
	}
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

func (api *AnalysisApi) retrieveActivities(ctx context.Context, client *strava.Client, athleteId int64) (fullActivities cache.ActivityList) {
	athletes := strava.NewAthletesService(client)
	fullActivities = make(cache.ActivityList, 0)
	cacheClient := api.Params.ActivityCacheAccessor(ctx)
	if cached, ok := cacheClient.Get(athleteId); ok {
		fullActivities = cached
	} else {
		for page := 1; ; page++ {
			call := athletes.ListActivities(athleteId)
			call.PerPage(pageSize)
			call.Page(page)
			log.Debugf(ctx, "Loading athlete %v page %v", athleteId, page)
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
		cacheClient.Store(athleteId, fullActivities)
	}
	return fullActivities
}

func (api *AnalysisApi) retrieveActivity(ctx context.Context, client *strava.Client, activityId int64) (*cache.ExtendedActivityInfo, error) {

	cacheClient := api.Params.ActivityCacheAccessor(ctx)
	if activity, ok := cacheClient.GetActivity(activityId); ok {
		log.Debugf(ctx, "using activity %v from cache", activityId)
		return activity, nil
	} else {
		log.Debugf(ctx, "did not find activity %v in cache, downloading", activityId)
		activitiesService := strava.NewActivitiesService(client)
		activityCall := activitiesService.Get(activityId)
		activity, err := activityCall.Do()
		if err != nil {
			return nil, err
		}

		zonesCall := activitiesService.ListZones(activityId)
		zones, err := zonesCall.Do()
		if err != nil {
			return nil, err
		}

		var hrZone *strava.ZonesSummary

		for _, zone := range zones {
			if zone.Type == "heartrate" {
				hrZone = zone
				break
			}
		}

		activityInfo := cache.ExtendedActivityInfo{
			Activity:     activity,
			ZonesSummary: hrZone,
		}

		cacheClient.StoreActivity(activityId, &activityInfo)

		return &activityInfo, nil
	}
}

func (api *AnalysisApi) getActivities(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	defer func() {
		if r := recover(); r != nil {
			log.Warningf(ctx, "Recovered: %v", r)
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
			log.Warningf(ctx, "Recovered: %v", r)
			fmt.Fprintln(w, r) // TODO proper json response
		}
	}()

	athleteId := api.getAthleteId(r)
	client := api.getStravaClient(r)
	fullActivities := api.retrieveActivities(ctx, client, athleteId)
	histogramData := make([]ActivityZoneInfo, 0)
	for _, activity := range fullActivities {
		if activity.Private {
			continue
		}
		activityExtended, err := api.retrieveActivity(ctx, client, activity.Id)
		if err != nil {
			log.Warningf(ctx, "Failed to retrieve activity %v: %v", activity.Id, err.Error())
			continue
		}
		zoneInfo := ActivityZoneInfo{
			ActivityInfo: activity,
			ZoneInfo:     activityExtended.ZonesSummary,
		}
		histogramData = append(histogramData, zoneInfo)
	}
	response := ZoneInfoResponse{
		Activities: histogramData,
	}
	content, _ := json.MarshalIndent(response, "", " ")
	fmt.Fprint(w, string(content))
}
