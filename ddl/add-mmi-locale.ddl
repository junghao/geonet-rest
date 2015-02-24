BEGIN;

-- Algorithm described in https://github.com/GeoNet/quakes/issues/159
CREATE OR REPLACE FUNCTION qrt.mmi_in_region(publicid VARCHAR, regioname VARCHAR)
RETURNS NUMERIC AS $$
DECLARE mmilocale NUMERIC := -1.0;
DECLARE mmi NUMERIC;
DECLARE distance NUMERIC;
DECLARE z NUMERIC := -9.0;
DECLARE mag NUMERIC := -9.0;
DECLARE slant_dist NUMERIC;
BEGIN
SELECT quake.depth, quake.magnitude INTO z, mag
FROM qrt.event as quake
WHERE quake.publicid = $1;
-- -9.0 is the unknown value for magnitude and depth
IF z != -9.0 THEN
	IF mag != -9.0 THEN
		-- Minimum depth to avoid numeric instability
		IF z < 5.0 THEN
			z := 5.0;
		END IF;
-- returns maxmmilocale in the region or -1 for invalid region or publicid
SELECT COALESCE(
	MAX( qrt.maxmmi(z, mag) - 1.18 * ln((|/(ST_Distance_Sphere(quake.geom, locality.locality_geom)/1000 * ST_Distance_Sphere(quake.geom, locality.locality_geom)/1000 + z * z)) / z) - 0.0044 * ((|/(ST_Distance_Sphere(quake.geom, locality.locality_geom)/1000) * ST_Distance_Sphere(quake.geom, locality.locality_geom)/1000 + z * z) - z)),
	 -1) INTO mmilocale   
FROM qrt.locality, qrt.event AS quake 
WHERE size IN  (0,1,2) 
AND ST_Contains((SELECT geom FROM qrt.region 
	WHERE regionname = $2), locality_geom) 
AND quake.publicid = $1;
END IF;
END IF;
RETURN round(mmilocale,2);
END;
$$ LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION qrt.maxmmi(depth numeric, magnitude numeric)
RETURNS NUMERIC AS $$
DECLARE mmi NUMERIC := -1.0;
DECLARE rupture NUMERIC;
DECLARE w NUMERIC;
BEGIN
rupture = depth;
IF abs(depth) < 100.0
THEN
    w = LEAST( 0.5 * power(10, (magnitude - 5.39)), 30.);
    rupture = GREATEST(abs(depth) - 0.5 * w * 0.85, 0.0);
END IF;

IF abs(depth) < 70.0
THEN
    mmi = 4.40 + 1.26 * magnitude -3.67 * log(rupture * rupture * rupture + 1634.691752) / 3.0 + 0.012 * depth + 0.409;
ELSE
    mmi = 3.76 + 1.48 * magnitude -3.50 * log(rupture * rupture * rupture) / 3.0 + 0.0031 * depth;
END IF;

IF mmi < 3.0
THEN
    mmi := -1.0;
END IF;
return round(mmi,2);
END;
$$ LANGUAGE plpgsql;
--
-- Create a materialized quake view.
-- Follows Chapter 12 of http://oreilly.com/catalog/9780596515201
--

create or replace view qrt.quake_unmaterialized
AS select
publicid,
originTime,
depth,
usedPhaseCount,
magnitude,
magnitudetype,
status,
type,
agency,
updateTime,
geom AS origin_geom,
qrt.closest_locality(publicid) AS locality,
ST_Contains((select geom from qrt.region where regionname = 'newzealand'), ST_Shift_Longitude(geom)) AS in_nz_region,
qrt.maxmmi(depth, magnitude) AS maxmmi,
qrt.mmi_in_region(publicid, 'newzealand') AS mmi_newzealand,
qrt.mmi_in_region(publicid, 'aucklandnorthland') AS mmi_aucklandnorthland,
qrt.mmi_in_region(publicid, 'tongagrirobayofplenty') AS mmi_tongagrirobayofplenty,
qrt.mmi_in_region(publicid, 'gisborne') AS mmi_gisborne,
qrt.mmi_in_region(publicid, 'hawkesbay') AS mmi_hawkesbay,
qrt.mmi_in_region(publicid, 'taranaki') AS mmi_taranaki,
qrt.mmi_in_region(publicid, 'wellington') AS mmi_wellington,
qrt.mmi_in_region(publicid, 'nelsonwestcoast') AS mmi_nelsonwestcoast,
qrt.mmi_in_region(publicid, 'canterbury') AS mmi_canterbury,
qrt.mmi_in_region(publicid, 'fiordland') AS mmi_fiordland,
qrt.mmi_in_region(publicid, 'otagosouthland') AS mmi_otagosouthland
FROM qrt.event;

ALTER table qrt.quake_materialized ADD column mmi_newzealand numeric;
ALTER table qrt.quake_materialized ADD column mmi_aucklandnorthland numeric;
ALTER table qrt.quake_materialized ADD column mmi_tongagrirobayofplenty numeric;
ALTER table qrt.quake_materialized ADD column mmi_gisborne numeric;
ALTER table qrt.quake_materialized ADD column mmi_hawkesbay numeric;
ALTER table qrt.quake_materialized ADD column mmi_taranaki numeric;
ALTER table qrt.quake_materialized ADD column mmi_wellington numeric;
ALTER table qrt.quake_materialized ADD column mmi_nelsonwestcoast numeric;
ALTER table qrt.quake_materialized ADD column mmi_canterbury numeric;
ALTER table qrt.quake_materialized ADD column mmi_fiordland numeric;
ALTER table qrt.quake_materialized ADD column mmi_otagosouthland numeric;
update qrt.quake_materialized SET mmi_newzealand = -1.0;
update qrt.quake_materialized SET mmi_aucklandnorthland = -1.0;
update qrt.quake_materialized SET mmi_tongagrirobayofplenty = -1.0;
update qrt.quake_materialized SET mmi_gisborne = -1.0;
update qrt.quake_materialized SET mmi_hawkesbay = -1.0;
update qrt.quake_materialized SET mmi_taranaki = -1.0;
update qrt.quake_materialized SET mmi_wellington = -1.0;
update qrt.quake_materialized SET mmi_nelsonwestcoast = -1.0;
update qrt.quake_materialized SET mmi_canterbury = -1.0;	
update qrt.quake_materialized SET mmi_fiordland = -1.0;
update qrt.quake_materialized SET mmi_otagosouthland = -1.0;	

create or replace view qrt.quakeinternal(publicid, originTime, depth, usedPhaseCount, magnitude, magnitudetype, status, type, agency, updateTime, origin_geom, locality, maxmmi)
AS select
publicid,
originTime,
depth,
usedPhaseCount,
magnitude,
magnitudetype,
status,
type,
agency,
updateTime,
origin_geom,
locality,
maxmmi,
mmi_newzealand,
mmi_aucklandnorthland,
mmi_tongagrirobayofplenty,
mmi_gisborne,
mmi_hawkesbay,
mmi_taranaki,
mmi_wellington,
mmi_nelsonwestcoast,
mmi_canterbury,
mmi_fiordland,
mmi_otagosouthland
FROM qrt.quake_materialized
WHERE
in_nz_region is true
AND status not in ('deleted', 'duplicate') and origintime > current_date - interval '1 year'
order by originTime desc;

-- The refresh is very long running
-- Only do 2012 events for now.
select qrt.quake_refresh_row(publicid) from qrt.quake_materialized where publicid like '2012%';

COMMIT;