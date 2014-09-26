// geojsonV1 provides version=1 vnd.geo+json services.
package geojsonV1

import (
	"database/sql"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"net/http"
)

const Accept = "application/vnd.geo+json; version=1;"

// makeHandler makes a handler that uses a DB connection.  Also sets http response headers.
func makeHandler(fn func(http.ResponseWriter, *http.Request, *sql.DB), db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=10")
		w.Header().Set("Content-Type", Accept)
		fn(w, r, db)
	}
}

// Routes adds handlers to the mux.Router.
func Routes(r *mux.Router, db *sql.DB) {
	// /region/wellington
	r.HandleFunc("/region/{regionID:[a-z]+}", makeHandler(region, db))

	// /region?type=quake
	r.HandleFunc("/region", makeHandler(quakeRegions, db)).Queries("type", "quake")

	// /quake/2013p407387
	r.HandleFunc("/quake/{publicID:[0-9]+[a-z0-9]+}", makeHandler(quake, db))

	// /quake?regionID=newzealand&intensity=weak&number=30
	r.HandleFunc("/quake", makeHandler(quakes, db)).
		Queries("regionID", "{regionID:[a-z]+}",
		"intensity", "{intensity:unnoticeable|weak|light|moderate|strong|severe}",
		"number", "{number:30|100|500|1000|1500}")
}
