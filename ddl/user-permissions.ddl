-- qrt and impact are kept as separate schemas with separate write
-- users for security.

GRANT CONNECT ON DATABASE hazard TO hazard_w;
GRANT USAGE ON SCHEMA qrt TO hazard_w;
GRANT ALL ON ALL TABLES IN SCHEMA qrt TO hazard_w;
GRANT ALL ON ALL SEQUENCES IN SCHEMA qrt TO hazard_w;

GRANT CONNECT ON DATABASE hazard TO hazard_r;
GRANT USAGE ON SCHEMA qrt TO hazard_r;
GRANT SELECT ON ALL TABLES IN SCHEMA qrt TO hazard_r;
GRANT USAGE ON SCHEMA impact TO hazard_r;
GRANT SELECT ON ALL TABLES IN SCHEMA impact TO hazard_r;

GRANT CONNECT ON DATABASE hazard TO impact_w;
GRANT USAGE ON SCHEMA impact TO impact_w;
GRANT ALL ON ALL TABLES IN SCHEMA impact TO impact_w;
GRANT ALL ON ALL SEQUENCES IN SCHEMA impact TO impact_w;
