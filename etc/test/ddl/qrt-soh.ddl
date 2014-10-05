-- This is a copy of the table from https://github.com/GeoNet/soh

DROP TABLE  IF EXISTS  qrt.soh;

CREATE TABLE qrt.soh (serverID varchar(255) NOT NULL unique,  timeReceived timestamp(6)  WITH TIME ZONE, PRIMARY KEY (serverID) );
