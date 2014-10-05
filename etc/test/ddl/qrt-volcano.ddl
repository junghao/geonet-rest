CREATE TABLE qrt.volcano
(
	volcano_id varchar(255) PRIMARY KEY,
	title varchar(255) NOT NULL,
	alert_level_updated_time timestamp(6) WITH TIME ZONE NOT NULL,
	aviation_code_updated_time timestamp(6) WITH TIME ZONE NOT NULL,
  aviation_code varchar(255) NOT NULL,
  aviation_status varchar(255) NOT NULL,
  alert_level integer NOT NULL,
  alert_activity varchar(255) NOT NULL,
  alert_hazards  varchar(255) NOT NULL,
	info_url varchar(255) NOT NULL
);

SELECT addgeometrycolumn('qrt', 'volcano', 'point', 4326, 'POINT', 2);

INSERT INTO qrt.volcano (volcano_id, title, alert_level_updated_time, aviation_code_updated_time, aviation_code, aviation_status, alert_level, alert_activity, alert_hazards, info_url, point)
VALUES ('aucklandvolcanicfield', 'Auckland Volcanic Field', now(), now(), 'GREEN', 'Volcano is in normal, non-eruptive state.', 0,
        'No volcanic unrest.', 'Volcanic environment hazards.', 'http://info.geonet.org.nz/x/z4AO', st_geomfromtext('POINT(174.77 -36.985)'::text, 4326));

INSERT INTO qrt.volcano (volcano_id, title, alert_level_updated_time, aviation_code_updated_time, aviation_code, aviation_status, alert_level, alert_activity, alert_hazards, info_url, point)
VALUES ('kermadecislands', 'Kermadec Islands', now(), now(), 'GREEN', 'Volcano is in normal, non-eruptive state.', 0,
        'No volcanic unrest.', 'Volcanic environment hazards.', 'http://info.geonet.org.nz/x/24AO', st_geomfromtext('POINT(-177.914 -29.254)'::text, 4326));

INSERT INTO qrt.volcano (volcano_id, title, alert_level_updated_time, aviation_code_updated_time, aviation_code, aviation_status, alert_level, alert_activity, alert_hazards, info_url, point)
VALUES ('mayorisland', 'Mayor Island', now(), now(), 'GREEN', 'Volcano is in normal, non-eruptive state.', 0,
        'No volcanic unrest.', 'Volcanic environment hazards.', 'http://info.geonet.org.nz/x/74AO', st_geomfromtext('POINT(176.251 -37.286)'::text, 4326));

INSERT INTO qrt.volcano (volcano_id, title, alert_level_updated_time, aviation_code_updated_time, aviation_code, aviation_status, alert_level, alert_activity, alert_hazards, info_url, point)
VALUES ('ngauruhoe', 'Ngauruhoe', now(), now(), 'GREEN', 'Volcano is in normal, non-eruptive state.', 0,
        'No volcanic unrest.', 'Volcanic environment hazards.', 'http://info.geonet.org.nz/x/94AO', st_geomfromtext('POINT(175.632 -39.156)'::text, 4326));

INSERT INTO qrt.volcano (volcano_id, title, alert_level_updated_time, aviation_code_updated_time, aviation_code, aviation_status, alert_level, alert_activity, alert_hazards, info_url, point)
VALUES ('northland', 'Northland', now(), now(), 'GREEN', 'Volcano is in normal, non-eruptive state.', 0,
        'No volcanic unrest.', 'Volcanic environment hazards.', 'http://info.geonet.org.nz/x/-oAO', st_geomfromtext('POINT(173.63 -35.395)'::text, 4326));

INSERT INTO qrt.volcano (volcano_id, title, alert_level_updated_time, aviation_code_updated_time, aviation_code, aviation_status, alert_level, alert_activity, alert_hazards, info_url, point)
VALUES ('okataina', 'Okataina', now(), now(), 'GREEN', 'Volcano is in normal, non-eruptive state.', 0,
        'No volcanic unrest.', 'Volcanic environment hazards.', 'http://info.geonet.org.nz/x/B4EO', st_geomfromtext('POINT(176.501 -38.119)'::text, 4326));

INSERT INTO qrt.volcano (volcano_id, title, alert_level_updated_time, aviation_code_updated_time, aviation_code, aviation_status, alert_level, alert_activity, alert_hazards, info_url, point)
VALUES ('rotorua', 'Rotorua', now(), now(), 'GREEN', 'Volcano is in normal, non-eruptive state.', 0,
        'No volcanic unrest.', 'Volcanic environment hazards.', 'http://info.geonet.org.nz/x/FYEO', st_geomfromtext('POINT(176.281 -38.093)'::text, 4326));

INSERT INTO qrt.volcano (volcano_id, title, alert_level_updated_time, aviation_code_updated_time, aviation_code, aviation_status, alert_level, alert_activity, alert_hazards, info_url, point)
VALUES ('ruapehu', 'Ruapehu', now(), now(), 'GREEN', 'Volcano is in normal, non-eruptive state.', 1,
        'Minor volcanic unrest.', 'Volcanic unrest hazards.', 'http://info.geonet.org.nz/x/GYEO', st_geomfromtext('POINT(175.563 -39.281)'::text, 4326));

INSERT INTO qrt.volcano (volcano_id, title, alert_level_updated_time, aviation_code_updated_time, aviation_code, aviation_status, alert_level, alert_activity, alert_hazards, info_url, point)
VALUES ('taupo', 'Taupo', now(), now(), 'GREEN', 'Volcano is in normal, non-eruptive state.', 0,
        'No volcanic unrest.', 'Volcanic environment hazards.', 'http://info.geonet.org.nz/x/bIEO', st_geomfromtext('POINT(175.896 -38.784)'::text, 4326));

INSERT INTO qrt.volcano (volcano_id, title, alert_level_updated_time, aviation_code_updated_time, aviation_code, aviation_status, alert_level, alert_activity, alert_hazards, info_url, point)
VALUES ('tongariro', 'Tongariro', now(), now(), 'GREEN', 'Volcano is in normal, non-eruptive state.', 1,
        'Minor volcanic unrest.', 'Volcanic unrest hazards.', 'http://info.geonet.org.nz/x/dIEO', st_geomfromtext('POINT(175.641727 -39.133318)'::text, 4326));

INSERT INTO qrt.volcano (volcano_id, title, alert_level_updated_time, aviation_code_updated_time, aviation_code, aviation_status, alert_level, alert_activity, alert_hazards, info_url, point)
VALUES ('taranakiegmont', 'Taranaki/Egmont', now(), now(), 'GREEN', 'Volcano is in normal, non-eruptive state.', 0,
        'No volcanic unrest.', 'Volcanic environment hazards.', 'http://info.geonet.org.nz/x/W4EO', st_geomfromtext('POINT(174.061 -39.298)'::text, 4326));

INSERT INTO qrt.volcano (volcano_id, title, alert_level_updated_time, aviation_code_updated_time, aviation_code, aviation_status, alert_level, alert_activity, alert_hazards, info_url, point)
VALUES ('whiteisland', 'White Island', now(), now(), 'GREEN', 'Volcano is in normal, non-eruptive state.', 1,
        'Minor volcanic unrest.', 'Volcanic unrest hazards.', 'http://info.geonet.org.nz/x/ioEO', st_geomfromtext('POINT(177.183 -37.521)'::text, 4326));



