package geojsonV1

import (
	"database/sql"
	"github.com/GeoNet/geonet-rest/util"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"net/http"
)

// quake serves GeoJSON for a specific publicID.
func quake(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	p := mux.Vars(r)

	var d string

	err := db.QueryRow(
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

	util.PrettyJSON(w, d)
}

// quakes serves GeoJSON of quakes above an intensity in a region.
// Returns 404 if the regionID is not for a valid quake region.
func quakes(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	p := mux.Vars(r)

	// Check that the quake region exists in the DB.
	rows, err := db.Query("select * FROM qrt.region where regionname = $1 and groupname in ('region', 'north', 'south')", p["regionID"])
	if err != nil {
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
                                qrt.mmi_to_intensity(maxmmi) as intensity,
                                qrt.mmi_to_intensity(mmi_`+p["regionID"]+`) as "regionIntensity",
                                qrt.quake_quality(status, usedphasecount, magnitudestationcount) as quality,
                                to_char(updatetime, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"') as "modificationTime"
                           ) as l
                         )) as properties FROM qrt.quakeinternal as q where mmi_`+p["regionID"]+` >= qrt.intensity_to_mmi($1) limit $2 ) as f ) as fc`, p["intensity"], p["number"]).Scan(&d)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	util.PrettyJSON(w, d)
}
