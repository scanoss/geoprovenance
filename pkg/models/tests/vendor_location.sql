-- Test data only. The schema comes from testSchemaDDL in pkg/models/test_schema.go; do not add DDL here.
-- Serving `vendor_location`: the curated country per contributor. Mining held it
-- as a nested array literal in curated_countries_ids, alongside a declared
-- location and a timezone-inferred country that serving does not carry.
INSERT INTO vendor_location (vendor_id, country_id, location_date) VALUES (3555999, 7, '2024-10-27');
INSERT INTO vendor_location (vendor_id, country_id, location_date) VALUES (1405397, 80, '2024-04-09');
INSERT INTO vendor_location (vendor_id, country_id, location_date) VALUES (1406026, 36, '2024-04-09');
INSERT INTO vendor_location (vendor_id, country_id, location_date) VALUES (1406025, 64, '2024-04-09');
INSERT INTO vendor_location (vendor_id, country_id, location_date) VALUES (283816, 163, '2024-04-09');
INSERT INTO vendor_location (vendor_id, country_id, location_date) VALUES (1405399, 7, '2024-04-09');
INSERT INTO vendor_location (vendor_id, country_id, location_date) VALUES (1405400, 163, '2024-04-09');
INSERT INTO vendor_location (vendor_id, country_id, location_date) VALUES (1405465, 7, '2024-04-09');
INSERT INTO vendor_location (vendor_id, country_id, location_date) VALUES (1406024, 7, '2024-04-09');
INSERT INTO vendor_location (vendor_id, country_id, location_date) VALUES (377779, 64, '2024-04-09');
INSERT INTO vendor_location (vendor_id, country_id, location_date) VALUES (3277960, 7, '2024-09-19');
INSERT INTO vendor_location (vendor_id, country_id, location_date) VALUES (1406028, 36, '2024-04-09');
INSERT INTO vendor_location (vendor_id, country_id, location_date) VALUES (1405398, 210, '2024-04-09');
INSERT INTO vendor_location (vendor_id, country_id, location_date) VALUES (1405490, 168, '2024-04-09');
