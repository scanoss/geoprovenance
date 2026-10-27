-- Test data only. The schema comes from testSchemaDDL in pkg/models/test_schema.go; do not add DDL here.
-- Real rows from the serving export (2026-10-16) for the contributors in contribution.sql.
INSERT INTO vendor (id, username, type, mine_id) VALUES (351225, 'torvalds', 'User', 5);
INSERT INTO vendor (id, username, type, mine_id) VALUES (1197052, 'ulm', 'User', 5);
INSERT INTO vendor (id, username, type, mine_id) VALUES (1320559, 'lijie', 'User', 5);
INSERT INTO vendor (id, username, type, mine_id) VALUES (1405399, 'scanoss-qg', 'User', 5);
INSERT INTO vendor (id, username, type, mine_id) VALUES (1405464, 'scanossjs', 'User', 5);
INSERT INTO vendor (id, username, type, mine_id) VALUES (1405465, 'mscasso-scanoss', 'User', 5);
INSERT INTO vendor (id, username, type, mine_id) VALUES (1406024, 'perezale', 'User', 5);
INSERT INTO vendor (id, username, type, mine_id) VALUES (3236055, 'jeronimoortiz', 'User', 5);
INSERT INTO vendor (id, username, type, mine_id) VALUES (3277960, 'scanossjeronimo', 'User', 5);
-- Made up: a non-GitHub vendor (mine 4) contributing to scanoss/engine. Mining counts only
-- GitHub vendors since 0.6.0, so it must not show in engine's origin or curated counts.
INSERT INTO vendor (id, username, type, mine_id) VALUES (9999001, 'not-github', 'User', 4);
INSERT INTO contribution (purl_id, contributor_id) VALUES ('b5ccd188-eb1c-5476-bfc4-b6a44d5bacc4', 9999001);
INSERT INTO vendor_location (vendor_id, country_id, location_date) VALUES (9999001, 80, '2024-04-09');
INSERT INTO vendor_location_declared (vendor_id, declared_location, timezone_country, timezone_certainty, location_date) VALUES (9999001, 'Tokyo', 'JP', 9, '2024-04-09');
