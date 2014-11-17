package main

import (
	"net/http"
)

const (
	regionLen = 8 // len("/region/")
)

// regions serves GeoJSON for classes of regions (quakes only atm).
// The regions change very infrequently so they are loaded on startup and cached see -lookups.go
func regionsV1(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", v1GeoJSON)

	// check we got the correct number of query params.  This rules out cache busters
	if len(r.URL.Query()) != 1 {
		badRequest(w, r, "detected extra stuff in the URL.")
		return
	}

	if r.URL.Query().Get("type") != "quake" {
		badRequest(w, r, "Invalid type: "+r.URL.Query().Get("type"))
		return
	}

	ok(w, r, qrV1GeoJSON)
}

// region serves GeoJSON for a region.
// The regions change very infrequently so they are loaded on startup and cached see -lookups.go
func regionV1(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", v1GeoJSON)

	q := &regionQuery{
		regionID:   r.URL.Path[regionLen:],
		queryCount: 0,
	}

	if ok := q.validate(w, r); !ok {
		return
	}

	ok(w, r, allRegion[q.regionID])
}

// quakeRegionsV1GJ queries the DB for GeoJSON for the quake regions.
func quakeRegionsV1GJ() ([]byte, error) {
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

	return []byte(d), err
}

// regionV1GJ queries the DB for GeoJSON from the regionID.
func regionV1GJ(regionID string) ([]byte, error) {
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
                         )) as properties FROM qrt.region as q where regionname = $1 ) as f ) as fc`, regionID).Scan(&d)

	return []byte(d), err
}
