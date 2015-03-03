-- This file adds the column magnitudestationcount to the event tables, quake views, and the materialized views.
-- Note that the new column is added to the end of qrt.quake_unmaterialized so that qrt.quake_materialized does 
-- not have to be completely rebuilt.


alter TABLE qrt.event add column magnitudeStationCount integer NOT NULL DEFAULT -1; 
alter TABLE qrt.eventhistory add column magnitudeStationCount integer NOT NULL DEFAULT -1;  

drop FUNCTION qrt.add_event(publicID_n TEXT, agency_n TEXT, latitude_n NUMERIC, longitude_n NUMERIC, originTime_n TIMESTAMP(6)  WITH TIME ZONE, updateTime_n TIMESTAMP(6)  WITH TIME ZONE, depth_n NUMERIC, usedPhaseCount_n INT, magnitude_n NUMERIC, magnitudeType_n TEXT, status_n TEXT, type_n TEXT );

CREATE FUNCTION qrt.add_event(publicID_n TEXT, agency_n TEXT, latitude_n NUMERIC, longitude_n NUMERIC, originTime_n TIMESTAMP(6)  WITH TIME ZONE, updateTime_n TIMESTAMP(6)  WITH TIME ZONE, depth_n NUMERIC, usedPhaseCount_n INT, magnitude_n NUMERIC, magnitudeType_n TEXT, magnitudeStationCount_n INT, status_n TEXT, type_n TEXT ) RETURNS VOID AS
$$
DECLARE
  tries INTEGER = 0;
  longitude_n numeric := longitude_n;
BEGIN
    LOOP
	IF longitude_n > 180.0 THEN
	   longitude_n = longitude_n -360.0;
        END IF;
        UPDATE qrt.event SET originTime = originTime_n, latitude = latitude_n, longitude = longitude_n, depth = depth_n, magnitude = magnitude_n, magnitudeType = magnitudeType_n, magnitudeStationCount = magnitudeStationCount_n, usedPhaseCount = usedPhaseCount_n, type = type_n, status = status_n, agency = agency_n, updateTime = updateTime_n WHERE publicID = publicID_n and updateTime_n > updateTime ;
        IF found THEN
            RETURN;
        END IF;

        BEGIN
            INSERT INTO qrt.event(publicID, originTime, latitude, longitude, depth, magnitude, magnitudeType, magnitudeStationCount, usedPhaseCount, type, status, agency, updateTime) VALUES (publicID_n, originTime_n, latitude_n, longitude_n, depth_n, magnitude_n, magnitudeType_n, magnitudeStationCount_n, usedPhaseCount_n, type_n, status_n, agency_n, updateTime_n);
            RETURN;
        EXCEPTION WHEN unique_violation THEN
            --  If we get to here the event update is probably old (updateTime_n <= updateTime).
            --  Loop once more to see if a different insert happend after the update but before
            --  our insert.
            tries = tries + 1;
            if tries > 1 THEN
               RETURN;
            END IF;
        END;
    END LOOP;
END;
$$
LANGUAGE plpgsql;

drop view qrt.quake_unmaterialized;

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

alter TABLE qrt.quake_materialized add column magnitudeStationCount integer NOT NULL DEFAULT -1;  

drop view qrt.quakeinternal;

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
