package jsonV1

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"strings"
)

const mlink = "http://info.geonet.org.nz/m/view-rendered-page.action?abstractPageId="
const newsURL = "http://info.geonet.org.nz/createrssfeed.action?types=blogpost&spaces=conf_all&title=GeoNet+News+RSS+Feed&labelString%3D&excludedSpaceKeys%3D&sort=created&maxResults=10&timeSpan=500&showContent=true&publicFeed=true&confirm=Create+RSS+Feed"

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
func news(w http.ResponseWriter, r *http.Request, client *http.Client) {
	res, err := client.Get(newsURL)
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

	e, err := unmarshalNews(b)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	j, err := json.Marshal(e)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var out bytes.Buffer
	err = json.Indent(&out, []byte(j), "", "   ")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	out.WriteTo(w)
}
