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

## Single Region

 **GET /region/(regionID)**

 Get a single region.

### Example request:

 `/region/wellington`

