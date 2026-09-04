-- Test data only. The schema comes from testSchemaDDL in pkg/models/test_schema.go; do not add DDL here.
-- Serving `component_version`: one release each, so the resolver can pin a version.
INSERT INTO component_version (purl_id, version, version_name, version_semver) VALUES ('b5ccd188-eb1c-5476-bfc4-b6a44d5bacc4', '1.0.0', '1.0.0', '1.0.0');
INSERT INTO component_version (purl_id, version, version_name, version_semver) VALUES ('b3cc9409-ad24-556c-8bdc-57c4393ef3d0', '1.0.0', '1.0.0', '1.0.0');
INSERT INTO component_version (purl_id, version, version_name, version_semver) VALUES ('c16f780d-fb3a-541a-8327-d9f6c965ffb9', '1.0.0', '1.0.0', '1.0.0');
