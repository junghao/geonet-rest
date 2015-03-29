package main

import (
	"database/sql"
	"github.com/GeoNet/web"
	"github.com/GeoNet/web/api/apidoc"
	"html/template"
	"net/http"
)

// These constants are the length of parts of the URI and are used for
// extracting query params embedded in the URI.
const (
	regionLen = 8 // len("/region/")
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

type regionsQuery struct {
	regionType string
}

func (q *regionsQuery) Validate(w http.ResponseWriter, r *http.Request) bool {
	switch {
	case len(r.URL.Query()) != 1:
		web.BadRequest(w, r, "incorrect number of query parameters.")
		return false
	case !web.ParamsExist(w, r, "type"):
		return false
	}

	q.regionType = r.URL.Query().Get("type")

	if q.regionType != "quake" {
		web.BadRequest(w, r, "type must be quake.")
		return false
	}

	return true
}

// just quake regions at the moment.
func (q *regionsQuery) Handle(w http.ResponseWriter, r *http.Request) {
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
		web.ServiceUnavailable(w, r, err)
		return
	}

	w.Header().Set("Surrogate-Control", web.MaxAge86400)
	b := []byte(d)
	web.Ok(w, r, &b)
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
	if len(r.URL.Query()) != 0 {
		web.BadRequest(w, r, "incorrect number of query parameters.")
		return false
	}

	q.regionID = r.URL.Path[regionLen:]

	var d string

	err := db.QueryRow("select regionname FROM qrt.region where regionname = $1", q.regionID).Scan(&d)
	if err == sql.ErrNoRows {
		web.BadRequest(w, r, "invalid regionID: "+q.regionID)
		return false
	}
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return false
	}
	return true
}

func (q *regionQuery) Handle(w http.ResponseWriter, r *http.Request) {
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
                         )) as properties FROM qrt.region as q where regionname = $1 ) as f ) as fc`, q.regionID).Scan(&d)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}

	w.Header().Set("Surrogate-Control", web.MaxAge86400)
	b := []byte(d)
	web.Ok(w, r, &b)
}
