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