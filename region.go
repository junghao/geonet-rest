package main

import (
	"net/http"
)

func (q *regionsQuery) validate(w http.ResponseWriter, r *http.Request) bool {
	return true
}

// /region?type=quake
// GeoJSON for classes of regions (quakes only atm).
type regionsQuery struct{}

// regions change very infrequently so they are loaded on startup and cached see -lookups.go
func (q *regionsQuery) handle(w http.ResponseWriter, r *http.Request) {
	ok(w, r, qrV1GeoJSON)
}

// /region/wellington
type regionQuery struct {
	regionID string
}

func (q *regionQuery) validate(w http.ResponseWriter, r *http.Request) bool {
	if _, ok := allRegion[q.regionID]; !ok {
		badRequest(w, r, "Invalid regionID: "+q.regionID)
		return false
	}

	return true
}

func (q *regionQuery) handle(w http.ResponseWriter, r *http.Request) {
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
