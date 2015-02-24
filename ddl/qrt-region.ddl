DROP TABLE IF EXISTS qrt.region;

CREATE TABLE qrt.region (regionname varchar(255) PRIMARY KEY, title varchar(255), groupname varchar(255));
SELECT addgeometrycolumn('qrt', 'region', 'geom', 4326, 'POLYGON', 2);
