# geonet-rest

Rest API for GeoNet web site data.


## Development 

Requires Go 1.2.1 or newer (for db.SetMaxOpenConns(n)).

### Dependencies and Compilation

Dependencies are included in this repo using godep vendoring.  There should be no need to `go get` the dependencies 
separately unless you are updating them.

* Install godep (you will need Git and Mercurial installed to do this). https://github.com/tools/godep
* Prefix go commands with godep.

Run:

```godep go run geonet-rest.go```

Run all tests (including those in sub dirs):

```godep go test ./...```

### API Design

* URIs should return a resource and the query parameters should be used to filter (search) for them.
* Use ISO8601 date times in UTC e.g., `2013-05-30T15:15:37.812Z`
* Use http methods in routes (`GET`, `PUT` etc).
* Use camelCase for query and property names.  Be consistent with SeisComPML or QuakeML e.g., `publicID` not `publicId` or `publicid`.
* The  http `Accept-Header` should be used to determine which data version and format to return.
* If there is no `Accept-Header` then route to the latest version of the API. This makes exploring the API with a browser easy.
* Pretty print the response.  This makes exploring the API with a browser easy.  Due to the gzip compression the extra response size is negligible.

### API Documentation

* Write public documentation for the rest API using Github flavoured markdown in the tests for each package. 
* Use scripts/docs to generate the documentation in `api-docs/endpoints`.

### API Changes

#### Non Breaking Changes

* Make non breaking **additions** as required.
* Add to the tests.
* Add Markdown documention to the tests and regenerate the API docs.

#### Breaking Changes

* Are you really sure you have to.  Discuss widely.
* Copy the current API verion code to the next API version (so as to support all queries at the new version)
* Monotonically increment the `Accept` constant e.g., `application/vnd.geo+json; version=1 -> application/vnd.geo+json; version=2`
* Change the tests.  
* Update the documentation.  
* Make the changes.  
* Update the routes.  
* Make sure that if no `Accept-Header` is present requests route to the latest version of the API.


### Database

Uses the database from the geonet web project.

This tutorial is very useful for database access with Go http://go-database-sql.org/

### Properties

Properties are read from `/etc/sysconfig/geonet-rest.json` and if this is not found then from `./geonet-rest.json` (this should contain testing values)

## Deployment

### Properties 

Copy an appropriately edited version of `geonet-rest.json` to `/etc/sysconfig/geonet-rest.json`  This should include read only credentials for accessing the hazard database.