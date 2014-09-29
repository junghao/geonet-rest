// pretty provides pretty printing functions for the geonet-rest project.
package pretty

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// JSON takes a byte array of JSON and pretty prints it to w.
// If an error is encountered a 500 error is sent w.
func JSON(w http.ResponseWriter, d []byte) (err error) {
	var out bytes.Buffer
	err = json.Indent(&out, d, "", "   ")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return err
	}

	out.WriteTo(w)
	return err
}
