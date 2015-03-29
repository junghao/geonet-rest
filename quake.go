package main

import (
	"database/sql"
	"github.com/GeoNet/web"
	"github.com/GeoNet/web/api/apidoc"
	"html/template"
	"net/http"
	"regexp"
	"strings"
)

// These constants are the length of parts of the URI and are used for
// extracting query params embedded in the URI.
const (
	quakeLen = 7 //  len("/quake/")
)

var quakeDoc = apidoc.Endpoint{Title: "Quake",
	Description: `Look up quake information.`,
	Queries: []*apidoc.Query{
		new(quakeQuery).Doc(),
		new(quakesQuery).Doc(),
		new(quakesRegionQuery).Doc(),
	},
}

var intensityRe = regexp.MustCompile(`^(unnoticeable|weak|light|moderate|strong|severe)$`)
var numberRe = regexp.MustCompile(`^(3|30|100|500|1000|1500)$`)
var qualityRe = regexp.MustCompile(`^(best|caution|deleted|good)$`)
var publicIDRe = regexp.MustCompile(`^[0-9a-z]+$`)

// all requests have the same properties in the return.
// this is a map for all apidoc.Query{} structs.
var propsD = map[string]template.HTML{
	`publicID`:         `the unique public identifier for this quake.`,
	`time`:             `the origin time of the quake.`,
	`depth`:            `the depth of the quake in km.`,
	`magnitude`:        `the summary magnitude for the quake.  This is <b>not</b> Richter magnitude.`,
	`type`:             `the event type; earthquake, landslide etc.`,
	`agency`:           `the agency that located this quake.  The official GNS/GeoNet agency name for this field is WEL(*).`,
	`locality`:         `distance and direction to the nearest locality.`,
	`intensity`:        `the calculated <a href="http://info.geonet.org.nz/x/b4Ih">intensity</a> at the surface above the quake (epicenter) e.g., strong.`,
	`regionIntensity`:  `the calculated intensity at the closest locality in the region for the request. `,
	`quality`:          `the quality of this information; <code>best</code>, <code>good</code>, <code>caution</code>, <code>unknown</code>, <code>deleted</code>.`,
	`modificationTime`: `the modification time of this information.`,
}

// /quake/2013p407387

var quakeQueryD = &apidoc.Query{
	Accept:      web.V1GeoJSON,
	Title:       "Quake",
	Description: "Information for a single quake.",
	Example:     "/quake/2013p407387",
	ExampleHost: exHost,
	URI:         "/quake/(publicID)",
	Params: map[string]template.HTML{
		"publicID": `a valid quake ID e.g., <code>2014p715167</code>`,
	},
	Props: propsD,
}

func (q *quakeQuery) Doc() *apidoc.Query {
	return quakeQueryD
}

type quakeQuery struct {
	publicID string
}

func (q *quakeQuery) Validate(w http.ResponseWriter, r *http.Request) bool {
	if len(r.URL.Query()) != 0 {
		web.BadRequest(w, r, "incorrect number of query parameters.")
		return false
	}

	q.publicID = r.URL.Path[quakeLen:]

	if !publicIDRe.MatchString(q.publicID) {
		web.BadRequest(w, r, "invalid publicID: "+q.publicID)
		return false
	}

	var d string

	// Check that the publicid exists in the DB.  This is needed as the handle method will return empty
	// JSON for an invalid publicID.
	err := db.QueryRow("select publicid FROM qrt.quake_materialized where publicid = $1", q.publicID).Scan(&d)
	if err == sql.ErrNoRows {
		web.NotFound(w, r, "invalid publicID: "+q.publicID)
		return false
	}
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return false
	}
	return true
}

func (q *quakeQuery) Handle(w http.ResponseWriter, r *http.Request) {
	var d string

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
		web.ServiceUnavailable(w, r, err)
		return
	}

	b := []byte(d)
	web.Ok(w, r, &b)
}

// /quake?regionID=newzealand&regionIntensity=unnoticeable&number=30&quality=best,caution,good
var quakesRegionQueryD = &apidoc.Query{
	Accept:      web.V1GeoJSON,
	Title:       "Quakes Possibly Felt in a Region",
	Description: "quakes possibly felt in a region during the last 365 days.",
	Example:     "/quake?regionID=newzealand&regionIntensity=weak&number=3&quality=best,caution,good",
	ExampleHost: exHost,
	URI:         "/quake?regionID=(region)&regionIntensity=(intensity)&number=(n)&quality=(quality)",
	Params: map[string]template.HTML{
		`regionID`: `a valid quake region identifier e.g., <code>newzealand</code>.`,
		`regionIntensity`: `the minimum intensity in the region e.g., <code>weak</code>.  
		Must be one of <code>unnoticeable</code>, <code>weak</code>, <code>light</code>, 
		<code>moderate</code>, <code>strong</code>, <code>severe</code>.`,
		`number`: `the maximum number of quakes to return.  Must be one of 
		<code>3</code>, <code>30</code>, <code>100</code>, <code>500</code>, <code>1000</code>, <code>1500</code>.`,
		`quality`: `a comma separated list of quality values to be included in the response: 
		<code>best</code>, <code>caution</code>, <code>deleted</code>, <code>good</code>.`,
	},
	Props: propsD,
}

func (q *quakesRegionQuery) Doc() *apidoc.Query {
	return quakesRegionQueryD
}

type quakesRegionQuery struct {
	regionID, regionIntensity, number string
	quality                           []string
}

func (q *quakesRegionQuery) Validate(w http.ResponseWriter, r *http.Request) bool {
	switch {
	case len(r.URL.Query()) != 4:
		web.BadRequest(w, r, "incorrect number of query parameters.")
		return false
	case !web.ParamsExist(w, r, "number", "regionID", "regionIntensity", "quality"):
		return false
	case !numberRe.MatchString(r.URL.Query().Get("number")):
		web.BadRequest(w, r, "Invalid query parameter number: "+r.URL.Query().Get("number"))
		return false
	case !intensityRe.MatchString(r.URL.Query().Get("regionIntensity")):
		web.BadRequest(w, r, "Invalid regionIntensity: "+r.URL.Query().Get("regionIntensity"))
		return false
	}

	q.number = r.URL.Query().Get("number")
	q.regionID = r.URL.Query().Get("regionID")
	q.regionIntensity = r.URL.Query().Get("regionIntensity")
	q.quality = strings.Split(r.URL.Query().Get("quality"), ",")

	var d string
	err := db.QueryRow("select regionname FROM qrt.region where regionname = $1 AND groupname in ('region', 'north', 'south')", q.regionID).Scan(&d)
	if err == sql.ErrNoRows {
		web.BadRequest(w, r, "invalid quake regionID: "+q.regionID)
		return false
	}
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return false
	}

	for _, q := range q.quality {
		if !qualityRe.MatchString(q) {
			web.BadRequest(w, r, "Invalid quality: "+q)
			return false
		}
	}

	return true
}

func (q *quakesRegionQuery) Handle(w http.ResponseWriter, r *http.Request) {
	var d string

	err := db.QueryRow(
		`SELECT row_to_json(fc)
                         FROM ( SELECT 'FeatureCollection' as type, COALESCE(array_to_json(array_agg(f)), '[]') as features
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
		web.ServiceUnavailable(w, r, err)
		return
	}

	b := []byte(d)
	web.Ok(w, r, &b)
}

// /quake?regionID=newzealand&intensity=unnoticeable&number=30&quality=best,caution,good

var quakesQueryD = &apidoc.Query{
	Accept:      web.V1GeoJSON,
	Title:       "Quakes in a Region",
	Description: "quakes in a region during the last 365 days.",
	Example:     "/quake?regionID=newzealand&intensity=weak&number=3&quality=best,caution,good",
	ExampleHost: exHost,
	URI:         " /quake?regionID=(region)&intensity=(intensity)&number=(n)&quality=(quality)",
	Params: map[string]template.HTML{
		`regionID`: `a valid quake region identifier e.g., <code>newzealand</code>.`,
		`intensity`: `the minimum intensity at the epicenter e.g., <code>weak</code>.  
		Must be one of <code>unnoticeable</code>, <code>weak</code>, <code>light</code>, 
		<code>moderate</code>, <code>strong</code>, <code>severe</code>.`,
		`number`: `the maximum number of quakes to return.  Must be one of 
		<code>3</code>, <code>30</code>, <code>100</code>, <code>500</code>, <code>1000</code>, <code>1500</code>.`,
		`quality`: `a comma separated list of quality values to be included in the response: 
		<code>best</code>, <code>caution</code>, <code>deleted</code>, <code>good</code>.`,
	},
	Props: propsD,
}

func (q *quakesQuery) Doc() *apidoc.Query {
	return quakesQueryD
}

type quakesQuery struct {
	regionID, intensity, number string
	quality                     []string
}

func (q *quakesQuery) Validate(w http.ResponseWriter, r *http.Request) bool {
	switch {
	case len(r.URL.Query()) != 4:
		web.BadRequest(w, r, "incorrect number of query parameters.")
		return false
	case !web.ParamsExist(w, r, "number", "regionID", "intensity", "quality"):
		return false
	case !numberRe.MatchString(r.URL.Query().Get("number")):
		web.BadRequest(w, r, "Invalid query parameter number: "+r.URL.Query().Get("number"))
		return false
	case !intensityRe.MatchString(r.URL.Query().Get("intensity")):
		web.BadRequest(w, r, "Invalid intensity: "+r.URL.Query().Get("intensity"))
		return false
	}

	q.number = r.URL.Query().Get("number")
	q.regionID = r.URL.Query().Get("regionID")
	q.intensity = r.URL.Query().Get("intensity")
	q.quality = strings.Split(r.URL.Query().Get("quality"), ",")

	var d string
	err := db.QueryRow("select regionname FROM qrt.region where regionname = $1 AND groupname in ('region', 'north', 'south')", q.regionID).Scan(&d)
	if err == sql.ErrNoRows {
		web.BadRequest(w, r, "invalid quake regionID: "+q.regionID)
		return false
	}
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return false
	}

	for _, q := range q.quality {
		if !qualityRe.MatchString(q) {
			web.BadRequest(w, r, "Invalid quality: "+q)
			return false
		}
	}

	return true
}

func (q *quakesQuery) Handle(w http.ResponseWriter, r *http.Request) {
	var d string

	err := db.QueryRow(
		`SELECT row_to_json(fc)
                         FROM ( SELECT 'FeatureCollection' as type, COALESCE(array_to_json(array_agg(f)), '[]') as features
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
		web.ServiceUnavailable(w, r, err)
		return
	}
	b := []byte(d)
	web.Ok(w, r, &b)
}
