package main

import (
	"database/sql"
	"errors"
	"github.com/GeoNet/app/web"
	"html/template"
	"io/ioutil"
	"net/http"
)

const (
	feltURL = "http://felt.geonet.org.nz/services/reports/"
)

// /felt/report?publicID=2013p407387

var feltQueryD = &doc{
	Title:       "Felt",
	Description: "Look up Felt Report information about earthquakes",
	Example:     "/felt/report?publicID=2013p407387",
	URI:         "/felt/report?publicID=(publicID)",
	Params: map[string]template.HTML{
		"publicID": `a valid quake ID e.g., <code>2014p715167</code>`,
	},
	Props: map[string]template.HTML{
		"todo": `todo`,
	},
	Result: `todo`,
}

func (q *feltQuery) doc() *doc {
	return feltQueryD
}

type feltQuery struct {
	publicID string
}

func (q *feltQuery) validate(w http.ResponseWriter, r *http.Request) bool {
	var d string

	// Check that the publicid exists in the DB.  This is needed as the geoJSON query will return empty
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

func (q *feltQuery) handle(w http.ResponseWriter, r *http.Request) {
	res, err := client.Get(feltURL + q.publicID + ".geojson")
	defer res.Body.Close()
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}

	// Felt returns a 400 when it should probably be a 404.  Tapestry quirk?
	switch {
	case 200 == res.StatusCode:
		web.Ok(w, r, &b)
		return
	case 4 == res.StatusCode/100:
		web.NotFound(w, r, string(b))
		return
	case 5 == res.StatusCode/500:
		web.ServiceUnavailable(w, r, errors.New("error proxying felt resports.  Shrug."))
		return
	}

	web.ServiceUnavailable(w, r, errors.New("unknown response from felt."))
}
