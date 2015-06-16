CREATE TABLE qrt.volcanic_alert_level (
	alert_level integer PRIMARY KEY,
	hazards TEXT NOT NULL,
	activity TEXT NOT NULL
);

CREATE TABLE qrt.volcano (
	id TEXT PRIMARY KEY,
	title TEXT NOT NULL,
	location GEOGRAPHY(POINT, 4326) NOT NULL,
	region GEOGRAPHY(POLYGON, 4326),
	info_url TEXT,
	alert_level integer references qrt.volcanic_alert_level(alert_level)
);

INSERT INTO qrt.volcanic_alert_level VALUES(0, 'Volcanic environment hazards.', 'No volcanic unrest.');
INSERT INTO qrt.volcanic_alert_level VALUES(1, 'Volcanic unrest hazards.', 'Minor volcanic unrest.');
INSERT INTO qrt.volcanic_alert_level VALUES(2, 'Volcanic unrest hazards, potential for eruption hazards.', 'Moderate to heightened volcanic unrest.');
INSERT INTO qrt.volcanic_alert_level VALUES(3, 'Eruption hazards near vent. Note: ash, lava flow, and lahar (mudflow) hazards may impact areas distant from the volcano.', 'Minor volcanic eruption.');
INSERT INTO qrt.volcanic_alert_level VALUES(4, 'Eruption hazards on and near volcano. Note: ash, lava flow, and lahar (mudflow) hazards may impact areas distant from the volcano.', 'Moderate volcanic eruption.');
INSERT INTO qrt.volcanic_alert_level VALUES(5, 'Eruption hazards on and beyond volcano. Note: ash, lava flow, and lahar (mudflow) hazards may impact areas distant from the volcano.', 'Major volcanic eruption.');

INSERT INTO qrt.volcano (id, title, location, alert_level, info_url, region)
VALUES ('aucklandvolcanicfield', 'Auckland Volcanic Field', ST_GeographyFromText('POINT(174.77 -36.985)'::text), 0, 'http://info.geonet.org.nz/x/z4AO', 
	ST_GeographyFromText('POLYGON((174.4585197 -37.16746562, 174.4585197 -36.58689239, 175.510701 -36.58689239, 175.510701 -37.16746562, 174.4585197 -37.16746562))'::text));

INSERT INTO qrt.volcano (id, title, location, alert_level, info_url, region)
VALUES ('kermadecislands', 'Kermadec Islands', ST_GeographyFromText('POINT(-177.914 -29.254)'::text), 0, 'http://info.geonet.org.nz/x/24AO',
	ST_GeographyFromText('POLYGON((-179.0291841 -32.93325524, -179.0291841 -25.70303694, -175.775 -25.70303694, -175.775 -32.93325524, -179.0291841 -32.93325524))'::text));

INSERT INTO qrt.volcano (id, title, location, alert_level, info_url, region)
VALUES ('mayorisland', 'Mayor Island', ST_GeographyFromText('POINT(176.251 -37.286)'::text), 0, 'http://info.geonet.org.nz/x/74AO',
	ST_GeographyFromText('POLYGON((175.870104 -37.53170262, 175.870104 -37.04070906, 176.6399397 -37.04070906, 176.6399397 -37.53170262, 175.870104 -37.53170262))'::text));

INSERT INTO qrt.volcano (id, title, location, alert_level, info_url, region)
VALUES ('ngauruhoe', 'Ngauruhoe', ST_GeographyFromText('POINT(175.632 -39.156)'::text), 0, 'http://info.geonet.org.nz/x/94AO',
	ST_GeographyFromText('POLYGON((175.5471825 -39.21615818, 175.5471825 -39.10384673, 175.728312 -39.10384673, 175.728312 -39.21615818, 175.5471825 -39.21615818))'::text));

INSERT INTO qrt.volcano (id, title, location, alert_level, info_url, region)
VALUES ('northland', 'Northland', ST_GeographyFromText('POINT(173.63 -35.395)'::text), 0, 'http://info.geonet.org.nz/x/-oAO',
	ST_GeographyFromText('POLYGON((173.2122957 -36.25470988, 173.2122957 -34.88581459, 175.0724475 -34.88581459, 175.0724475 -36.25470988, 173.2122957 -36.25470988))'::text));

INSERT INTO qrt.volcano (id, title, location, alert_level, info_url, region)
VALUES ('okataina', 'Okataina', ST_GeographyFromText('POINT(176.501 -38.119)'::text), 0, 'http://info.geonet.org.nz/x/B4EO',
	ST_GeographyFromText('POLYGON((176.3158211 -38.33990913, 176.3158211 -37.94704823, 176.8111052 -37.94704823, 176.8111052 -38.33990913, 176.3158211 -38.33990913))'::text));

INSERT INTO qrt.volcano (id, title, location, alert_level, info_url, region)
VALUES ('rotorua', 'Rotorua', ST_GeographyFromText('POINT(176.281 -38.093)'::text), 0, 'http://info.geonet.org.nz/x/FYEO',
	ST_GeographyFromText('POLYGON((176.11533 -38.20135287, 176.11533 -37.97620536, 176.4250812 -37.97620536, 176.4250812 -38.20135287, 176.11533 -38.20135287))'::text));

INSERT INTO qrt.volcano (id, title, location, alert_level, info_url, region)
VALUES ('ruapehu', 'Ruapehu', ST_GeographyFromText('POINT(175.563 -39.281)'::text),1, 'http://info.geonet.org.nz/x/GYEO',
	ST_GeographyFromText('POLYGON((175.3707552 -39.481325, 175.3707552 -39.09468564, 175.7744228 -39.09468564, 175.7744228 -39.481325, 175.3707552 -39.481325))'::text));

INSERT INTO qrt.volcano (id, title, location, alert_level, info_url, region)
VALUES ('taupo', 'Taupo', ST_GeographyFromText('POINT(175.896 -38.784)'::text), 0, 'http://info.geonet.org.nz/x/bIEO',
	ST_GeographyFromText('POLYGON((175.564837 -39.08056833, 175.564837 -38.58664502, 176.2482749 -38.58664502, 176.2482749 -39.08056833, 175.564837 -39.08056833))'::text));

INSERT INTO qrt.volcano (id, title, location, alert_level, info_url, region)
VALUES ('tongariro', 'Tongariro', ST_GeographyFromText('POINT(175.641727 -39.133318)'::text),1, 'http://info.geonet.org.nz/x/dIEO',
	ST_GeographyFromText('POLYGON((175.5689901 -39.17961512, 175.5689901 -39.06727363, 175.7499926 -39.06727363, 175.7499926 -39.17961512, 175.5689901 -39.17961512))'::text));

INSERT INTO qrt.volcano (id, title, location, alert_level, info_url, region)
VALUES ('taranakiegmont', 'Taranaki/Egmont', ST_GeographyFromText('POINT(174.061 -39.298)'::text), 0, 'http://info.geonet.org.nz/x/W4EO',
	ST_GeographyFromText('POLYGON((173.6983776 -39.67527512, 173.6983776 -38.94831596, 174.4993628 -38.94831596, 174.4993628 -39.67527512, 173.6983776 -39.67527512))'::text));

INSERT INTO qrt.volcano (id, title, location, alert_level, info_url, region)
VALUES ('whiteisland', 'White Island', ST_GeographyFromText('POINT(177.183 -37.521)'::text),1, 'http://info.geonet.org.nz/x/ioEO',
	ST_GeographyFromText('POLYGON((176.6867564 -38.00383212, 176.6867564 -37.33926271, 177.400852 -37.33926271, 177.400852 -38.00383212, 176.6867564 -38.00383212))'::text));
