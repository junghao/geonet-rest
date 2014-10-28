package main

import (
	"net/http"
)

// regions serves GeoJSON for classes of regions (quakes only atm).
func regionsV1(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", v1GeoJSON)

	if r.URL.Query().Get("type") != "quake" {
		nope(w, r, "Invalid type: "+r.URL.Query().Get("type"))
	}

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
		fail(w, r, err)
		return
	}

	win(w, r, []byte(d))
}

// region serves GeoJSON for a region.
// Returns 404 if the region is does not exist.
func regionV1(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", v1GeoJSON)

	regionID := r.URL.Path[len("/region/"):]

	// check the regionID query is valid.
	if _, ok := allRegion[regionID]; !ok {
		nope(w, r, "Invalid regionID: "+regionID)
	}

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
	if err != nil {
		fail(w, r, err)
		return
	}

	win(w, r, []byte(d))
}
