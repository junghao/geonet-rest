package main

import (
	"database/sql"
	"github.com/GeoNet/app/web"
	"html/template"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var intensityRe = regexp.MustCompile(`^(unnoticeable|weak|light|moderate|strong|severe)$`)
var numberRe = regexp.MustCompile(`^(3|30|100|500|1000|1500)$`)
var qualityRe = regexp.MustCompile(`^(best|caution|deleted|good)$`)

// all requests have the same properties in the return.
// this is a map for all doc{} structs.
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

var quakeQueryD = &doc{
	Title:       "Quake",
	Description: "Information for a single quake.",
	Example:     "/quake/2013p407387",
	URI:         "/quake/(publicID)",
	Params: map[string]template.HTML{
		"publicID": `a valid quake ID e.g., <code>2014p715167</code>`,
	},
	Props: propsD,
	Result: `{"type":"FeatureCollection","features":[{"type":"Feature",
	"geometry":{"type":"Point","coordinates":[172.94479,-43.359699]},
	"properties":{"publicID":"2013p407387","time":"2013-05-31T17:36:02.129Z","depth":20.334389,
	"magnitude":4.027879,"type":"","agency":"WEL(GNS_Primary)","locality":"25 km south-east of Amberley",
	"intensity":"moderate","regionIntensity":"light","quality":"best","modificationTime":"2013-05-31T21:37:41.549Z"}}]}`,
}

func (q *quakeQuery) doc() *doc {
	return quakeQueryD
}

type quakeQuery struct {
	publicID string
}

func (q *quakeQuery) validate(w http.ResponseWriter, r *http.Request) bool {
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
		web.ServiceUnavailable(w, r, err)
		return
	}
	web.DBTime.Track(start, "DB quakeV1")

	b := []byte(d)
	web.Ok(w, r, &b)
}

// /quake?regionID=newzealand&regionIntensity=unnoticeable&number=30&quality=best,caution,good

var quakesRegionQueryD = &doc{
	Title:       "Quakes Possibly Felt in a Region",
	Description: "quakes possibly felt in a region during the last 365 days.",
	Example:     "/quake?regionID=newzealand&regionIntensity=weak&number=3&quality=best",
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
	Result: `{"type":"FeatureCollection","features":[{"type":"Feature",
	"geometry":{"type":"Point","coordinates":[176.9824732,-37.94838701]},
	"properties":{"publicID":"2014p894006","time":"2014-11-27T18:46:52.925Z","depth":5.05859375,
	"magnitude":2.312619978,"type":"earthquake","agency":"WEL(GNS_Primary)",
	"locality":"Within 5 km of Whakatane","intensity":"weak","regionIntensity":"weak",
	"quality":"best","modificationTime":"2014-11-27T19:51:26.754Z"}},{"type":"Feature","geometry":
	{"type":"Point","coordinates":[176.942929,-37.92112384]},"properties":{"publicID":"2014p893874",
	"time":"2014-11-27T17:36:15.092Z","depth":9.453125,"magnitude":3.109248459,"type":"earthquake",
	"agency":"WEL(GNS_Primary)","locality":"5 km north-west of Whakatane","intensity":"light","regionIntensity":
	"light","quality":"best","modificationTime":"2014-11-27T18:27:57.570Z"}},{"type":"Feature",
	"geometry":{"type":"Point","coordinates":[175.3091607,-39.11718634]},"properties":
	{"publicID":"2014p893844","time":"2014-11-27T17:20:32.282Z","depth":14.19921875,
	"magnitude":3.036859794,"type":"earthquake","agency":"WEL(GNS_Primary)",
	"locality":"25 km south of Taumarunui","intensity":"light","regionIntensity":
	"weak","quality":"best","modificationTime":"2014-11-27T18:30:12.251Z"}}]}`,
}

func (q *quakesRegionQuery) doc() *doc {
	return quakesRegionQueryD
}

type quakesRegionQuery struct {
	regionID, regionIntensity, number string
	quality                           []string
}

func (q *quakesRegionQuery) validate(w http.ResponseWriter, r *http.Request) bool {

	if !numberRe.MatchString(q.number) {
		web.BadRequest(w, r, "Invalid number: "+q.number)
		return false
	}

	if !intensityRe.MatchString(q.regionIntensity) {
		web.BadRequest(w, r, "Invalid region intensity: "+q.regionIntensity)
		return false
	}

	if _, ok := quakeRegion[q.regionID]; !ok {
		web.BadRequest(w, r, "Invalid regionID: "+q.regionID)
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

func (q *quakesRegionQuery) handle(w http.ResponseWriter, r *http.Request) {
	var d string

	start := time.Now()
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
	web.DBTime.Track(start, "DB quakeRegionV1")

	b := []byte(d)
	web.Ok(w, r, &b)
}

// /quake?regionID=newzealand&intensity=unnoticeable&number=30&quality=best,caution,good

var quakesQueryD = &doc{
	Title:       "Quakes in a Region",
	Description: "quakes in a region during the last 365 days.",
	Example:     "/quake?regionID=newzealand&intensity=weak&number=3&quality=best",
	URI:         " /quake?regionID=(region)&intensity=(intensity)&number=(n)&quality=(quality)",
	Params: map[string]template.HTML{
		`regionID`: `a valid quake region identifier e.g., <code>newzealand</code>.`,
		`ntensity`: `the minimum intensity at the epicenter e.g., <code>weak</code>.  
		Must be one of <code>unnoticeable</code>, <code>weak</code>, <code>light</code>, 
		<code>moderate</code>, <code>strong</code>, <code>severe</code>.`,
		`number`: `the maximum number of quakes to return.  Must be one of 
		<code>3</code>, <code>30</code>, <code>100</code>, <code>500</code>, <code>1000</code>, <code>1500</code>.`,
		`quality`: `a comma separated list of quality values to be included in the response: 
		<code>best</code>, <code>caution</code>, <code>deleted</code>, <code>good</code>.`,
	},
	Props: propsD,
	Result: `{"type":"FeatureCollection","features":[{"type":"Feature","geometry":
	{"type":"Point","coordinates":[176.9824732,-37.94838701]},"properties":{"publicID":"2014p894006",
	"time":"2014-11-27T18:46:52.925Z","depth":5.05859375,"magnitude":2.312619978,"type":"earthquake",
	"agency":"WEL(GNS_Primary)","locality":"Within 5 km of Whakatane","intensity":"weak","regionIntensity"
	:"weak","quality":"best","modificationTime":"2014-11-27T19:51:26.754Z"}},
	{"type":"Feature","geometry":{"type":"Point","coordinates":[176.942929,-37.92112384]},
	"properties":{"publicID":"2014p893874","time":"2014-11-27T17:36:15.092Z","depth":9.453125,
	"magnitude":3.109248459,"type":"earthquake","agency":"WEL(GNS_Primary)","locality":
	"5 km north-west of Whakatane","intensity":"light","regionIntensity":"light","quality":"best",
	"modificationTime":"2014-11-27T18:27:57.570Z"}},{"type":"Feature","geometry":{"type":
	"Point","coordinates":[177.0859355,-36.7157567]},"properties":{"publicID":"2014p893857",
	"time":"2014-11-27T17:27:16.545Z","depth":20.234375,"magnitude":3.313044562,"type":
	"earthquake","agency":"WEL(GNS_Primary)","locality":"90 km north of White Island","intensity":
	"light","regionIntensity":"unnoticeable","quality":"best","modificationTime":"2014-11-27T20:03:47.907Z"}}]}`,
}

func (q *quakesQuery) doc() *doc {
	return quakesQueryD
}

type quakesQuery struct {
	regionID, intensity, number string
	quality                     []string
}

func (q *quakesQuery) validate(w http.ResponseWriter, r *http.Request) bool {

	if !numberRe.MatchString(q.number) {
		web.BadRequest(w, r, "Invalid number: "+q.number)
		return false
	}

	if !intensityRe.MatchString(q.intensity) {
		web.BadRequest(w, r, "Invalid intensity: "+q.intensity)
		return false
	}

	if _, ok := quakeRegion[q.regionID]; !ok {
		web.BadRequest(w, r, "Invalid regionID: "+q.regionID)
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

func (q *quakesQuery) handle(w http.ResponseWriter, r *http.Request) {
	var d string

	start := time.Now()
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
	web.DBTime.Track(start, "DB quakesV1")

	b := []byte(d)
	web.Ok(w, r, &b)
}
