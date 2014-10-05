-- The hazard_w user here is only for running the tests.  The user for Mule needs more
-- privileges.

GRANT CONNECT ON DATABASE hazard TO hazard_w;
GRANT USAGE ON SCHEMA qrt TO hazard_w;
GRANT ALL ON ALL TABLES IN SCHEMA qrt TO hazard_w;
GRANT ALL ON ALL SEQUENCES IN SCHEMA qrt TO hazard_w;

GRANT CONNECT ON DATABASE hazard TO hazard_r;
GRANT USAGE ON SCHEMA qrt TO hazard_r;
GRANT SELECT ON ALL TABLES IN SCHEMA qrt TO hazard_r;
