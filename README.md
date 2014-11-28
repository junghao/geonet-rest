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

```godep go build && ./geonet-rest```

Run all tests (including those in sub dirs):

```godep go test ./...```

### API Design

* URIs should return a resource and the query parameters should be used to filter (search) for them.
* Use ISO8601 date times in UTC e.g., `2013-05-30T15:15:37.812Z`
* Use http methods in routes (`GET`, `PUT` etc).
* Use camelCase for query and property names.  Be consistent with SeisComPML or QuakeML e.g., `publicID` not `publicId` or `publicid`.
* The  http `Accept-Header` should be used to determine which data version and format to return.

### API Documentation

API documentation is generated from doc{} structs in the code.  Run the application and visit `http://localhost:8080/api-docs`.

### API Changes

#### Non Breaking Changes

* Make non breaking **additions** as required.
* Add to the tests.
* Add Markdown documention to the tests and regenerate the API docs.

#### Breaking Changes

* Are you really sure you have to.  Discuss widely.
* Copy the current API verion code to the next API version (so as to support all queries at the new version)
* Monotonically increment the `Accept` constant e.g., `application/vnd.geo+json;version=1 -> application/vnd.geo+json;version=2`
* Change the tests.  
* Update the documentation.  
* Make the changes.  
* Update the routes.  


### Database

Uses the database from the geonet web project.

This tutorial is very useful for database access with Go http://go-database-sql.org/

### Properties

Properties are read from `/etc/sysconfig/geonet-rest.json` and if this is not found then from `./geonet-rest.json` (this should contain testing values)

## Deployment

### Properties 

Copy an appropriately edited version of `geonet-rest.json` to `/etc/sysconfig/geonet-rest.json`  This should include read only credentials for accessing the hazard database.

### Monitoring

Expvar is used to expose counters at http://.../debug/vars.  As well as the Go memstats counters there are counters for resquests and responses e.g.,

```
{
  "responses": {
    "5xx": 0,
    "4xx": 1,
    "2xx": 5001
  },
  "requests": 5002,
  "memstats": {
    "BySize": [
      {
        "Frees": 0,
        "Mallocs": 0,
        "Size": 0
      },
...
```

Fatal application errors, 4xx and 5xx requests are syslogged.

### Lookups.

Some database tables change very rarely e.g., regions.  The data for these is loaded on startup and then cached.
If a new region is added to the database or a region changed then the geonet-rest application will need to be 
restarted to pick up the changes.
