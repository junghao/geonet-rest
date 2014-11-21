package main

import (
	"database/sql"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var intensityRe = regexp.MustCompile(`^(unnoticeable|weak|light|moderate|strong|severe)$`)
var numberRe = regexp.MustCompile(`^(3|30|100|500|1000|1500)$`)
var qualityRe = regexp.MustCompile(`^(best|caution|deleted|good)$`)

// /quake/2013p407387
type quakeQuery struct {
	publicID string
}

func (q *quakeQuery) validate(w http.ResponseWriter, r *http.Request) bool {
	var d string

	// Check that the publicid exists in the DB.  This is needed as the handle method will return empty
	// JSON for an invalid publicID.
	err := db.QueryRow("select publicid FROM qrt.quake_materialized where publicid = $1", q.publicID).Scan(&d)
	if err == sql.ErrNoRows {
		notFound(w, r, "invalid publicID: "+q.publicID)
		return false
	}
	if err != nil {
		serviceUnavailable(w, r, err)
		return false
	}
	return true
}

func (q *quakeQuery) handle(w http.ResponseWriter, r *http.Request) {
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

// /quake?regionID=newzealand&regionIntensity=unnoticeable&number=30&quality=best,caution,good
// Quakes thay may have been felt above an intensity in a region.
// The quakes could be outside the region.
type quakesRegionQuery struct {
	regionID, regionIntensity, number string
	quality                           []string
}

func (q *quakesRegionQuery) validate(w http.ResponseWriter, r *http.Request) bool {

	if !numberRe.MatchString(q.number) {
		badRequest(w, r, "Invalid number: "+q.number)
		return false
	}

	if !intensityRe.MatchString(q.regionIntensity) {
		badRequest(w, r, "Invalid region intensity: "+q.regionIntensity)
		return false
	}

	if _, ok := quakeRegion[q.regionID]; !ok {
		badRequest(w, r, "Invalid regionID: "+q.regionID)
		return false
	}

	for _, q := range q.quality {
		if !qualityRe.MatchString(q) {
			badRequest(w, r, "Invalid quality: "+q)
			return false
		}
	}

	return true
}

func (q *quakesRegionQuery) handle(w http.ResponseWriter, r *http.Request) {
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
                         AND quality in ('`+strings.Join(q.quality, `','`)+`') limit $2 ) as f ) as fc`, q.regionIntensity, q.number).Scan(&d)
	if err != nil {
		serviceUnavailable(w, r, err)
		return
	}
	dbTime.track(start, "quakeRegionV1 db")

	ok(w, r, []byte(d))
}

// /quake?regionID=newzealand&intensity=unnoticeable&number=30&quality=best,caution,good
// Quakes that occured in a region filtered by intensity.
type quakesQuery struct {
	regionID, intensity, number string
	quality                     []string
}

func (q *quakesQuery) validate(w http.ResponseWriter, r *http.Request) bool {

	if !numberRe.MatchString(q.number) {
		badRequest(w, r, "Invalid number: "+q.number)
		return false
	}

	if !intensityRe.MatchString(q.intensity) {
		badRequest(w, r, "Invalid intensity: "+q.intensity)
		return false
	}

	if _, ok := quakeRegion[q.regionID]; !ok {
		badRequest(w, r, "Invalid regionID: "+q.regionID)
		return false
	}

	for _, q := range q.quality {
		if !qualityRe.MatchString(q) {
			badRequest(w, r, "Invalid quality: "+q)
			return false
		}
	}

	return true
}

func (q *quakesQuery) handle(w http.ResponseWriter, r *http.Request) {
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
