package main

import (
	"encoding/json"
	"encoding/xml"
	"github.com/GeoNet/app/web"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	mlink   = "http://info.geonet.org.nz/m/view-rendered-page.action?abstractPageId="
	newsURL = "http://info.geonet.org.nz/createrssfeed.action?types=blogpost&spaces=conf_all&title=GeoNet+News+RSS+Feed&labelString%3D&excludedSpaceKeys%3D&sort=created&maxResults=10&timeSpan=500&showContent=true&publicFeed=true&confirm=Create+RSS+Feed"
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

// /news/geonet

var newsQueryD = &doc{
	Title:       "News",
	Description: " Returns a simple JSON version of the GeoNet News RSS feed.",
	Example:     "/news/geonet",
	URI:         "/news/geonet",
	Params: map[string]template.HTML{
		"none": `no query parameters are requiresd.`,
	},
	Props: map[string]template.HTML{
		"mlink":     "a link to a mobile version of the news story.",
		"link":      "a link to the news story.",
		"published": "the date the story was published",
		"title":     "the title of the story.",
	},
	Result: `{
		"feed": [
		{
		"mlink": "http://info.geonet.org.nz/m/view-rendered-page.action?abstractPageId=12222528",
		"link": "http://info.geonet.org.nz/display/home/2014/11/26/GeoNet+News+Issue+20",
		"published": "2014-11-25T22:00:01Z",
		"title": "GeoNet News Issue 20"
		},
		{
		"mlink": "http://info.geonet.org.nz/m/view-rendered-page.action?abstractPageId=12222520",
		"link": "http://info.geonet.org.nz/display/appdata/2014/11/19/Run%2C+don%27t+walk+to+your+app+store%3A+GeoNet+Quake+app+upgrade+now+available",
		"published": "2014-11-18T23:42:15Z",
		"title": "Run, don't walk to your app store: GeoNet Quake app upgrade now available"
		}]}`,
}

func (q *newsQuery) doc() *doc {
	return newsQueryD
}

type newsQuery struct{}

func (q *newsQuery) validate(w http.ResponseWriter, r *http.Request) bool {
	return true
}

func (q *newsQuery) handle(w http.ResponseWriter, r *http.Request) {
	res, err := client.Get(newsURL)
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

	e, err := unmarshalNews(b)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}

	j, err := json.Marshal(e)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}

	w.Header().Set("Surrogate-Control", web.MaxAge300)

	web.Ok(w, r, &j)
}
