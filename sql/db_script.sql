BEGIN;
CREATE USER koko WITH LOGIN PASSWORD 'koko';
GRANT USAGE ON SCHEMA public TO koko;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO koko;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO koko;
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO koko;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE ON SEQUENCES TO koko;

DROP TABLE IF EXISTS "canary";

CREATE TABLE "canary"
(
    id bigint primary key,
    ts timestamp
);

INSERT INTO canary values(1, CURRENT_TIMESTAMP);

DROP TABLE IF EXISTS "replication_canary";

CREATE TABLE "replication_canary"
(
    id bigint primary key,
    ts timestamp
);

INSERT INTO replication_canary values(1, CURRENT_TIMESTAMP);

COMMIT;