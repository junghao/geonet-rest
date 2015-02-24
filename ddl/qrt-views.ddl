-- For everthing to behave well the default timezone needs to be set e.g.,
-- in /var/lib/pgsql/data/postgresql.conf
--
-- timezone = UTC
-- timezone_abbreviations = 'Default'
--
-- This DDL builds on the schema created by seiscomp-wfs-client.
-- It adds regionalised views to the data.

DROP VIEW IF EXISTS qrt.quake;
DROP TABLE IF EXISTS qrt.quake_materialized CASCADE;
DROP VIEW IF EXISTS qrt.quake_unmaterialized;
DROP FUNCTION IF EXISTS qrt.closest_locality(publicid VARCHAR);
DROP FUNCTION IF EXISTS qrt.compass_azimuth(DOUBLE PRECISION);
DROP FUNCTION IF EXISTS qrt.mmi_to_intensity(DOUBLE PRECISION);
DROP FUNCTION IF EXISTS qrt.quake_quality(TEXT, INTEGER, INTEGER);

DELETE FROM geometry_columns where f_table_schema='qrt' AND f_table_name='quake';
DELETE FROM geometry_columns where f_table_schema='qrt' AND f_table_name='quakeinternal';
DELETE FROM qrt.gt_pk_metadata_table where table_name='quake';


-- An optimisation would be to add intensity strings for each region to the materialised view.
-- This would get quite complicated though.
--
-- Converts mmi to an intensity string
create function qrt.mmi_to_intensity(mmi DOUBLE PRECISION)
  returns text as $$
BEGIN
IF $1 >= 7
THEN
   return 'severe';
ELSEIF $1 >= 6
THEN
   return 'strong';
ELSEIF $1 >= 5
THEN
    return 'moderate';
ELSEIF $1 >= 4
THEN
   return 'light';
ELSEIF $1 >= 3
THEN
   return 'weak';
END IF;
return 'unnoticeable';
END;
$$ LANGUAGE plpgsql;

-- Converts the intensity string to min MMI.  This allows the intensity to be used as a query value instead of MMI.
create function qrt.intensity_to_mmi(intensity TEXT)
  returns DOUBLE PRECISION as $$
BEGIN
IF $1 = 'weak'
THEN
  return 3.0;
ELSEIF $1 = 'light'
THEN
  return 4.0;
ELSEIF $1 = 'moderate'
THEN
  return 5.0;
ELSEIF $1 = 'strong'
THEN
  return 6.0;
ELSEIF $1 = 'severe'
THEN
  return 7.0;
END IF;
return -9.0;
END;
$$ LANGUAGE plpgsql;

-- Calculates the quake quality.  Could be added as a column in the materialised view.
create function qrt.quake_quality(status TEXT, phase_count INTEGER, mag_count INTEGER)
RETURNS TEXT as $$
BEGIN
IF $1 = 'reviewed'
THEN
  return 'best';
ELSEIF $1 = 'deleted'
THEN
  return 'deleted';
ELSEIF ($2 >= 20) AND ($3 >= 10)
THEN
  return 'good';
ELSEIF ($2 >= 20) AND ($3 = -1)
THEN
  return 'good';
ELSEIF (($2 > 0) AND ($2 < 20)) OR (($3 > 0) AND ($3 < 10))
THEN
  return 'caution';
END IF;
return 'unknown';
END;
$$ LANGUAGE plpgsql;

-- N 0 or 360, NE 45, E 90, SE, 135, S 180, SW 225, W 270, NW 315
create function qrt.compass_azimuth(azimuth DOUBLE PRECISION)
RETURNS TEXT as $$
BEGIN
IF $1 >= 337.5 AND  $1 <= 360
THEN 
     return 'north';
ELSIF $1 >= 0 AND  $1 <= 22.5
THEN
	return 'north';
ELSIF $1 > 22.5 AND  $1 < 67.5
THEN
	return 'north-east';
ELSIF $1 >= 67.5 AND  $1 <= 112.5
THEN
	return 'east';
ELSIF $1 > 112.5 AND  $1 < 157.5
THEN
	return 'south-east';
ELSIF $1 >= 157.5 AND  $1 <= 202.5
THEN
	return 'south';
ELSIF $1 > 202.5 AND  $1 < 247.5
THEN
	return 'south-west';
ELSIF $1 >= 247.5 AND  $1 <= 292.5
THEN
	return 'west';
ELSIF $1 > 292.5 AND  $1 < 337.5
THEN
	return 'north-west';
END IF;
return 'north';
END; 
$$ LANGUAGE plpgsql;

create function qrt.closest_locality(publicid VARCHAR)
RETURNS TEXT AS $$
DECLARE locale TEXT;
DECLARE distance INTEGER;
DECLARE bearing TEXT;
BEGIN
SELECT (round(ST_Distance_Sphere(quake.geom, locality.locality_geom)/5000) * 5), qrt.compass_azimuth(ST_Azimuth(locality.locality_geom, ST_Shift_Longitude(quake.geom))/(2*pi())*360), locality.name INTO distance, bearing, locale
FROM qrt.event As quake, qrt.locality As locality
WHERE quake.publicid = $1 AND ST_DWithin(ST_Shift_Longitude(quake.geom), locality.locality_geom, 3) AND locality.size in (0,1,2)
order by round(ST_Distance_Sphere(quake.geom, locality.locality_geom)/1000)
limit 1;
IF NOT FOUND
THEN
	SELECT (round(ST_Distance_Sphere(quake.geom, locality.locality_geom)/5000) * 5), qrt.compass_azimuth(ST_Azimuth(locality.locality_geom, ST_Shift_Longitude(quake.geom))/(2*pi())*360), locality.name INTO distance, bearing, locale
	FROM qrt.event As quake, qrt.locality As locality
	WHERE quake.publicid = $1 AND locality.size IN (0,1)
	order by round(ST_Distance_Sphere(quake.geom, locality.locality_geom)/1000)
	limit 1;
END IF;
IF distance = 0
THEN
	locale := 'Within 5 km of ' || locale;
ELSE
	locale := CAST(distance AS TEXT) || ' km ' || bearing || ' of ' || locale;
END IF;
RETURN locale;
END;
$$ LANGUAGE plpgsql;

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

--
-- Additional functions to use on the ESB to enrich quake messages
-- Prevents race condition caused by trying to read these values from
-- qrt.quake_materialized.  These replicate other functions but with 
-- different signature.  TODO consolidate logic.
--
create or REPLACE function qrt.closest_locality_func(longitude NUMERIC, latitude NUMERIC)
RETURNS TEXT AS $$
DECLARE locale TEXT;
DECLARE distance INTEGER;
DECLARE bearing TEXT;
DECLARE quake_location geometry;
BEGIN
quake_location := ST_SetSRID(ST_MakePoint(longitude, latitude), 4326);
SELECT (round(ST_Distance_Sphere(quake_location, locality.locality_geom)/5000) * 5), qrt.compass_azimuth(ST_Azimuth(locality.locality_geom, ST_Shift_Longitude(quake_location))/(2*pi())*360), locality.name INTO distance, bearing, locale
FROM qrt.locality As locality
WHERE ST_DWithin(ST_Shift_Longitude(quake_location), locality.locality_geom, 3) AND locality.size in (0,1,2)
order by round(ST_Distance_Sphere(quake_location, locality.locality_geom)/1000)
limit 1;
IF NOT FOUND
THEN
	SELECT (round(ST_Distance_Sphere(quake_location, locality.locality_geom)/5000) * 5), qrt.compass_azimuth(ST_Azimuth(locality.locality_geom, ST_Shift_Longitude(quake_location))/(2*pi())*360), locality.name INTO distance, bearing, locale
	FROM qrt.locality As locality
	WHERE locality.size IN (0,1)
	order by round(ST_Distance_Sphere(quake_location, locality.locality_geom)/1000)
	limit 1;
END IF;
IF distance = 0
THEN
	locale := 'Within 5 km of ' || locale;
ELSE
	locale := CAST(distance AS TEXT) || ' km ' || bearing || ' of ' || locale;
END IF;
RETURN locale;
END;
$$ LANGUAGE plpgsql;

-- Algorithm described in https://github.com/GeoNet/quakes/issues/159
-- CREATE OR REPLACE FUNCTION qrt.mmi_in_region(publicid VARCHAR, regioname VARCHAR)
CREATE OR REPLACE FUNCTION qrt.mmi_in_nz_func(longitude NUMERIC, latitude NUMERIC, depth NUMERIC, magnitude NUMERIC)
RETURNS NUMERIC AS $$
DECLARE mmilocale NUMERIC := -1.0;
DECLARE mmi NUMERIC;
DECLARE distance NUMERIC;
DECLARE slant_dist NUMERIC;
DECLARE z NUMERIC := depth;
DECLARE mag NUMERIC := magnitude;
DECLARE quake_location geometry;
BEGIN
quake_location := ST_SetSRID(ST_MakePoint(longitude, latitude), 4326);
-- -9.0 is the unknown value for magnitude and depth
IF z != -9.0 THEN
	IF mag != -9.0 THEN
		-- Minimum depth to avoid numeric instability
		IF z < 5.0 THEN
			z := 5.0;
		END IF;
-- returns maxmmilocale in the region or -1 for invalid region or publicid
SELECT COALESCE(
	MAX( qrt.maxmmi(z, mag) - 1.18 * ln((|/(ST_Distance_Sphere(quake_location, locality.locality_geom)/1000 * ST_Distance_Sphere(quake_location, locality.locality_geom)/1000 + z * z)) / z) - 0.0044 * ((|/(ST_Distance_Sphere(quake_location, locality.locality_geom)/1000) * ST_Distance_Sphere(quake_location, locality.locality_geom)/1000 + z * z) - z)),
	 -1) INTO mmilocale   
FROM qrt.locality 
WHERE size IN  (0,1,2) 
AND ST_Contains((SELECT geom FROM qrt.region 
	WHERE regionname = 'newzealand'), locality_geom);
END IF;
END IF;
RETURN round(mmilocale,2);
END;
$$ LANGUAGE plpgsql;
--
-- End of ESB enrich functions.
--
-- Add new columns to the end of the view so that the materialized table does not have to be completely
-- rebuilt.  See also add-station-mags.ddl
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
qrt.mmi_in_region(publicid, 'otagosouthland') AS mmi_otagosouthland,
magnitudestationcount
FROM qrt.event;

create table qrt.quake_materialized AS
SELECT *
FROM qrt.quake_unmaterialized;

CREATE INDEX quake_materialized_oritime_idx ON qrt.quake_materialized (originTime);
CREATE INDEX quake_materialized_publicid_idx ON qrt.quake_materialized (publicid);
--
-- To force refresh all rows:
-- select qrt.quake_refresh_row(publicid) from qrt.quake_materialized;
--
create or replace function qrt.quake_refresh_row(publicid VARCHAR) returns void
security definer
language 'plpgsql' as $$
BEGIN
DELETE FROM qrt.quake_materialized qm 
WHERE qm.publicid = quake_refresh_row.publicid;
INSERT INTO qrt.quake_materialized 
SELECT * 
FROM qrt.quake_unmaterialized qu
WHERE qu.publicid = quake_refresh_row.publicid;
end
$$;

create or replace function  qrt.event_ut() returns trigger
security definer language 'plpgsql' as $$ 
begin 
if old.publicid = new.publicid then 
perform qrt.quake_refresh_row(new.publicid); 
else 
perform qrt.quake_refresh_row(old.publicid); 
perform qrt.quake_refresh_row(new.publicid); 
end if; 
return null; 
end 
$$; 

create or replace function  qrt.event_dt() returns trigger
security definer language 'plpgsql' as $$ 
begin 
perform qrt.quake_refresh_row(old.publicid); 
return null; 
end 
$$; 

create or replace function  qrt.event_it() returns trigger
security definer language 'plpgsql' as $$ 
begin 
perform qrt.quake_refresh_row(new.publicid); 
return null; 
end 
$$; 

--
-- TODO what happens if there is a crash before a trigger runs?
-- Do we need a periodic sweeper to run 
-- select qrt.quake_refresh_row(publicid) from qrt.event 
-- with a join for missing publicids ?
--
DROP TRIGGER IF EXISTS event_quake_ut on qrt.event;
DROP TRIGGER IF EXISTS event_quake_dt on qrt.event;
DROP TRIGGER IF EXISTS event_quake_it on qrt.event;

create trigger event_quake_ut after update on qrt.event for each row execute procedure qrt.event_ut();  

create trigger event_quake_dt after delete on qrt.event for each row execute procedure qrt.event_dt();  

create trigger event_quake_it after insert on qrt.event for each row execute procedure qrt.event_it();  

--
-- This view is designed to be published through Geoserver.
-- Only add things to it that should go into public webservices
--
create view qrt.quake(publicid, originTime, longitude, latitude, depth, magnitude, magnitudetype, magnitudestationcount, status, phases, type, agency, updateTime, origin_geom)
AS select
publicid,
originTime,
ST_X(origin_geom) as longtiude, 
ST_Y(origin_geom) as latitude,
depth,
magnitude,
magnitudetype,
magnitudestationcount,
status,
usedphasecount as phases,
type,
agency,
updateTime,
origin_geom
FROM qrt.quake_materialized 
WHERE
in_nz_region is true
AND status not in ('deleted', 'duplicate') 
order by originTime desc;

INSERT INTO geometry_columns(f_table_catalog, f_table_schema, f_table_name, f_geometry_column, coord_dimension, srid, "type") VALUES ('', 'qrt', 'quake', 'origin_geom', 2, 4326, 'POINT');
INSERT INTO qrt.gt_pk_metadata_table(table_schema, table_name, pk_column, pk_column_idx, pk_policy, pk_sequence) VALUES ('qrt', 'quake', 'publicid', null, null, null);

--
-- This view should not be published via Geoserver.
-- It is for use with the SpringDao.
--
create or replace view qrt.quakeinternal(publicid, originTime, depth, usedPhaseCount, magnitude, magnitudetype, magnitudestationcount, status, type, agency, updateTime, origin_geom, locality, maxmmi)
AS select
publicid,
originTime,
depth,
usedPhaseCount,
magnitude,
magnitudetype,
magnitudestationcount,
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

-- This view includs deleted quakes.
-- The columns that call functions can be added to the materialised view instead if this is 
-- Better for performance reasons.
-- If the number of quakes selected is 1500 or less the performance difference (~40ms) is not currently
-- worth the time it would take to change the materialised view.
create or replace view qrt.quakeinternal_v2
AS select
publicid,
originTime,
depth,
magnitude,
qrt.quake_quality(status, usedphasecount, magnitudestationcount) as quality,
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
mmi_otagosouthland,
qrt.mmi_to_intensity(maxmmi) as intensity,
qrt.mmi_to_intensity(mmi_newzealand) as intensity_newzealand,
qrt.mmi_to_intensity(mmi_aucklandnorthland) as intensity_aucklandnorthland,
qrt.mmi_to_intensity(mmi_tongagrirobayofplenty) as intensity_tongagrirobayofplenty,
qrt.mmi_to_intensity(mmi_gisborne) as intensity_gisborne,
qrt.mmi_to_intensity(mmi_hawkesbay) as intensity_hawkesbay,
qrt.mmi_to_intensity(mmi_taranaki) as intensity_taranaki,
qrt.mmi_to_intensity(mmi_wellington) as intensity_wellington,
qrt.mmi_to_intensity(mmi_nelsonwestcoast) as intensity_nelsonwestcoast,
qrt.mmi_to_intensity(mmi_canterbury) as intensity_canterbury,
qrt.mmi_to_intensity(mmi_fiordland) as intensity_fiordland,
qrt.mmi_to_intensity(mmi_otagosouthland) as intensity_otagosouthland
FROM qrt.quake_materialized
WHERE
in_nz_region is true
AND status != 'duplicate' 
AND origintime > current_date - interval '1 year'
order by originTime desc;

INSERT INTO geometry_columns(f_table_catalog, f_table_schema, f_table_name, f_geometry_column, coord_dimension, srid, "type") VALUES ('', 'qrt', 'quakeinternal', 'origin_geom', 2, 4326, 'POINT');
INSERT INTO qrt.gt_pk_metadata_table(table_schema, table_name, pk_column, pk_column_idx, pk_policy, pk_sequence) VALUES ('qrt', 'quakeinternal', 'publicid', null, null, null);

CREATE INDEX eventhistory_publicid__idx ON qrt.eventhistory (publicid);
