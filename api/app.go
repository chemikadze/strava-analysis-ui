package api

import (
	"fmt"
	"github.com/strava/go.strava"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var cookieStravaToken = "strava-token"
var cookieAthleteId = "athlete-id"
var cookieAthleteName = "athlete-name"

var epoch = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

type AnalysisApp struct {
	Params Params
	auth   *strava.OAuthAuthenticator
}

type templateContext struct {
	LoggedIn        bool
	AthleteName     string
	AthleteId       int64
	LoginLink       string
	GraphScriptLink string
}

func NewApp(params Params) *AnalysisApp {
	auth := &strava.OAuthAuthenticator{
		CallbackURL:            params.RootUrl + "/exchange_token",
		RequestClientGenerator: params.RequestClientGenerator,
	}
	strava.ClientId = params.ClientId
	strava.ClientSecret = params.ClientSecret
	return &AnalysisApp{
		params,
		auth,
	}
}

func forceAtoI64(s string) int64 {
	x, _ := strconv.Atoi(s)
	return int64(x)
}

func (api *AnalysisApp) AttachHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/", api.getIndex)
	mux.HandleFunc("/logout", api.getLogout)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("ui/static"))))

	path, err := api.auth.CallbackPath()
	if err != nil {
		// possibly that the callback url set above is invalid
		fmt.Println(err)
		os.Exit(1)
	}
	mux.HandleFunc(path, api.auth.HandlerFunc(api.oAuthSuccess, api.oAuthFailure))
}

func (api *AnalysisApp) graphFromRequest(r *http.Request) string {
	query := r.URL.Query()
	graph, ok := query["graph"]
	if !ok {
		graph = []string{"distance-time"}
	}
	return fmt.Sprintf("/static/graphs/%s.js", graph[0])
}

func (api *AnalysisApp) getTemplateContext(r *http.Request) templateContext {
	athleteIdCookie, err := r.Cookie(cookieAthleteId)
	athleteNameCookie, _ := r.Cookie(cookieAthleteName)
	if err != nil || len(athleteIdCookie.Value) == 0 {
		return templateContext{
			LoggedIn:  false,
			LoginLink: api.auth.AuthorizationURL("state1", strava.Permissions.Public, true),
		}
	} else {
		return templateContext{
			LoggedIn:        true,
			AthleteName:     athleteNameCookie.Value,
			AthleteId:       forceAtoI64(athleteIdCookie.Value),
			GraphScriptLink: api.graphFromRequest(r),
		}
	}
}

func (api *AnalysisApp) getIndex(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	ctx := api.getTemplateContext(r)
	template, _ := template.ParseFiles("templates/main.html", "templates/index.html")
	err := template.ExecuteTemplate(w, "main", ctx)
	if err != nil {
		fmt.Fprintf(w, "<h1>Oops, something went wrong</h1>%s", err.Error())
	}
}

func (api *AnalysisApp) oAuthSuccess(auth *strava.AuthorizationResponse, w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: cookieStravaToken, Value: auth.AccessToken})
	http.SetCookie(w, &http.Cookie{Name: cookieAthleteName, Value: auth.Athlete.FirstName})
	http.SetCookie(w, &http.Cookie{Name: cookieAthleteId, Value: strconv.Itoa(int(auth.Athlete.Id))})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (api *AnalysisApp) oAuthFailure(err error, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Authorization Failure:\n")

	// some standard error checking
	if err == strava.OAuthAuthorizationDeniedErr {
		fmt.Fprint(w, "The user clicked the 'Do not Authorize' button on the previous page.\n")
		fmt.Fprint(w, "This is the main error your application should handle.")
	} else if err == strava.OAuthInvalidCredentialsErr {
		fmt.Fprint(w, "You provided an incorrect client_id or client_secret.\nDid you remember to set them at the begininng of this file?")
	} else if err == strava.OAuthInvalidCodeErr {
		fmt.Fprint(w, "The temporary token was not recognized, this shouldn't happen normally")
	} else if err == strava.OAuthServerErr {
		fmt.Fprint(w, "There was some sort of server error, try again to see if the problem continues")
	} else {
		fmt.Fprint(w, err)
	}
}

func (api *AnalysisApp) getLogout(w http.ResponseWriter, r *http.Request) {
	for _, cookie := range r.Cookies() {
		http.SetCookie(w, &http.Cookie{Name: cookie.Name, Value: "", Expires: epoch})
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
