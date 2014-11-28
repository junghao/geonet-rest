package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"
)

var funcMap = template.FuncMap{
	"pretty": pretty,
	"anchor": anchor,
}

const apiHost = `api.geonet.org.nz`

var t = template.Must(template.New("all").Funcs(funcMap).Parse(templ))

// doc holds query documentation.  Any template.HTML types can hold markup.
type doc struct {
	Title       string
	URI         string
	Example     string                   // example URI for a query
	Description string                   // a short description.
	Discussion  template.HTML            // Optional. Any longer discussion.  Inserted verbatim so use <p> tags etc.
	Params      map[string]template.HTML // query parameters
	Props       map[string]template.HTML // response properties
	Result      string                   // example json.  This is pretty printed into the html or ommited if it doesn't parse for some reason.
}

// endpoint holds documentation for an api endpoint e.g., /impact/intensity
type endpoint struct {
	Title       string
	Description template.HTML
	Discussion  template.HTML
	Queries     []*doc
}

// endpoints documentation.  The string in the map is used to match pages names
// in the URI e.g., intensity -> /api-docs/endpoint/intensity
var endpoints = map[string]*endpoint{
	"quake": &endpoint{
		Title:       "Quake",
		Description: `Look up quake information.`,
		Discussion: `<p>All queries return <a href="http://geojson.org/">GeoJSON</a>.  Specify your <code>Accept</code> header exactly as:
		<code>Accept: application/vnd.geo+json;version=1</code></p>`,
		Queries: []*doc{
			new(quakeQuery).doc(),
			new(quakesQuery).doc(),
			new(quakesRegionQuery).doc(),
		},
	},
	"region": &endpoint{
		Title:       "Region",
		Description: `Look up region information.`,
		Discussion: `<p>All queries return <a href="http://geojson.org/">GeoJSON</a> with Point features.  Specify your <code>Accept</code> header exactly as:
		<code>Accept: application/vnd.geo+json;version=1</code></p>`,
		Queries: []*doc{
			new(regionQuery).doc(),
			new(regionsQuery).doc(),
		},
	},
	"felt": &endpoint{
		Title:       "Felt",
		Description: `Look up Felt Report information.`,
		Discussion: `<p>All queries return <a href="http://geojson.org/">GeoJSON</a>.  Specify your <code>Accept</code> header exactly as:
		<code>Accept: application/vnd.geo+json;version=1</code></p>`,
		Queries: []*doc{
			new(feltQuery).doc(),
		},
	},
	"news": &endpoint{
		Title:       "News",
		Description: `GeoNet news stories.`,
		Discussion: `<p>All queries return JSON.  Specify your <code>Accept</code> header exactly as:
		<code>Accept: application/json;version=1</code></p>`,
		Queries: []*doc{
			new(newsQuery).doc(),
		},
	},
}

// We only need a few short templates so inline them to keep deployment simpler.
const (
	templ = `{{define "header"}}
			<html>
			<head>
			<meta charset="utf-8"/>
			<meta http-equiv="X-UA-Compatible" content="IE=edge"/>
			<meta name="viewport" content="width=device-width, initial-scale=1"/>
			<title>GeoNet API</title>
			<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.1/css/bootstrap.min.css">
			<style>
			body { padding-top: 60px; }
			a.anchor { 
				display: block; position: relative; top: -60px; visibility: hidden; 
			}
			.footer {
				margin-top: 20px;
				padding: 20px 0 20px;
				border-top: 1px solid #e5e5e5;
			}

			.footer p {
				text-align: center;
			}

			#logo{position:relative;}
			#logo li{margin:0;padding:0;list-style:none;position:absolute;top:0;}
			#logo li a span
			{
				position: absolute;
				left: -10000px;
			}

			#gns li, #gns a
			{
				float: left;
				display:block;
				height: 90px;
				width: 54px;
			}

			#gns{left:-20px;height:90px;width:54px;}
			#gns{background:url('http://static.geonet.org.nz/geonet-2.0.2/images/logos.png') -0px -0px;}

			#eqc li, #eqc a
			{
				display:block;
				height: 61px;
				width: 132px;
			}

			#eqc{right:0px;height:79px;width:132px;}
			#eqc{background:url('http://static.geonet.org.nz/geonet-2.0.2/images/logos.png') -0px -312px;}

			#ccby li, #ccby a
			{
				display:block;
				height: 15px;
				width: 80px;
			}
			#ccby{left:15px;height:15px;width:80px; }
			#ccby{background:url('http://static.geonet.org.nz/geonet-2.0.2/images/logos.png') -0px -100px;}

			#geonet{
				background:url('http://static.geonet.org.nz/geonet-2.0.2/images/logos.png') 0px -249px; 
				width:137px; 
				height:53px;
				display:block;
			}


			</style>
			</head>
			<body>
			<div class="navbar navbar-inverse navbar-fixed-top" role="navigation">
			<div class="container">
			<div class="navbar-header">
			<a class="navbar-brand" href="http://geonet.org.nz">GeoNet</a>
			</div>
			</div>
			</div>

			<div class="container-fluid">
			{{if not .Production}}
			<div class="alert alert-danger" role="alert">So you found this API just laying around on the internet and that's cool.
			If you're seeing this message then we still view this as experimental or beta so if you use this thing you found
			then please be aware that we may change it or take it away without warning.  If you have some feed back on the 
			API functionality then please write your comment on a box of New Zealand craft IPA and mail it to us.  
			Multiple submissions welcome.</div>
			{{end}}
			{{end}}

			{{define "footer"}}
			<div id="footer" class="footer">
			<div class="row">
			<div class="col-sm-3 hidden-xs">
			<ul id="logo">
			<li id="geonet"><a target="_blank" href="http://www.geonet.org.nz"><span>GeoNet</span></a></li>
			</ul>            
			</div>

			<div class="col-sm-6">
			<p>GeoNet is a collaboration between the <a target="_blank" href="http://www.eqc.govt.nz">Earthquake Commission</a> and <a target="_blank" href="http://www.gns.cri.nz/">GNS Science</a>.</p>
			<p><a target="_blank" href="http://info.geonet.org.nz/x/loYh">about</a> | <a target="_blank" href="http://info.geonet.org.nz/x/JYAO">contact</a> | <a target="_blank" href="http://info.geonet.org.nz/x/RYAo">privacy</a> | <a target="_blank" href="http://info.geonet.org.nz/x/EIIW">disclaimer</a> </p>
			<p>GeoNet content is copyright <a target="_blank" href="http://www.gns.cri.nz/">GNS Science</a> and is licensed under a <a rel="license" target="_blank" href="http://creativecommons.org/licenses/by/3.0/nz/">Creative Commons Attribution 3.0 New Zealand License</a></p>
			</div>

			<div  class="col-sm-2 hidden-xs">
			<ul id="logo">
			<li id="eqc"><a target="_blank" href="http://www.eqc.govt.nz" ><span>EQC</span></a></li>
			</ul>
			</div>
			<div  class="col-sm-1 hidden-xs">
			<ul id="logo">
			<li id="gns"><a target="_blank" href="http://www.gns.cri.nz"><span>GNS Science</span></a></li>
			</ul>  
			</div>
			</div>

			<div class="row">
			<div class="col-sm-1 col-sm-offset-5 hidden-xs">
			<ul id="logo">
			<li id="ccby"><a href="http://creativecommons.org/licenses/by/3.0/nz/" ><span>CC-BY</span></a></li>
			</ul>
			</div>
			</div>

			</div>
			</div>
			</body>
			</html>
			{{end}}

			{{define "index"}}
			{{template "header" .Header}}
			<ol class="breadcrumb">
			<li class="active"><a href="/api-docs">Index</a></li>
			</ol>
			<h1 class="page-header">GeoNet API</h1>
			<p class="lead">Welcome to the GeoNet API.</p>

			<p>The GeoNet project makes all its data and images freely available.
			Please ensure you have read and understood our 
			<a href="http://info.geonet.org.nz/x/BYIW">Data Policy</a> and <a href="http://info.geonet.org.nz/x/EIIW">Disclaimer</a> 
			before using any of these services.</p>

			<p>The data provided here is used for the GeoNet web site and other similar services. 
			If you are looking for data for research or other purposes then please check the <a href="http://info.geonet.org.nz/x/DYAO">full 
			range of data</a> available from GeoNet. </p>
			
			<p>Endpoints return either <a href="http://geojson.org/">GeoJSON</a> or <a href="http://www.json.org/">JSON</a> depending 
			on the nature of the content for the endpoint.</p>
			
			<p>This API is versioned via the Accept header.  The current version is <code>version=1</code></p>  
			
			<p>To use <code>version=1</code> of the API check the <code>Accept</code> header specified for the 
			endpoint you are using then specify the <code>Accept</code> header on your requests exactly as either:<br/><br/>
			<code>Accept: application/vnd.geo+json;version=1</code> or <code>Accept: application/json;version=1</code>
			</p>
			
			<p>This is a little painful when exploring the API but it pays dividends in the future for any client that you write.  
			We use the <a href="https://github.com/stedolan/jq">jq</a> command for JSON pretty printing etc.  A curl command for exploring API 
			endpoints that return GeoJSON then looks like:</p>

			<pre>curl -H "Accept: application/vnd.geo+json;version=1" "http://...API-QUERY..." | jq .</pre>

			<p>You may also be able to find a browser plugin to help with setting the Accept header for requests.</p>

			<h2 class="page-header">Endpoints</h2>

			<p>The following endpoints are available:</p>
			<ul class=>
			{{range $k, $v := .Endpoints}}
			<li><a href="/api-docs/endpoint/{{$k}}">/{{$k}}</a> - {{$v.Description}}</li>
			{{end}}
			</ul>
			{{template "footer"}}
			{{end}}


			{{define "endpoint"}}
			{{template "header" .Header}}
			<ol class="breadcrumb">
			<li><a href="/api-docs">Index</a></li>
			<li>Endpoint</li>
			<li class="active">{{.Endpoint.Title}}</li>
			</ol>
			<h1 class="page-header">{{.Endpoint.Title}}</h1>
			<p class="lead">{{.Endpoint.Description}}</p>
			{{.Endpoint.Discussion}}
			<h4>Query Index:</h4>
			{{range .Endpoint.Queries}} 
			<ul>
			<li><a href="#{{anchor .Title}}">{{.Title}}</a> - {{.Description}}</li>
			</ul>
			{{end}}

			{{range .Endpoint.Queries}} 
			<a id="{{anchor .Title}}" class="anchor"></a>
			<h2 class="page-header">{{.Title}}</h2>
			<p class="lead">{{.Description}}</p>
			{{.Discussion}}
			<b>GET</b> <code>{{.URI}}</code>
			<h3>Query Parameters</h3>
			<dl>
			{{range $k, $v := .Params}}
			<dt>{{$k}}</dt>
			<dd>{{$v}}</dd>
			{{end}}
			</dl>
			<h3>Response Properties</h3>
			<dl>
			{{range $k, $v := .Props}}
			<dt>{{$k}}</dt>
			<dd>{{$v}}</dd>
			{{end}}
			</dl>
			<h3>Example Query</h3>
			<code>http://{{$.APIHost}}{{.Example}}</code>
			<br />
			<h3>Example Response</h3>
			<p>Example of response data - this is not necessarily for the example query.</p>
			<pre>{{pretty .Result}}</pre>
			{{end}}
			{{template "footer"}}
			{{end}}
			`
)

// structs for web page template data.

type headerT struct {
	Production bool
}

var headerD = headerT{
	Production: config.Production,
}

type indexT struct {
	Header    headerT
	Endpoints map[string]*endpoint
}

var indexD = indexT{
	Header:    headerD,
	Endpoints: endpoints,
}

type endpointT struct {
	Header   headerT
	Endpoint *endpoint
	APIHost  string
}

// documentation queries

type indexQuery struct{}

var indexQueryD = &doc{
	Title: "No docs for the index docs query",
}

func (q *indexQuery) validate(w http.ResponseWriter, r *http.Request) bool {
	return true
}

func (q *indexQuery) doc() *doc {
	return indexQueryD
}

func (q *indexQuery) handle(w http.ResponseWriter, r *http.Request) {
	var b bytes.Buffer
	if err := t.ExecuteTemplate(&b, "index", indexD); err != nil {
		log.Println(err)
		serviceUnavailable(w, r, err)
		return
	}
	okBuf(w, r, &b)
}

var endpointQueryD = &doc{
	Title: "No docs for the docs query.",
}

type endpointQuery struct {
	e string
}

func (q *endpointQuery) validate(w http.ResponseWriter, r *http.Request) bool {
	if _, ok := endpoints[q.e]; !ok {
		notFound(w, r, "page not found.")
		return false
	}
	return true
}

func (q *endpointQuery) handle(w http.ResponseWriter, r *http.Request) {
	var b bytes.Buffer
	if err := t.ExecuteTemplate(&b, "endpoint", &endpointT{Header: headerD, Endpoint: endpoints[q.e], APIHost: apiHost}); err != nil {
		log.Println(err)
		serviceUnavailable(w, r, err)
		return
	}
	okBuf(w, r, &b)
}

func (q *endpointQuery) doc() *doc {
	return endpointQueryD
}

// Template functions.

// pretty JSON pretty prints the JSON in the string j.  Returns an empty string if unmarshalling j fails.
func pretty(j string) (p string) {
	p = ""

	var dat map[string]interface{}

	if err := json.Unmarshal([]byte(j), &dat); err != nil {
		return p
	}

	if d, err := json.MarshalIndent(dat, "   ", "  "); err == nil {
		p = string(d)
	}

	return p
}

// anchor lowercases and removes all white space from s.  This
func anchor(s string) (a string) {
	a = strings.TrimSpace(s)
	a = strings.ToLower(a)
	a = strings.Replace(a, " ", "", -1)

	return a
}
