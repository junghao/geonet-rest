package main

import (
	"database/sql"
	"net/http"
	"strings"
)

// quakeV1 serves version 1 GeoJSON for a specific publicID.
func quakeV1(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", v1GeoJSON)

	publicID := r.URL.Path[len("/quake/"):]
	var d string

	// Check that the publicid exists in the DB.  This is needed as the geoJSON query will return empty
	// JSON for an invalid publicID.
	err := db.QueryRow("select publicid FROM qrt.quake_materialized where publicid = $1", publicID).Scan(&d)
	if err == sql.ErrNoRows {
		nope(w, r, "invalid publicID: "+publicID)
		return
	}
	if err != nil {
		fail(w, r, err)
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
                         )) as properties FROM qrt.quake_materialized as q where publicid = $1 ) As f )  as fc`, publicID).Scan(&d)
	if err != nil {
		fail(w, r, err)
		return
	}

	win(w, r, []byte(d))
}

// quakesV1 serves GeoJSON of quakes above an intensity in a region.
// Returns 404 if the regionID is not for a valid quake region.
func quakesV1(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", v1GeoJSON)

	//  check the number query is for valid options
	if _, ok := number[r.URL.Query().Get("number")]; !ok {
		nope(w, r, "Invalid number: "+r.URL.Query().Get("number"))
	}

	// check the intensity query is for valid options.
	if _, ok := intensity[r.URL.Query().Get("intensity")]; !ok {
		nope(w, r, "Invalid intensity: "+r.URL.Query().Get("intensity"))
	}

	// check the regionID query is for valid.
	if _, ok := quakeRegion[r.URL.Query().Get("regionID")]; !ok {
		nope(w, r, "Invalid regionID: "+r.URL.Query().Get("regionID"))
	}

	// check that the quality query is for valid options.
	qual := strings.Split(r.URL.Query().Get("quality"), ",")
	for _, q := range qual {
		if _, ok := quality[q]; !ok {
			nope(w, r, "Invalid quality: "+q)
		}
	}

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
                                intensity,
                                intensity_`+r.URL.Query().Get("regionID")+` as "regionIntensity",
                                quality,
                                to_char(updatetime, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"') as "modificationTime"
                           ) as l
                         )) as properties FROM qrt.quakeinternal_v2 as q where mmi_`+r.URL.Query().Get("regionID")+` >= qrt.intensity_to_mmi($1)
                         AND quality in ('`+strings.Join(qual, `','`)+`') limit $2 ) as f ) as fc`, r.URL.Query().Get("intensity"), r.URL.Query().Get("number")).Scan(&d)
	if err != nil {
		fail(w, r, err)
		return
	}

	win(w, r, []byte(d))
}
