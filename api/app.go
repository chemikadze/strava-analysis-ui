package api

import (
	"fmt"
	"github.com/chemikadze/strava-analysis-ui/templates"
	"github.com/chemikadze/strava-analysis-ui/ui/static"
	"github.com/strava/go.strava"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
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

func callbackUrl(rootUrl string) string {
	return rootUrl + "/exchange_token"
}

func NewApp(params Params) *AnalysisApp {
	auth := &strava.OAuthAuthenticator{
		CallbackURL:            callbackUrl(params.RootUrl),
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

func (app *AnalysisApp) AttachHandlers(mux *http.ServeMux) {
	path, err := app.auth.CallbackPath()
	if err != nil {
		panic(err)
	}

	mux.HandleFunc("/", app.getIndex)
	mux.HandleFunc("/logout", app.getLogout)
	mux.Handle("/static/", NewStaticServer(app.Params.StaticServerType))
	mux.HandleFunc(path, app.auth.HandlerFunc(app.oAuthSuccess, app.oAuthFailure))
}

func (app *AnalysisApp) graphFromRequest(r *http.Request) string {
	query := r.URL.Query()
	graph, ok := query["graph"]
	if !ok {
		graph = []string{"distance-time"}
	}
	return fmt.Sprintf("/static/graphs/%s.js", graph[0])
}

func (app *AnalysisApp) getTemplateContext(r *http.Request) templateContext {
	athleteIdCookie, err := r.Cookie(cookieAthleteId)
	athleteNameCookie, _ := r.Cookie(cookieAthleteName)
	if err != nil || len(athleteIdCookie.Value) == 0 {
		ctx := appengine.NewContext(r)
		defaultHostname, _ := appengine.ModuleHostname(ctx, "", "", "")
		log.Debugf(ctx, "Default hostname: %s", defaultHostname)
		// TODO: not thread-safe
		if !strings.Contains(app.auth.CallbackURL, defaultHostname) {
			app.auth.CallbackURL = callbackUrl("http://" + defaultHostname)
		}
		return templateContext{
			LoggedIn:  false,
			LoginLink: app.auth.AuthorizationURL("state1", strava.Permissions.ViewPrivate, true),
		}
	} else {
		return templateContext{
			LoggedIn:        true,
			AthleteName:     athleteNameCookie.Value,
			AthleteId:       forceAtoI64(athleteIdCookie.Value),
			GraphScriptLink: app.graphFromRequest(r),
		}
	}
}

func NewStaticServer(serverType string) http.Handler {
	if serverType == FILE_STATIC {
		return http.StripPrefix("/static/", http.FileServer(http.Dir("ui/static")))
	} else if serverType == RESOURCE_STATIC {
		return http.StripPrefix("/static/", &StaticServer{})
	} else {
		return nil
	}
}

type StaticServer struct{}

func (StaticServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	log.Infof(ctx, "Seen: %v", r.URL)
	asset, err := static.Asset(r.URL.Path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Write(asset)
}

func parseTemplateResources(names ...string) (result *template.Template) {
	result = template.New(filepath.Base(names[0]))
	for _, filename := range names {
		name := filepath.Base(filename)
		var templ *template.Template
		if result.Name() == name {
			templ = result
		} else {
			templ = result.New(name)
		}
		asset, err := templates.Asset(filename)
		if err != nil {
			panic(fmt.Sprintf("Can not locate template %s", filename))
		}
		_, err = templ.Parse(string(asset))
		if err != nil {
			panic(fmt.Sprintf("Can not parse template: %s", err.Error()))
		}
	}
	return result
}

func (app *AnalysisApp) getIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	ctx := app.getTemplateContext(r)
	template := parseTemplateResources("templates/main.html", "templates/index.html")

	err := template.ExecuteTemplate(w, "main", ctx)
	if err != nil {
		fmt.Fprintf(w, "<h1>Oops, something went wrong</h1>%s", err.Error())
	}
}

func (app *AnalysisApp) oAuthSuccess(auth *strava.AuthorizationResponse, w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: cookieStravaToken, Value: auth.AccessToken})
	http.SetCookie(w, &http.Cookie{Name: cookieAthleteName, Value: auth.Athlete.FirstName})
	http.SetCookie(w, &http.Cookie{Name: cookieAthleteId, Value: strconv.Itoa(int(auth.Athlete.Id))})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *AnalysisApp) oAuthFailure(err error, w http.ResponseWriter, r *http.Request) {
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

func (app *AnalysisApp) getLogout(w http.ResponseWriter, r *http.Request) {
	for _, cookie := range r.Cookies() {
		http.SetCookie(w, &http.Cookie{Name: cookie.Name, Value: "", Expires: epoch})
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
