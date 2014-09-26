# GeoNet REST API

Welcome to the GeoNet REST api.

The data provided here is used for the GeoNet web site and other similar services.  If you are looking for data for research or other purposes then please check the full [range of data available](http://info.geonet.org.nz/x/DYAO) from GeoNet.  

## GeoJSON Endpoints

These end points all return [GeoJSON](http://geojson.org/).

### API Versioning

The current version is `version 1`.

All end points use a monotonically increasing version number.  You may specify the version of the API you require using the Accept header.  If you do not specify a version in the header then the highest version of the API is assumed for your request. 

**Example:**

To use version 1 of the API specify the Accept header exactly as below. 

```
Accept: application/vnd.geo+json; version=1;
```

### Version 1 Endpoints:

* [/quake](endpoints/geojsonV1/quakeV1.md)
* [/region](endpoints/geojsonV1/regionV1.md)