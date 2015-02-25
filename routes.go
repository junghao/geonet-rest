package main

// To add a route do the following:
// 1. Create a struct to hold the request parameters that are needed for generating content for the request (this will occasionally be empty).
// 2. Implement the api.Query interface:
//     * Implement the Validate method on the struct (this will occasionally simply return true).
//     * Implement the Handle method on the struct to create and write the content.
//     * Implement the Doc method.  This is used for the html docs of the api.  Put the doc struct near the query struct as code documentation.
// 3. Add the route for the request to router.
// 4. When you are ready to publish public html documentation for the query then add it's doc method to the appropriate endpoint documentation.

import (
	"github.com/GeoNet/app/web"
	"github.com/GeoNet/app/web/api"
	"github.com/GeoNet/app/web/api/apidoc"
	"net/http"
	"strings"
)

var docs = apidoc.Docs{
	Production: config.WebServer.Production,
	APIHost:    config.WebServer.CNAME,
	Title:      `GeoNet API`,
	Description: `<p>The data provided here is used for the GeoNet web site and other similar services. 
			If you are looking for data for research or other purposes then please check the 
			<a href="http://info.geonet.org.nz/x/DYAO">full range of data</a> available from GeoNet. </p>`,
	// RepoURL: `https://github.com/GeoNet/geonet-rest`,
	StrictVersioning: true,
}

func init() {
	docs.AddEndpoint("quake", &quakeDoc)
	docs.AddEndpoint("region", &regionDoc)
	docs.AddEndpoint("felt", &feltDoc)
	docs.AddEndpoint("news", &newsDoc)
}

var exHost = "http://localhost:" + config.WebServer.Port

func router(w http.ResponseWriter, r *http.Request) {
	switch {
	// application/vnd.geo+json;version=1
	case r.Header.Get("Accept") == web.V1GeoJSON:
		w.Header().Set("Content-Type", web.V1GeoJSON)
		switch {
		// /quake?regionID=newzealand&intensity=unnoticeable&number=30&quality=best,caution,good
		case r.URL.Path == "/quake" && r.URL.Query().Get("intensity") != "":
			q := &quakesQuery{}
			api.Serve(q, w, r)
		// /quake?regionID=newzealand&regionIntensity=unnoticeable&number=30&quality=best,caution,good
		case r.URL.Path == "/quake" && r.URL.Query().Get("regionIntensity") != "":
			q := &quakesRegionQuery{}
			api.Serve(q, w, r)
		// /quake/2013p407387
		case strings.HasPrefix(r.URL.Path, "/quake/"):
			q := &quakeQuery{}
			api.Serve(q, w, r)
		// /felt/report?publicID=2013p407387
		case r.URL.Path == "/felt/report":
			q := &feltQuery{}
			api.Serve(q, w, r)
		// /region/wellington
		case strings.HasPrefix(r.URL.Path, "/region/"):
			q := &regionQuery{}
			api.Serve(q, w, r)
		// /region?type=quake
		case r.URL.Path == "/region":
			q := &regionsQuery{}
			api.Serve(q, w, r)
		default:
			web.BadRequest(w, r, "service not found.")
		}
	// application/json;version=1
	case r.Header.Get("Accept") == web.V1JSON:
		w.Header().Set("Content-Type", web.V1JSON)
		switch {
		// /news/geonet
		case r.URL.Path == "/news/geonet":
			q := &newsQuery{}
			api.Serve(q, w, r)
		default:
			web.BadRequest(w, r, "service not found.")
		}
	// api-doc queries.
	case strings.HasPrefix(r.URL.Path, apidoc.Path):
		docs.Serve(w, r)
	case r.URL.Path == "/soh":
		soh(w, r)
	default:
		web.NotAcceptable(w, r, "Can't find a route for Accept header. Please refer to /api-docs")
	}
}
