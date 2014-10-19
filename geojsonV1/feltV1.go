package geojsonV1

import (
	"errors"
	"github.com/GeoNet/geonet-rest/web"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

const feltURL = "http://felt.geonet.org.nz/services/reports/"

// reports fetches felt reports from the existing web service and prints them.
func reports(w http.ResponseWriter, r *http.Request, client *http.Client) {
	p := mux.Vars(r)

	res, err := client.Get(feltURL + p["publicID"] + ".geojson")
	defer res.Body.Close()
	if err != nil {
		web.Fail(w, r, err)
		return
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		web.Fail(w, r, err)
		return
	}

	// Felt returns a 400 when it should probably be a 404.  Tapestry quirk?
	switch {
	case 200 == res.StatusCode:
		web.Win(w, r, b)
		return
	case 4 == res.StatusCode/100:
		web.Nope(w, r, string(b))
		return
	case 5 == res.StatusCode/500:
		web.Fail(w, r, errors.New("error proxying felt resports.  Shrug."))
		return
	}

	web.Fail(w, r, errors.New("unknown response from felt."))
}
