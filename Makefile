default: bindata test
	go build ./appengine/default/main.go
.PHONY: default

deps:
	go get github.com/jteeuwen/go-bindata/go-bindata
	go get github.com/strava/go.strava
	go get google.golang.org/appengine
.PHONY: deps

templates/bindata.go: templates/*.html
	go-bindata -o templates/bindata.go -pkg templates templates/*

ui/static/bindata.go: ui/static/*/*.js
	go-bindata -o ui/static/bindata.go -prefix ui/static/ -pkg static ui/static/ ui/static/*

bindata: templates/bindata.go ui/static/bindata.go
.PHONY: bindata

test: bindata
	go test ./cache/ ./api/ ./appengine/default/
.PHONY: test

deploy: bindata	test
	envsubst < appengine/default/app.yaml > appengine/default/app.expanded.yaml && \
	gcloud app deploy appengine/default/app.expanded.yaml
.PHONY: deploy

clean:
	rm -f main
	rm -f templates/bindata.go
	rm -f ui/static/bindata.go
.PHONY: clean

localserver: bindata test
	envsubst < appengine/default/app.yaml > appengine/default/app.expanded.yaml && \
	dev_appserver.py ${ADDITIONAL_LOCALSERVER_PARAMS} --log_level debug appengine/default/app.expanded.yaml
.PHONY: localserver
