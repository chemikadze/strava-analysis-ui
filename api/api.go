package api

import (
	"encoding/json"
	"fmt"
	"github.com/strava/go.strava"
	"log"
	"net/http"
	"strconv"
)

var pageSize = 200

type Params struct {
	RootUrl      string
	ClientId     int
	ClientSecret string
}

type AnalysisApi struct {
	ActivityCache *ActivityCache
}

func NewApi() *AnalysisApi {
	cache := NewActivityCache()
	return &AnalysisApi{
		cache,
	}
}

func (api *AnalysisApi) AttachHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/activities", api.getActivities)
}

func (api *AnalysisApi) getActivities(w http.ResponseWriter, r *http.Request) {
	// YOLO error handling
	tokenCookie, _ := r.Cookie(cookieStravaToken)
	token := tokenCookie.Value
	athleteCookie, _ := r.Cookie(cookieAthleteId)
	athleteIdInt, _ := strconv.Atoi(athleteCookie.Value)
	athleteId := int64(athleteIdInt)
	client := strava.NewClient(token)
	athletes := strava.NewAthletesService(client)
	fullActivities := make(ActivityList, 0)
	if cached, ok := api.ActivityCache.Get(athleteId); ok {
		fullActivities = cached
	} else {
		for page := 1; ; page++ {
			call := athletes.ListActivities(athleteId)
			call.PerPage(pageSize)
			call.Page(page)
			log.Printf("Loading athlete %s page %s", athleteId, page)
			activities, err := call.Do()
			if err != nil {
				log.Fatal(err)
				w.Write([]byte(err.Error())) // TODO
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
