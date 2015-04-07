package main

import (
	"database/sql"
	"github.com/GeoNet/web"
	"github.com/GeoNet/web/api/apidoc"
	"html/template"
	"net/http"
	"regexp"
	"time"
)

var impactDoc = apidoc.Endpoint{Title: "Impact",
	Description: `Look up impact information`,
	Queries: []*apidoc.Query{
		new(intensityReportedQuery).Doc(),
		new(intensityReportedLatestQuery).Doc(),
		new(intensityMeasuredLatestQuery).Doc(),
	},
}

var zoomRe = regexp.MustCompile(`^(5|6)$`)

var intensityMeasuredLatestQueryD = &apidoc.Query{
	Accept:      web.V1GeoJSON,
	Title:       "Measured Intensity - Latest",
	Description: "Retrieve measured intensity information in the last sixty minutes.",
	Example:     "/intensity?type=measured",
	ExampleHost: exHost,
	URI:         "/intensity?type",
	Params: map[string]template.HTML{
		"type": `<code>measured</code> is the only allowed value.`,
	},
	Props: map[string]template.HTML{
		"max_mmi": `the maximum <a href="http://info.geonet.org.nz/x/w4IO">Modified Mercalli Intensity (MMI)</a> measured at the point in the last sixty minutes.`,
	},
}

type intensityMeasuredLatestQuery struct {
}

func (q *intensityMeasuredLatestQuery) Validate(w http.ResponseWriter, r *http.Request) bool {
	switch {
	case len(r.URL.Query()) != 1:
		web.BadRequest(w, r, "incorrect number of query parameters.")
		return false
	case r.URL.Query().Get("type") != "measured":
		web.BadRequest(w, r, "type must be measured.")
		return false
	}

	return true
}

func (q *intensityMeasuredLatestQuery) Handle(w http.ResponseWriter, r *http.Request) {
	var d string

	err := db.QueryRow(
		`SELECT row_to_json(fc)
				FROM ( SELECT 'FeatureCollection' as type, COALESCE(array_to_json(array_agg(f)), '[]') as features
					FROM (SELECT 'Feature' as type,
						ST_AsGeoJSON(s.location)::json as geometry,
						row_to_json(( select l from 
							( 
								select mmi
								) as l )) 
			as properties from (select location, mmi 
				FROM impact.intensity_measured) as s 
			) As f )  as fc`).Scan(&d)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}

	b := []byte(d)
	web.Ok(w, r, &b)
}

func (q *intensityMeasuredLatestQuery) Doc() *apidoc.Query {
	return intensityMeasuredLatestQueryD
}

// latest reported intensity

var intensityReportedLatestQueryD = &apidoc.Query{
	Accept:      web.V1GeoJSON,
	Title:       "Reported Intensity - Latest",
	Description: "Retrieve reported intensity information in the last sixty minutes.",
	Example:     "/intensity?type=reported&zoom=5",
	ExampleHost: exHost,
	URI:         "/intensity?type=reported&zoom=(int)",
	Params: map[string]template.HTML{
		"zoom": `The zoom level to aggregate values at.  This controls the size of the area that values are aggregated at.  The point returned
				will be the center of each area.  Allowed values are one of <code>5, 6</code>.`,
	},
	Props: map[string]template.HTML{
		"max_mmi": `the maximum <a href="http://info.geonet.org.nz/x/w4IO">Modified Mercalli Intensity (MMI)</a> 
					in the area around the point in the last sixty minutes.`,
		"min_mmi": `the minimum <a href="http://info.geonet.org.nz/x/w4IO">Modified Mercalli Intensity (MMI)</a> 
					in the area of around the point in the last sixty minutes.`,
		"count": `the count of <a href="http://info.geonet.org.nz/x/w4IO">Modified Mercalli Intensity (MMI)</a> 
					values reported in the area of around the point in the last sixty minutes.`,
	},
}

type intensityReportedLatestQuery struct {
	zoom string
}

func (q *intensityReportedLatestQuery) Validate(w http.ResponseWriter, r *http.Request) bool {
	switch {
	case len(r.URL.Query()) != 2:
		web.BadRequest(w, r, "incorrect number of query parameters.")
		return false
	case !web.ParamsExist(w, r, "type", "zoom"):
		return false
	case !zoomRe.MatchString(r.URL.Query().Get("zoom")):
		web.BadRequest(w, r, "Invalid zoom")
		return false
	case r.URL.Query().Get("type") != "reported":
		web.BadRequest(w, r, "Invalid type")
		return false
	}

	q.zoom = r.URL.Query().Get("zoom")

	return true
}

func (q *intensityReportedLatestQuery) Handle(w http.ResponseWriter, r *http.Request) {
	var d string

	err := db.QueryRow(
		`SELECT row_to_json(fc)
						FROM ( SELECT 'FeatureCollection' as type, COALESCE(array_to_json(array_agg(f)), '[]') as features
							FROM (SELECT 'Feature' as type,
								ST_AsGeoJSON(s.location)::json as geometry,
								row_to_json(( select l from 
									( 
										select max_mmi,
										min_mmi,
										count
										) as l )) 
					as properties from (select st_pointfromgeohash(geohash` + q.zoom + `) as location, 
						min(mmi) as min_mmi, 
						max(mmi) as max_mmi, 
						count(mmi) as count 
						FROM impact.intensity_reported  
						WHERE time >= (now() - interval '60 minutes')
						group by (geohash` + q.zoom + `)) as s
					) As f )  as fc`).Scan(&d)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}

	b := []byte(d)
	web.Ok(w, r, &b)
}

func (q *intensityReportedLatestQuery) Doc() *apidoc.Query {
	return intensityReportedLatestQueryD
}

// reported intensity

var intensityReportedQueryD = &apidoc.Query{
	Accept:      web.V1GeoJSON,
	Title:       "Reported Intensity",
	Description: "Retrieve reported intensity information in a 15 minute time window after an event.",
	Example:     "/intensity?type=reported&zoom=5&publicID=2013p407387",
	ExampleHost: exHost,
	URI:         "/intensity?type=reported&zoom=(int)&publicID=(publicID)",
	Params: map[string]template.HTML{
		"zoom": `The zoom level to aggregate values at.  This controls the size of the area that values are aggregated at.  The point returned
						will be the center of each area.  Allowed values are one of <code>5, 6</code>.`,
		"publicID": `a valid quake ID e.g., <code>2014p715167</code>`,
	},
	Props: map[string]template.HTML{
		"max_mmi": `the maximum <a href="http://info.geonet.org.nz/x/w4IO">Modified Mercalli Intensity (MMI)</a> 
							in the area around the point in the last sixty minutes.`,
		"min_mmi": `the minimum <a href="http://info.geonet.org.nz/x/w4IO">Modified Mercalli Intensity (MMI)</a> 
							in the area of around the point in the last sixty minutes.`,
		"count": `the count of <a href="http://info.geonet.org.nz/x/w4IO">Modified Mercalli Intensity (MMI)</a> 
							values reported in the area of around the point in the last sixty minutes.`,
	},
}

type intensityReportedQuery struct {
	zoom       string
	originTime time.Time
}

func (q *intensityReportedQuery) Validate(w http.ResponseWriter, r *http.Request) bool {
	switch {
	case len(r.URL.Query()) != 3:
		web.BadRequest(w, r, "incorrect number of query parameters.")
		return false
	case !web.ParamsExist(w, r, "type", "zoom", "publicID"):
		return false
	case !zoomRe.MatchString(r.URL.Query().Get("zoom")):
		web.BadRequest(w, r, "Invalid zoom")
		return false
	case !publicIDRe.MatchString(r.URL.Query().Get("publicID")):
		web.BadRequest(w, r, "Invalid publicID")
		return false
	case r.URL.Query().Get("type") != "reported":
		web.BadRequest(w, r, "Invalid type")
		return false
	}

	q.zoom = r.URL.Query().Get("zoom")

	// Check that the publicid exists in the DB.
	// If it does we keep the origintime - we need it later on.
	err := db.QueryRow("select origintime FROM qrt.quake_materialized where publicid = $1", r.URL.Query().Get("publicID")).Scan(&q.originTime)
	if err == sql.ErrNoRows {
		web.NotFound(w, r, "invalid publicID: "+r.URL.Query().Get("publicID"))
		return false
	}
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return false
	}

	return true
}

func (q *intensityReportedQuery) Handle(w http.ResponseWriter, r *http.Request) {
	query := `SELECT row_to_json(fc)
				FROM ( SELECT 'FeatureCollection' as type, COALESCE(array_to_json(array_agg(f)), '[]') as features
					FROM (SELECT 'Feature' as type,
						ST_AsGeoJSON(s.location)::json as geometry,
						row_to_json(( select l from 
							( 
							select max_mmi,
							min_mmi,
							count
							) as l )) 
							as properties from (select st_pointfromgeohash(geohash` + q.zoom + `) as location, 
							min(mmi) as min_mmi, 
							max(mmi) as max_mmi, 
							count(mmi) as count 
							FROM impact.intensity_reported 
							WHERE time >= $1
							AND time <= $2
							group by (geohash` + q.zoom + `)) as s
			) As f )  as fc`

	var d string

	err := db.QueryRow(query, q.originTime.Add(time.Duration(-1*time.Minute)), q.originTime.Add(time.Duration(15*time.Minute))).Scan(&d)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}

	b := []byte(d)
	web.Ok(w, r, &b)
}

func (q *intensityReportedQuery) Doc() *apidoc.Query {
	return intensityReportedQueryD
}
