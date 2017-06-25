# Strava graphs UI

Simple strava client webapp in Golang + d3.js graphs.

# Compiling

    make deps
    make 

# Running dev server

Requires appengine SDK (dev_appserver.py) to be installed:

    make localserver

# Deploying to Appengine

Requires gcloud to be installed:

    make deploy