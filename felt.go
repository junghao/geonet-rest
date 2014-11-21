package main

import (
	"database/sql"
	"errors"
	"io/ioutil"
	"net/http"
)

const (
	feltURL = "http://felt.geonet.org.nz/services/reports/"
)

// /felt/report?publicID=2013p407387
type feltQuery struct {
	publicID string
}

func (q *feltQuery) validate(w http.ResponseWriter, r *http.Request) bool {
	var d string

	// Check that the publicid exists in the DB.  This is needed as the geoJSON query will return empty
	// JSON for an invalid publicID.
	err := db.QueryRow("select publicid FROM qrt.quake_materialized where publicid = $1", q.publicID).Scan(&d)
	if err == sql.ErrNoRows {
		notFound(w, r, "invalid publicID: "+q.publicID)
		return false
	}
	if err != nil {
		serviceUnavailable(w, r, err)
		return false
	}
	return true
}

func (q *feltQuery) handle(w http.ResponseWriter, r *http.Request) {
	res, err := client.Get(feltURL + q.publicID + ".geojson")
	defer res.Body.Close()
	if err != nil {
		serviceUnavailable(w, r, err)
		return
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		serviceUnavailable(w, r, err)
		return
	}

	// Felt returns a 400 when it should probably be a 404.  Tapestry quirk?
	switch {
	case 200 == res.StatusCode:
		ok(w, r, b)
		return
	case 4 == res.StatusCode/100:
		notFound(w, r, string(b))
		return
	case 5 == res.StatusCode/500:
		serviceUnavailable(w, r, errors.New("error proxying felt resports.  Shrug."))
		return
	}

	serviceUnavailable(w, r, errors.New("unknown response from felt."))
}
