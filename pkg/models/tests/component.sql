-- Test data only. The schema comes from testSchemaDDL in pkg/models/test_schema.go; do not add DDL here.
-- Serving `component`: the identity the contribution lookups key on, replacing
-- the purl name that mining's github_contributors carried.
INSERT INTO component (purl, vendor, component, purl_id, source_purl_id) VALUES ('pkg:github/scanoss/engine', NULL, NULL, 'b5ccd188-eb1c-5476-bfc4-b6a44d5bacc4', NULL);
INSERT INTO component (purl, vendor, component, purl_id, source_purl_id) VALUES ('pkg:github/scanoss/minr', NULL, NULL, 'b3cc9409-ad24-556c-8bdc-57c4393ef3d0', NULL);
INSERT INTO component (purl, vendor, component, purl_id, source_purl_id) VALUES ('pkg:github/torvalds/uemacs', NULL, NULL, 'c16f780d-fb3a-541a-8327-d9f6c965ffb9', NULL);
