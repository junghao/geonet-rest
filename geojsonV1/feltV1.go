package geojsonV1

import (
	"github.com/GeoNet/geonet-rest/pretty"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

const feltURL = "http://felt.geonet.org.nz/services/reports/"

// reports fetches felt reports from the existing web service and pretty prints them.
func reports(w http.ResponseWriter, r *http.Request, client *http.Client) {
	p := mux.Vars(r)

	res, err := client.Get(feltURL + p["publicID"] + ".geojson")
	defer res.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Errors for non 200 response.  Felt returns a 400 when it should probably be a 404.
	// Tapestry quirk?
	if res.StatusCode != 200 {
		if res.StatusCode == 400 {
			http.Error(w, string(b), 404)
		} else {
			http.Error(w, string(b), res.StatusCode)
		}
		return
	}

	pretty.JSON(w, b)
}
