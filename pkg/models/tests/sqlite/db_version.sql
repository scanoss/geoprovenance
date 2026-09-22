DROP TABLE IF EXISTS db_version;
CREATE TABLE db_version (package_name TEXT NOT NULL, schema_version TEXT NOT NULL, created_at TEXT NOT NULL, db_release TEXT NOT NULL);
INSERT INTO db_version VALUES ('provenance', '1.0.0', '2026-09-22T19:00Z', '26.09');
