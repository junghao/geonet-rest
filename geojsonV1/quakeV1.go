package geojsonV1

import (
	"database/sql"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strings"
)

var quality map[string]int

func init() {
	quality = make(map[string]int)
	quality = map[string]int{
		"best":    1,
		"caution": 1,
		"deleted": 1,
		"good":    1,
	}
}

// quake serves GeoJSON for a specific publicID.
func quake(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	p := mux.Vars(r)

	var d string

	// Check that the publicid exists in the DB.
	rows, err := db.Query("select * FROM qrt.quake_materialized where publicid = $1", p["publicID"])
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	if !rows.Next() {
		http.Error(w, "invalid publicID "+p["publicID"], 404)
		return
	}

	err = db.QueryRow(
		`SELECT row_to_json(fc)
                         FROM ( SELECT 'FeatureCollection' as type, array_to_json(array_agg(f)) as features
                         FROM (SELECT 'Feature' as type,
                         ST_AsGeoJSON(q.origin_geom)::json as geometry,
                         row_to_json((SELECT l FROM 
                         	(
                         		SELECT 
                         		publicid AS "publicID",
                                to_char(origintime, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"') as "time",
                                depth, 
                                magnitude, 
                                type, 
                                agency, 
                                locality,
                                qrt.mmi_to_intensity(maxmmi) as intensity,
                                qrt.mmi_to_intensity(mmi_newzealand) as "regionIntensity",
                                qrt.quake_quality(status, usedphasecount, magnitudestationcount) as quality,
                                to_char(updatetime, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"') as "modificationTime"
                           ) as l
                         )) as properties FROM qrt.quake_materialized as q where publicid = $1 ) As f )  as fc`, p["publicID"]).Scan(&d)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte(d))
}

// quakes serves GeoJSON of quakes above an intensity in a region.
// Returns 404 if the regionID is not for a valid quake region.
func quakes(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	p := mux.Vars(r)

	// check that the quality query is for valid options.
	qual := strings.Split(p["quality"], ",")
	for _, q := range qual {
		if _, ok := quality[q]; !ok {
			http.Error(w, "Invalid quality: "+q, 500)
		}
	}

	// Check that the quake region exists in the DB.
	rows, err := db.Query("select * FROM qrt.region where regionname = $1 and groupname in ('region', 'north', 'south')", p["regionID"])
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	if !rows.Next() {
		http.Error(w, "invalid region: "+p["regionID"], 404)
		return
	}

	var d string

	err = db.QueryRow(
		`SELECT row_to_json(fc)
                         FROM ( SELECT 'FeatureCollection' as type, array_to_json(array_agg(f)) as features
                         FROM (SELECT 'Feature' as type,
                         ST_AsGeoJSON(q.origin_geom)::json as geometry,
                         row_to_json((SELECT l FROM 
                         	(
                         		SELECT 
                         		publicid AS "publicID",
                                to_char(origintime, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"') as "time",
                                depth, 
                                magnitude, 
                                type, 
                                agency, 
                                locality,
                                intensity,
                                intensity_`+p["regionID"]+` as "regionIntensity",
                                quality,
                                to_char(updatetime, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"') as "modificationTime"
                           ) as l
                         )) as properties FROM qrt.quakeinternal_v2 as q where mmi_`+p["regionID"]+` >= qrt.intensity_to_mmi($1) 
                         AND quality in ('`+strings.Join(qual, `','`)+`') limit $2 ) as f ) as fc`, p["intensity"], p["number"]).Scan(&d)
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte(d))
}
