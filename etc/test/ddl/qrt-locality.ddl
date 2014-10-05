DROP TABLE IF EXISTS qrt.locality;
DELETE FROM geometry_columns where f_table_schema='qrt' AND f_table_name='locality';

CREATE TABLE qrt.locality (
localityID SERIAL PRIMARY KEY,
name varchar(255) NOT NULL,
size INTEGER NOT NULL
);
SELECT addgeometrycolumn('qrt', 'locality', 'locality_geom', 4326, 'POINT', 2);
ALTER TABLE qrt.locality ALTER locality_geom SET NOT NULL;
