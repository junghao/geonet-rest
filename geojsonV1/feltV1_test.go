//# Felt Reports
//
//##/felt
//
// Look up felt report information.
//
package geojsonV1

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

//## Felt Reports for a Quake.
//
// **GET /felt/report?publicID=(publicID)**
//
// Get Felt Reports for a quake.
//
//### Parameters
//
// * `publicID` - a valid quake identfier.
//
//### Example request:
//
// [/felt/report?publicID=2013p407387](SERVER/felt/report?publicID=2013p407387)
//
func TestReports(t *testing.T) {
	req, _ := http.NewRequest("GET", "/felt/report?publicID=2013p407387", nil)
	res := httptest.NewRecorder()

	serve(req, res)

	if res.Code != 200 {
		t.Errorf("Non 200 error code: %d", res.Code)
	}

	if res.HeaderMap.Get("Content-Type") != "application/vnd.geo+json; version=1;" {
		t.Errorf("incorrect Content-Type")
	}

	req, _ = http.NewRequest("GET", "/felt/report?publicID=2013p40738", nil)
	res = httptest.NewRecorder()

	serve(req, res)

	if res.Code != 404 {
		t.Errorf("Non 404 error code: %d", res.Code)
	}
}
