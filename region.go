package main

import (
	"html/template"
	"net/http"
)

// /region?type=quake

var regionsQueryD = &doc{
	Title:       "Quake Region",
	Description: "retrieve quake regions.",
	Example:     "/region?type=quake",
	URI:         "/region?type=quake",
	Params: map[string]template.HTML{
		"type": `the region type.  The only allowable value is <code>quake</code>.`,
	},
	Props: map[string]template.HTML{
		`regionID`: `a unique indentifier for the region.`,
		`title`:    `the region title.`,
		`group`:    `the region group.`,
	},
	Result: `{"type":"FeatureCollection","features":[{"type":"Feature","geometry":{"type":"Polygon","coordinates":[[[190,-20],[182,-37],[184,-44],[167,-49],[160,-54],[164,-47],[165,-44],[170,-35],[174,-32],[190,-20]]]},"properties":{"regionID":"newzealand","title":"New Zealand","group":"region"}},
	{"type":"Feature","geometry":{"type":"Polygon","coordinates":[[[173.251,-38.138],[175.583,-38.045],[176.474,-36.379],[174.285,-34.026],[171.857,-34.135],[173.251,-38.138]]]},"properties":{"regionID":"aucklandnorthland","title":"Auckland and Northland","group":"north"}},
	{"type":"Feature","geometry":{"type":"Polygon","coordinates":[[[176.931,-38.688],[175.722,-39.809],[177.56,-40.638],[178.561,-39.274],[176.931,-38.688]]]},"properties":{"regionID":"hawkesbay","title":"Hawke's Bay","group":"north"}},
	{"type":"Feature","geometry":{"type":"Polygon","coordinates":[[[172.004,-39.632],[174.156,-40.456],[175.028,-39.526],[175.583,-38.045],[173.251,-38.138],[172.004,-39.632]]]},"properties":{"regionID":"taranaki","title":"Taranaki","group":"north"}}]}`,
}

func (q *regionsQuery) doc() *doc {
	return regionsQueryD
}

type regionsQuery struct{}

func (q *regionsQuery) validate(w http.ResponseWriter, r *http.Request) bool {
	return true
}

// regions change very infrequently so they are loaded on startup and cached see -lookups.go
func (q *regionsQuery) handle(w http.ResponseWriter, r *http.Request) {
	ok(w, r, qrV1GeoJSON)
}

// /region/wellington

var regionQueryD = &doc{
	Title:       "Region Information",
	Description: "Look up region information",
	Example:     "/region/wellington",
	URI:         "/region/(regionID)",
	Params: map[string]template.HTML{
		"regionID": `A region ID e.g., <code>wellington</code>.`,
	},
	Props: map[string]template.HTML{
		`regionID`: `a unique indentifier for the region.`,
		`title`:    `the region title.`,
		`group`:    `the region group.`,
	},
	Result: `{"type":"FeatureCollection","features":[{"type":"Feature","geometry":
	{"type":"Polygon","coordinates":[[[172.951,-41.767],[175.748,-42.908],
	[177.56,-40.638],[175.028,-39.526],[174.109,-40.462],[172.951,-41.767]]]},
	"properties":{"regionID":"wellington","title":"Wellington and Marlborough","group":"north"}}]}`,
}

func (q *regionQuery) doc() *doc {
	return regionQueryD
}

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
