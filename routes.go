package main

import (
	"github.com/GeoNet/web"
	"github.com/GeoNet/web/api/apidoc"
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
	// docs.AddEndpoint("impact", &impactDoc)
}

var exHost = "http://localhost:" + config.WebServer.Port

func router(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Header.Get("Accept") == web.V1GeoJSON:
		w.Header().Set("Content-Type", web.V1GeoJSON)
		switch {
		case strings.HasPrefix(r.URL.Path, "/quake"):
			switch {
			case r.URL.Query().Get("intensity") != "":
				quakes(w, r)
			case r.URL.Query().Get("regionIntensity") != "":
				quakesRegion(w, r)
			case strings.HasPrefix(r.URL.Path, "/quake/"):
				quake(w, r)
			default:
				web.BadRequest(w, r, "service not found.")
			}
		case r.URL.Path == "/intensity":
			switch {
			case r.URL.Query().Get("type") == "measured":
				intensityMeasuredLatest(w, r)
			case r.URL.Query().Get("type") == "reported" && r.URL.Query().Get("publicID") == "":
				intensityReportedLatest(w, r)
			case r.URL.Query().Get("type") == "reported" && r.URL.Query().Get("publicID") != "":
				intensityReported(w, r)
			default:
				web.BadRequest(w, r, "service not found.")
			}
		case r.URL.Path == "/felt/report":
			felt(w, r)
		case strings.HasPrefix(r.URL.Path, "/region/"):
			region(w, r)
		case r.URL.Path == "/region":
			regions(w, r)
		default:
			web.BadRequest(w, r, "service not found.")
		}
	case r.Header.Get("Accept") == web.V1JSON:

		switch {
		case r.URL.Path == "/news/geonet":
			news(w, r)
		default:
			web.BadRequest(w, r, "service not found.")
		}
	// api-doc queries.
	case strings.HasPrefix(r.URL.Path, apidoc.Path):
		docs.Serve(w, r)
	case r.URL.Path == "/soh":
		soh(w, r)
	case r.URL.Path == "/soh/impact":
		impactSOH(w, r)
	default:
		web.NotAcceptable(w, r, "Can't find a route for Accept header. Please refer to /api-docs")
	}
}
