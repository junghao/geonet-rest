#!/bin/bash

#
# This file is auto generated.  Do not edit.
#
# It was created from the JSON config file and shows the env var that can be used to config the app.
# The docker run command will set the env vars on the container.
# You will need to adjust the image name in the Docker command.
#
# The values shown for the env var are the app defaults from the JSON file.
#
# database host name.
# GEONET_REST_DATABASE_HOST=localhost
#
# database User password (unencrypted).
# GEONET_REST_DATABASE_PASSWORD=test
#
# usually disable or require.
# GEONET_REST_DATABASE_SSL_MODE=disable
#
# database connection pool.
# GEONET_REST_DATABASE_MAX_OPEN_CONNS=30
#
# database connection pool.
# GEONET_REST_DATABASE_MAX_IDLE_CONNS=20
#
# web server port.
# GEONET_REST_WEB_SERVER_PORT=8080
#
# public CNAME for the service.
# GEONET_REST_WEB_SERVER_CNAME=localhost
#
# true if the app is production.
# GEONET_REST_WEB_SERVER_PRODUCTION=false
#
# username for Librato.
# LIBRATO_USER=XXX
#
# key for Librato.
# LIBRATO_KEY=XXX
#
# token for Logentries.
# LOGENTRIES_TOKEN=XXX

docker run -e "GEONET_REST_DATABASE_HOST=localhost" -e "GEONET_REST_DATABASE_PASSWORD=test" -e "GEONET_REST_DATABASE_SSL_MODE=disable" -e "GEONET_REST_DATABASE_MAX_OPEN_CONNS=30" -e "GEONET_REST_DATABASE_MAX_IDLE_CONNS=20" -e "GEONET_REST_WEB_SERVER_PORT=8080" -e "GEONET_REST_WEB_SERVER_CNAME=localhost" -e "GEONET_REST_WEB_SERVER_PRODUCTION=false" -e "LIBRATO_USER=XXX" -e "LIBRATO_KEY=XXX" -e "LOGENTRIES_TOKEN=XXX" busybox
