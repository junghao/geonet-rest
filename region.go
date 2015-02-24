package main

import (
	"github.com/GeoNet/app/web"
	"github.com/GeoNet/app/web/api/apidoc"
	"html/template"
	"net/http"
)

var regionDoc = apidoc.Endpoint{
	Title:       "Region",
	Description: `Look up region information.`,
	Queries: []*apidoc.Query{
		new(regionQuery).Doc(),
		new(regionsQuery).Doc(),
	},
}

var regionsQueryD = &apidoc.Query{
	Accept:      web.V1GeoJSON,
	Title:       "Regions",
	Description: "Retrieve regions.",
	Example:     "/region?type=quake",
	ExampleHost: exHost,
	URI:         "/region?type=(type)",
	Params: map[string]template.HTML{
		"type": `the region type.  The only allowable value is <code>quake</code>.`,
	},
	Props: map[string]template.HTML{
		`regionID`: `a unique indentifier for the region.`,
		`title`:    `the region title.`,
		`group`:    `the region group.`,
	},
}

func (q *regionsQuery) Doc() *apidoc.Query {
	return regionsQueryD
}

type regionsQuery struct{}

func (q *regionsQuery) Validate(w http.ResponseWriter, r *http.Request) bool {
	return true
}

// regions change very infrequently so they are loaded on startup and cached see -lookups.go
func (q *regionsQuery) Handle(w http.ResponseWriter, r *http.Request) {
	web.Ok(w, r, &qrV1GeoJSON)
}

// /region/wellington

var regionQueryD = &apidoc.Query{
	Accept:      web.V1GeoJSON,
	Title:       "Region",
	Description: "Retrieve a single region.",
	Example:     "/region/wellington",
	ExampleHost: exHost,
	URI:         "/region/(regionID)",
	Params: map[string]template.HTML{
		"regionID": `A region ID e.g., <code>wellington</code>.`,
	},
	Props: map[string]template.HTML{
		`regionID`: `a unique indentifier for the region.`,
		`title`:    `the region title.`,
		`group`:    `the region group.`,
	},
}

func (q *regionQuery) Doc() *apidoc.Query {
	return regionQueryD
}

type regionQuery struct {
	regionID string
}

func (q *regionQuery) Validate(w http.ResponseWriter, r *http.Request) bool {
	q.regionID = r.URL.Path[regionLen:]

	if _, ok := allRegion[q.regionID]; !ok {
		web.BadRequest(w, r, "Invalid regionID: "+q.regionID)
		return false
	}

	return true
}

func (q *regionQuery) Handle(w http.ResponseWriter, r *http.Request) {
	b := allRegion[q.regionID]
	web.Ok(w, r, &b)
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
