package api

import (
	"encoding/json"
	"fmt"
	"github.com/strava/go.strava"
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

func NewApi(params Params) *AnalysisApi {
	cache := NewActivityCache()
	return &AnalysisApi{
		cache,
		params,
	}
}

func (api *AnalysisApi) AttachHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/activities", api.getActivities)
}

func (api *AnalysisApi) getActivities(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(w, r) // TODO proper json response
		}
	}()

	// YOLO error handling
	tokenCookie, _ := r.Cookie(cookieStravaToken)
	token := tokenCookie.Value
	athleteCookie, _ := r.Cookie(cookieAthleteId)
	athleteIdInt, _ := strconv.Atoi(athleteCookie.Value)
	athleteId := int64(athleteIdInt)
	var client *strava.Client
	if api.Params.RequestClientGenerator != nil {
		httpClient := api.Params.RequestClientGenerator(r)
		client = strava.NewClient(token, httpClient)
	} else {
		client = strava.NewClient(token)
	}

	athletes := strava.NewAthletesService(client)
	fullActivities := make(ActivityList, 0)
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
				fmt.Fprintln(w, err.Error()) // TODO proper json response
				return
			}
			if len(activities) == 0 {
				break
			}
			fullActivities = append(fullActivities, activities...)
		}
		api.ActivityCache.Store(athleteId, fullActivities)
	}
	content, _ := json.MarshalIndent(fullActivities, "", " ")
	fmt.Fprint(w, string(content))
}
