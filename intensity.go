package main

import (
	"github.com/GeoNet/app/web"
	"github.com/GeoNet/app/web/api/apidoc"
	"html/template"
	"net/http"
	"regexp"
)

var impactDoc = apidoc.Endpoint{Title: "Impact",
	Description: `Look up impact information`,
	Queries: []*apidoc.Query{
		// new(intensityReportedQuery).Doc(),  // hide reported docs for now.  Not collecting any yet.
		// new(intensityReportedLatestQuery).Doc(),
		new(intensityMeasuredLatestQuery).Doc(),
	},
}

// Not the most accurate ISO8601 match in the world but it will do.
var startRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:00Z$`)
var bboxRe = regexp.MustCompile(`^-*\d+,-*\d+,-*\d+,-*\d+$`)
var windowRe = regexp.MustCompile(`^(15)$`)
var zoomRe = regexp.MustCompile(`^(5|6|7)$`)

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
	Example:     "/intensity?type=reported&bbox=165,-34,179,-47&zoom=5",
	ExampleHost: exHost,
	URI:         "/intensity?type=reported&bbox=(bbox)&zoom=(int)",
	Params: map[string]template.HTML{
		"bbox": `A spatial bounding box for the query describing the upper left and lower right longitude and latitude corners.  Integer values only.
				e.g., <code>165,-34,179,-47</code> would include mainland New Zealand.`,
		"zoom": `The zoom level to aggregate values at.  This controls the size of the area that values are aggregated at.  The point returned
				will be the center of each area.  Allowed values are one of <code>5, 6, 7</code>.`,
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
	bbox, zoom string
}

func (q *intensityReportedLatestQuery) Validate(w http.ResponseWriter, r *http.Request) bool {
	switch {
	case len(r.URL.Query()) != 3:
		web.BadRequest(w, r, "incorrect number of query parameters.")
		return false
	case !web.ParamsExist(w, r, "type", "bbox", "zoom"):
		return false
	case !bboxRe.MatchString(r.URL.Query().Get("bbox")):
		web.BadRequest(w, r, "Invalid bbox")
		return false
	case !zoomRe.MatchString(r.URL.Query().Get("zoom")):
		web.BadRequest(w, r, "Invalid zoom")
		return false
	case r.URL.Query().Get("type") != "reported":
		web.BadRequest(w, r, "Invalid type")
		return false
	}

	q.bbox = r.URL.Query().Get("bbox")
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
						AND location && ST_MakeEnvelope(` + q.bbox + `, 4326) group by (geohash` + q.zoom + `)) as s
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
	Description: "Retrieve reported intensity information in a time window.",
	Example:     "/intensity?type=reported&bbox=165,-34,179,-47&start=2014-01-08T12:00:00Z&window=15&zoom=5",
	ExampleHost: exHost,
	URI:         "/intensity?type=reported&bbox=(bbox)&zoom=(int)&start=(ISO8601 date time)&window=(int)",
	Params: map[string]template.HTML{
		"start": `the date time in ISO8601 format for the start of the time window for the request. Only queries to the nearest minute
						are supported so the parameter must always end in <code>:00Z</code> e.g., <code>2014-01-08T12:00:00Z</code>`,
		"window": `The length of time window in minutes for the request.  Currently only the value <code>15</code> is supported.`,
		"bbox": `A spatial bounding box for the query describing the upper left and lower right longitude and latitude corners.  Integer values only.
						e.g., <code>165,-34,179,-47</code> would include mainland New Zealand.`,
		"zoom": `The zoom level to aggregate values at.  This controls the size of the area that values are aggregated at.  The point returned
						will be the center of each area.  Allowed values are one of <code>5, 6, 7</code>.`,
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
	bbox, zoom, start, window string
}

func (q *intensityReportedQuery) Validate(w http.ResponseWriter, r *http.Request) bool {
	switch {
	case len(r.URL.Query()) != 5:
		web.BadRequest(w, r, "incorrect number of query parameters.")
		return false
	case !web.ParamsExist(w, r, "type", "bbox", "zoom", "window", "start"):
		return false
	case !bboxRe.MatchString(r.URL.Query().Get("bbox")):
		web.BadRequest(w, r, "Invalid bbox")
		return false
	case !zoomRe.MatchString(r.URL.Query().Get("zoom")):
		web.BadRequest(w, r, "Invalid zoom")
		return false
	case !windowRe.MatchString(r.URL.Query().Get("window")):
		web.BadRequest(w, r, "Invalid window")
		return false
	case !startRe.MatchString(r.URL.Query().Get("start")):
		web.BadRequest(w, r, "Invalid start")
		return false
	case r.URL.Query().Get("type") != "reported":
		web.BadRequest(w, r, "Invalid type")
		return false
	}

	q.bbox = r.URL.Query().Get("bbox")
	q.zoom = r.URL.Query().Get("zoom")
	q.window = r.URL.Query().Get("window")
	q.start = r.URL.Query().Get("start")

	return true
}

func (q *intensityReportedQuery) Handle(w http.ResponseWriter, r *http.Request) {
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
								WHERE time >= timestamp with time zone '` + q.start + `'
								AND time <= (timestamp with time zone '` + q.start + `' + interval '` + q.window + ` minutes')
								AND location && ST_MakeEnvelope(` + q.bbox + `, 4326) group by (geohash` + q.zoom + `)) as s
							) As f )  as fc`).Scan(&d)
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
