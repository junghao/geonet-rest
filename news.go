package main

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"strings"
)

// Feed is used for unmarshaling XML (from the GeoNet RSS news feed)
// and marshaling JSON
type Feed struct {
	Entries []Entry `xml:"entry" json:"feed"`
}

// Entry is used for unmarshaling XML and marshaling JSON.
// JSON tags with a - will not be include in the output.
type Entry struct {
	Title     string `xml:"title" json:"title"`
	Published string `xml:"published" json:"published"`
	Link      Link   `xml:"link" json:"-"`
	Id        string `xml:"id" json:"-"`
	Href      string `json:"link"`
	MHref     string `json:"mlink"`
}

// Link is used for unmarshaling XML.
type Link struct {
	Href string `xml:"href,attr"`
}

// unmarshalNews unmarshals the GeoNet News RSS XML.
func unmarshalNews(b []byte) (f Feed, err error) {
	err = xml.Unmarshal(b, &f)
	if err != nil {
		return f, err
	}

	// Copy the story link and make the link to the
	// mobile friendly version of the story.
	for i, _ := range f.Entries {
		f.Entries[i].Href = f.Entries[i].Link.Href
		f.Entries[i].MHref = mlink + strings.Split(f.Entries[i].Id, "-")[1]
	}

	return f, err
}

// news fetches the GeoNet News RSS feed and converts it to simple JSON.
func newsV1(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", v1JSON)

	// will only be geonet at the moment.
	newsID := r.URL.Path[len("/news/geonet"):]

	// check there isn't extra stuff in the URL - like a cache buster
	if len(r.URL.Query()) > 0 || strings.Contains(newsID, "/") || strings.Contains(newsID, ";") {
		badRequest(w, r, "detected extra stuff in the URL.")
		return
	}

	res, err := client.Get(newsURL)
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

	e, err := unmarshalNews(b)
	if err != nil {
		serviceUnavailable(w, r, err)
		return
	}

	j, err := json.Marshal(e)
	if err != nil {
		serviceUnavailable(w, r, err)
		return
	}

	w.Header().Set("Surrogate-Control", cacheMedium)

	ok(w, r, j)
}
