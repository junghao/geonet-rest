# Region Information

##/region

 Look up region information.  All calls return [GeoJSON](http://geojson.org/) with Polygon features.

### Properties

 * regionID - a unique indentifier for the region.
 * title - the region title.
 * group - the region group.

## Quake Regions

 **GET /region?type=quake**

 Get all quake regions.

### Example request:

 [/region?type=quake](http://ec2-54-253-219-100.ap-southeast-2.compute.amazonaws.com:8080/region?type=quake)

## Single Region

 **GET /region/(regionID)**

 Get a single region.

### Example request:

 [/region/wellington](http://ec2-54-253-219-100.ap-southeast-2.compute.amazonaws.com:8080/region/wellington)

