
# GeoNet REST API

Welcome to the GeoNet REST api.

The data provided here is used for the GeoNet web site and other similar services.  If you are looking for data for research or other purposes then please check the full [range of data available](http://info.geonet.org.nz/x/DYAO) from GeoNet.  


### API Versioning

The current version is `version 1`.

All end points use a monotonically increasing version number.  You may specify the version of the API you require using the Accept header.  If you do not specify a version in the header then the highest version of the API is assumed for your request. 

There are two classes of endpoint; those returning GeoJSON and those returning plain JSON.  Use the correct MIME type (along with the version number) in your Accept header depending on which class of endpoint you are calling.  See the examples below.

### Version 1 GeoJSON Endpoints:

These end points all return [GeoJSON](http://geojson.org/).

To use version 1 of the GeoJSON endpoints specify the Accept header exactly as below. 

```
Accept: application/vnd.geo+json; version=1;
```

* [/quake](endpoints/quakeV1.md)
* [/region](endpoints/regionV1.md)
* [/felt](endpoints/feltV1.md)


### Version 1 JSON Endpoints:

To use version 1 of the JSON endpoints specify the Accept header exactly as below. 

```
Accept: application/json; version=1;
```

* [/news](endpoints/newsV1.md)

