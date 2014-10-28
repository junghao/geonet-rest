# Quake Information

##/quake

 Look up quake information.  All calls return [GeoJSON](http://geojson.org/) with Point features.

### Quake Properties

 Each quake in the returned geojson has properties associated with it.
 Please follow this link for information about how the [quake properties](http://info.geonet.org.nz/x/J4IW) are derived.

 * `publicID` - the unique public identifier for this quake.
 * `time` - the origin time of the quake.
 * `depth` - the depth of the quake in km.
 * `magnitude` - the summary magnitude for the quake.  This is *not* Richter magnitude.
 * `type` - the event type; earthquake, landslide etc.
 * `agency` - the agency that located this quake.  The official GNS/GeoNet agency name for this field is WEL(*).
 * `locality` - distance and direction to the nearest locality.
 * `intensity` - the calculated [intensity](http://info.geonet.org.nz/x/b4Ih) at the surface above the quake (epicenter) e.g., `strong`.
 * `regionIntensity` - the calculated intensity at the closest locality in the region for the request.  If no region is specified for the query then this is the intensity in the `newzealand` region.
 * `quality` - the quality of this information; `best`, `good`, `caution`, `unknown`, `deleted`.
 * `modificationTime` - the modification time of this information.
## Single Quake

  **GET /quake/(publicID)**

 Get information for a single quake.

### Parameters

 * `publicID` - a valid quake ID e.g., `2014p715167`.

### Example request:

 [/quake/2013p407387](http://ec2-54-253-219-100.ap-southeast-2.compute.amazonaws.com:8080/quake/2013p407387)

## Quakes in a Region

 **GET /quake?regionID=(region)&intensity=(intensity)&number=(n)&quality=(quality)**

 Get quake information from the last 365 days.
 If no quakes are found for the query parameters then a null features array is returned.

### Parameters

 * `regionID` - a valid quake region identifier e.g., `newzealand`.
 * `intensity` - the minimum intensity at the epicenter e.g., `weak`.  Must be one of `unnoticeable`, `weak`, `light`, `moderate`, `strong`, `severe`.
 * `number` - the maximum number of quakes to return.  Must be one of `30`, `100`, `500`, `1000`, `1500`.
 * `quality` - a comma separated list of quality values to be included in the response; `best`, `caution`, `deleted`, `good`.

 *The `number` of quakes that can be returned is restricted to a range of options to improve caching.*

### Example request:

 [/quake?regionID=newzealand&intensity=weak&number=30](http://ec2-54-253-219-100.ap-southeast-2.compute.amazonaws.com:8080/quake?regionID=newzealand&intensity=weak&number=30&quality=best,caution,deleted,good)

