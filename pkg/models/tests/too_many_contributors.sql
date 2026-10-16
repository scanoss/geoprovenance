-- Test data only. The schema comes from testSchemaDDL in pkg/models/test_schema.go; do not add DDL here.
-- pkg:github/lineageos/android_frameworks_base is in the real too_many_contributors
-- table; the purl_id and its one contribution are made up for the fixture.
INSERT INTO component (purl, vendor, component, purl_id, source_purl_id) VALUES ('pkg:github/lineageos/android_frameworks_base', NULL, NULL, '0b5c1a8e-0000-4000-8000-00000000tmc1', NULL);
INSERT INTO contribution (purl_id, contributor_id) VALUES ('0b5c1a8e-0000-4000-8000-00000000tmc1', 1197052);
INSERT INTO too_many_contributors (purl_id) VALUES ('0b5c1a8e-0000-4000-8000-00000000tmc1');
-- pkg:github/lineageos/android_kernel_asus_sdm660 is listed too; made-up id, and no
-- contribution at all, as 408 of the 1,150 listed components have no location data.
INSERT INTO component (purl, vendor, component, purl_id, source_purl_id) VALUES ('pkg:github/lineageos/android_kernel_asus_sdm660', NULL, NULL, '0b5c1a8e-0000-4000-8000-00000000tmc2', NULL);
INSERT INTO too_many_contributors (purl_id) VALUES ('0b5c1a8e-0000-4000-8000-00000000tmc2');
