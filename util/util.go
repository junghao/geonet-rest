// util provides utility functions for the geonet-rest project.
package util

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// PrettyJSON takes a JSON string and pretty prints it to w.
func PrettyJSON(w http.ResponseWriter, d string) {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(d), "", "   ")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	out.WriteTo(w)
}
