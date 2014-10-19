package geojsonV1

import (
	"database/sql"
	"github.com/GeoNet/geonet-rest/web"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"net/http"
)

// quakeRegions serves GeoJSON for all quake regions.
func quakeRegions(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var d string

	err := db.QueryRow(`SELECT row_to_json(fc)
                         FROM ( SELECT 'FeatureCollection' as type, array_to_json(array_agg(f)) as features
                         FROM (SELECT 'Feature' as type,
                         ST_AsGeoJSON(q.geom)::json as geometry,
                         row_to_json((SELECT l FROM 
                         	(
                         		SELECT 
                         		regionname as "regionID",
                         		title, 
                         		groupname as group
                           ) as l
                         )) as properties FROM qrt.region as q where groupname in ('region', 'north', 'south')) as f ) as fc`).Scan(&d)
	if err != nil {
		web.Fail(w, r, err)
		return
	}

	web.Win(w, r, []byte(d))
}

// region serves GeoJSON for a region.
// Returns 404 if the region is does not exist.
func region(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	p := mux.Vars(r)

	// Check that the region exists in the DB.
	rows, err := db.Query("select * FROM qrt.region where regionname = $1", p["regionID"])
	if err != nil {
		web.Fail(w, r, err)
		return
	}
	defer rows.Close()

	if !rows.Next() {
		web.Nope(w, r, "invalid region: "+p["regionID"])
		return
	}

	var d string

	err = db.QueryRow(`SELECT row_to_json(fc)
                         FROM ( SELECT 'FeatureCollection' as type, array_to_json(array_agg(f)) as features
                         FROM (SELECT 'Feature' as type,
                         ST_AsGeoJSON(q.geom)::json as geometry,
                         row_to_json((SELECT l FROM 
                         	(
                         		SELECT 
                         		regionname as "regionID",
                         		title, 
                         		groupname as group
                           ) as l
                         )) as properties FROM qrt.region as q where regionname = $1 ) as f ) as fc`, p["regionID"]).Scan(&d)
	if err != nil {
		web.Fail(w, r, err)
		return
	}

	web.Win(w, r, []byte(d))
}
