-- This is a copy of the table from https://github.com/GeoNet/soh

DROP TABLE  IF EXISTS  qrt.soh;

CREATE TABLE qrt.soh (serverID varchar(255) NOT NULL unique,  timeReceived timestamp(6)  WITH TIME ZONE, PRIMARY KEY (serverID) );

CREATE FUNCTION qrt.add_soh(serverID_n TEXT, timeReceived_n TIMESTAMP(6)  WITH TIME ZONE) RETURNS VOID AS
$$
DECLARE
  tries INTEGER = 0;
BEGIN
    LOOP
        UPDATE qrt.soh SET timeReceived = timeReceived_n WHERE serverID = serverID_n and timeReceived_n > timeReceived ;
        IF found THEN
            RETURN;
        END IF;

        BEGIN
            INSERT INTO qrt.soh(serverID, timeReceived) VALUES (serverID_n, timeReceived_n);
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