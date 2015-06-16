package main

import (
	"github.com/GeoNet/web"
	"github.com/GeoNet/web/api/apidoc"
	"html/template"
	"net/http"
)

var volcanoDoc = apidoc.Endpoint{Title: "Volcano",
	Description: "Look up volcano information.  <b>Caution - under development, subject to change.</b>",
	Queries: []*apidoc.Query{
		alertLevelD,
	},
}

var alertLevelD = &apidoc.Query{
	Accept:      web.V1GeoJSON,
	Title:       "Volcanic Alert Level",
	Description: `Volcanic Alert Level information for all volcanoes.`,
	Discussion:  `<p>Volcanic Alert Level information for all volcanoes.  Please refer to <a href="http://info.geonet.org.nz/x/PYAO">Volcanic Alert Levels</a> for additional information.</p>`,
	Example:     "/volcano/alertlevel",
	ExampleHost: exHost,
	URI:         "/volcano/alertlevel",
	Required: map[string]template.HTML{
		"none": `no query parameters required.`,
	},
	Props: map[string]template.HTML{
		`volcanoID`:    `a unique identifier for the volcano.`,
		`volcanoTitle`: `the volcano title.`,
		`level`:        `volcanic alert level.`,
		`activity`:     `volcanic activity.`,
		`hazards`:      `most likely hazards.`,
	},
}

func alertLevel(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Query()) != 0 {
		web.BadRequest(w, r, "incorrect number of query parameters.")
		return
	}

	var d string

	err := db.QueryRow(`SELECT row_to_json(fc)
                         FROM ( SELECT 'FeatureCollection' as type, array_to_json(array_agg(f)) as features
                         FROM (SELECT 'Feature' as type,
                         ST_AsGeoJSON(v.location)::json as geometry,
                         row_to_json((SELECT l FROM 
                         	(
                         		SELECT 
                                id AS "volcanoID",
                                title AS "volcanoTitle",
                                alert_level as "level",
                                activity,
                                hazards 
                           ) as l
                         )) as properties FROM (qrt.volcano JOIN qrt.volcanic_alert_level using (alert_level)) as v ) As f )  as fc`).Scan(&d)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}

	b := []byte(d)
	web.Ok(w, r, &b)
}
