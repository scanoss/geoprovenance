-- Serving-schema fixture (SQLite flavour of the DuckDB export produced by
-- dba_handoff/etl_copy/load_duckdb.sh + split_duckdb.sh, group `geoprovenance`).
--
--   pkg:npm/express               -> contributors in USA (2) and Spain (1); source = expressjs/express
--   pkg:github/expressjs/express  -> one contributor in Germany
--   pkg:npm/nogeo                 -> no contributors of its own; source = expressjs/express (fallback)
--   pkg:github/scanoss/orphan     -> no contributors and no source (NO_INFO)
--   pkg:github/scanoss/blank      -> contributor whose country has an empty name (skipped)

DROP TABLE IF EXISTS component;
CREATE TABLE component (purl TEXT, vendor TEXT, component TEXT, source_purl_id TEXT, purl_id TEXT PRIMARY KEY);

DROP TABLE IF EXISTS component_version;
CREATE TABLE component_version (purl_id TEXT, version TEXT, version_name TEXT, version_semver TEXT, PRIMARY KEY (purl_id, version));

DROP TABLE IF EXISTS contribution;
CREATE TABLE contribution (purl_id TEXT, contributor_id INTEGER);

DROP TABLE IF EXISTS vendor_location;
CREATE TABLE vendor_location (vendor_id INTEGER, country_id INTEGER, location_date TEXT);

DROP TABLE IF EXISTS country;
CREATE TABLE country (id INTEGER, name TEXT, code TEXT);

DROP TABLE IF EXISTS db_version;
CREATE TABLE db_version (package_name TEXT NOT NULL, schema_version TEXT NOT NULL, db_release TEXT NOT NULL);

INSERT INTO component (purl, vendor, component, source_purl_id, purl_id) VALUES
  ('pkg:npm/express',              'expressjs', 'express', 'u-src', 'u-express'),
  ('pkg:github/expressjs/express', 'expressjs', 'express', NULL,    'u-src'),
  ('pkg:npm/nogeo',                'nogeo',     'nogeo',   'u-src', 'u-nogeo'),
  ('pkg:github/scanoss/orphan',    'scanoss',   'orphan',  NULL,    'u-orphan'),
  ('pkg:github/scanoss/blank',     'scanoss',   'blank',   NULL,    'u-blank');

INSERT INTO component_version (purl_id, version) VALUES ('u-express', '4.18.2'), ('u-src', '4.18.2');

INSERT INTO contribution (purl_id, contributor_id) VALUES
  ('u-express', 10), ('u-express', 11), ('u-express', 12),
  ('u-src', 20),
  ('u-blank', 30);

INSERT INTO vendor_location (vendor_id, country_id, location_date) VALUES
  (10, 1, '2024-01-01'), (11, 1, '2024-01-01'), (12, 2, '2024-01-01'),
  (20, 3, '2024-01-01'),
  (30, 4, '2024-01-01');

INSERT INTO country (id, name, code) VALUES (1, 'USA', 'US'), (2, 'Spain', 'ES'), (3, 'Germany', 'DE'), (4, '', '');

INSERT INTO db_version (package_name, schema_version, db_release) VALUES ('geoprovenance', '1', '26.09.21.10.00');
