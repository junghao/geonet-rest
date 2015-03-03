-- For everthing to behave well the default timezone needs to be set e.g.,
-- in /var/lib/pgsql/data/postgresql.conf
--
-- timezone = UTC
-- timezone_abbreviations = 'Default'

DROP SCHEMA IF EXISTS qrt CASCADE;

DELETE from geometry_columns where f_table_schema='qrt';

CREATE SCHEMA qrt;

-- numeric without any precision or scale creates a column in which numeric
-- values of any precision and scale can be stored, up to the implementation
-- limit on precision.
-- http://www.postgresql.org/docs/8.3/static/datatype-numeric.html

CREATE TABLE qrt.event (
publicID varchar(255) PRIMARY KEY,
originTime timestamp(6) WITH TIME ZONE NOT NULL,
latitude numeric NOT NULL,
longitude numeric NOT NULL,
depth numeric NOT NULL,
magnitude numeric NOT NULL,
magnitudeType varchar(255) NOT NULL,
magnitudeStationCount integer NOT NULL DEFAULT -1,
usedPhaseCount integer NOT NULL DEFAULT -1,
type varchar(255) NOT NULL DEFAULT 'earthquake',
status varchar(255) NOT NULL DEFAULT 'automatic',
agency varchar(255) NOT NULL,
updateTime timestamp(6) WITH TIME ZONE NOT NULL,
epsgCode int4 DEFAULT 4326 NOT NULL);

CREATE TABLE qrt.eventhistory (
eventID  SERIAL NOT NULL,
publicID varchar(255) NOT NULL,
originTime timestamp(6) WITH TIME ZONE NOT NULL,
latitude numeric NOT NULL,
longitude numeric NOT NULL,
depth numeric NOT NULL,
magnitude numeric NOT NULL,
magnitudeType varchar(255) NOT NULL,
magnitudeStationCount integer NOT NULL DEFAULT -1,
usedPhaseCount integer NOT NULL DEFAULT -1,
type varchar(255) NOT NULL DEFAULT 'earthquake',
status varchar(255) NOT NULL DEFAULT 'automatic',
updateTime timestamp(6) WITH TIME ZONE NOT NULL,
PRIMARY KEY (eventID));

CREATE TABLE qrt.gt_pk_metadata_table (table_schema varchar(255), table_name varchar(255), pk_column varchar(255), pk_column_idx integer, pk_policy varchar(255), pk_sequence varchar(255));

SELECT addgeometrycolumn('qrt', 'event', 'geom', 4326, 'POINT', 2);

CREATE FUNCTION qrt.update_origin_geom() RETURNS  TRIGGER AS E' BEGIN NEW.geom = st_transform(st_setsrid(st_makepoint(NEW.longitude, NEW.latitude),New.epsgCode),4326); RETURN NEW;  END; ' LANGUAGE plpgsql;

CREATE TRIGGER origin_geom_trigger BEFORE INSERT OR UPDATE ON qrt.event
  FOR EACH ROW EXECUTE PROCEDURE qrt.update_origin_geom();

CREATE INDEX event_oritime_idx ON qrt.event (originTime);

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
