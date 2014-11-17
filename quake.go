package main

import (
	"net/http"
	"strings"
	"time"
)

const (
	quakeLen = 7 //  len("/quake/")
)

// quakeV1 serves version 1 GeoJSON for a specific publicID.
func quakeV1(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", v1GeoJSON)

	q := &quakeQuery{
		publicID:   r.URL.Path[quakeLen:],
		queryCount: 0,
	}

	if ok := q.validate(w, r); !ok {
		return
	}

	var d string

	start := time.Now()
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
                         )) as properties FROM qrt.quake_materialized as q where publicid = $1 ) As f )  as fc`, q.publicID).Scan(&d)
	if err != nil {
		serviceUnavailable(w, r, err)
		return
	}
	dbTime.track(start, "quakeV1 db")

	ok(w, r, []byte(d))
}

// quakesRegionV1 serves GeoJSON of quakes thay may have been felt above an intensity in a region.
// The quakes could be outside the region.
func quakesRegionV1(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", v1GeoJSON)

	q := &quakesQuery{
		number:     r.URL.Query().Get("number"),
		regionID:   r.URL.Query().Get("regionID"),
		intensity:  r.URL.Query().Get("regionIntensity"),
		quality:    strings.Split(r.URL.Query().Get("quality"), ","),
		queryCount: 4,
	}

	if ok := q.validate(w, r); !ok {
		return
	}

	var d string

	start := time.Now()
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
                                intensity_`+q.regionID+` as "regionIntensity",
                                quality,
                                to_char(updatetime, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"') as "modificationTime"
                           ) as l
                         )) as properties FROM qrt.quakeinternal_v2 as q where mmi_`+q.regionID+` >= qrt.intensity_to_mmi($1)
                         AND quality in ('`+strings.Join(q.quality, `','`)+`') limit $2 ) as f ) as fc`, q.intensity, q.number).Scan(&d)
	if err != nil {
		serviceUnavailable(w, r, err)
		return
	}
	dbTime.track(start, "quakeRegionV1 db")

	ok(w, r, []byte(d))
}

// quakesV1 serves GeoJSON of quakes that occured in a region filtered by intensity.
func quakesV1(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", v1GeoJSON)

	q := &quakesQuery{
		number:     r.URL.Query().Get("number"),
		regionID:   r.URL.Query().Get("regionID"),
		intensity:  r.URL.Query().Get("intensity"),
		quality:    strings.Split(r.URL.Query().Get("quality"), ","),
		queryCount: 4,
	}

	if ok := q.validate(w, r); !ok {
		return
	}

	var d string

	start := time.Now()
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
                                intensity_`+q.regionID+` as "regionIntensity",
                                quality,
                                to_char(updatetime, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"') as "modificationTime"
                           ) as l
                         )) as properties FROM qrt.quakeinternal_v2 as q where maxmmi >= qrt.intensity_to_mmi($1)
                         AND quality in ('`+strings.Join(q.quality, `','`)+`')  AND ST_Contains((select geom from qrt.region where regionname = $3), ST_Shift_Longitude(origin_geom)) limit $2 ) as f ) as fc`, q.intensity, q.number, q.regionID).Scan(&d)
	if err != nil {
		serviceUnavailable(w, r, err)
		return
	}
	dbTime.track(start, "quakesV1 db")

	ok(w, r, []byte(d))
}
