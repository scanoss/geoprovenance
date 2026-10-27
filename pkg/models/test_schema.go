// SPDX-License-Identifier: GPL-2.0-or-later
/*
 * Copyright (C) 2018-2025 SCANOSS.COM
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 2 of the License, or
 * (at your option) any later version.
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

// The serving schema, used to build the database the tests run against.
//
// The fixtures under tests/ used to carry their own CREATE TABLE statements,
// several of them written in PostgreSQL dialect (serial4, int4). Keeping the
// schema here means every test is built from one description, and a query naming
// a column the real database does not have fails in CI rather than in production.
//
// This is a transcription of the serving schema, taken from the production DuckDB
// file with `SELECT sql FROM duckdb_tables()`. The types are the SQLite
// equivalents, since the tests run on SQLite: DuckDB UUID and DATE both become
// TEXT, which the queries already tolerate because they cast the UUIDs.
//
// Keep this in step with the real database. It is a transcription, so it can drift;
// when a query fails in production but passes here, suspect this file first.

package models

// testSchemaDDL is the schema every DB-backed test is built from.
const testSchemaDDL = `
CREATE TABLE component (purl TEXT, vendor TEXT, component TEXT, source_purl_id TEXT, purl_id TEXT);
CREATE INDEX idx_component_purl ON component (purl);
CREATE INDEX idx_component_purl_id ON component (purl_id);

CREATE TABLE component_version (purl_id TEXT, version TEXT, version_name TEXT, version_semver TEXT);
CREATE INDEX idx_component_version_purl_id ON component_version (purl_id);

-- Who contributed to a component. Mining reached this through
-- github_contributors (purl_name → contributor login) joined to vendors to turn
-- the login into an id; serving stores the pair directly.
CREATE TABLE contribution (purl_id TEXT, contributor_id INTEGER);
CREATE INDEX idx_contribution_purl_id ON contribution (purl_id);

-- Where a contributor is from, as curated.
CREATE TABLE vendor_location (vendor_id INTEGER, country_id INTEGER, location_date TEXT);
CREATE INDEX idx_vendor_location_vendor_id ON vendor_location (vendor_id);

-- What a contributor declares on their profile, and the country their commit
-- timezones point to (an ISO code). Mining kept both on vendor_locations.
CREATE TABLE vendor_location_declared (vendor_id INTEGER, declared_location TEXT, timezone_country TEXT,
	timezone_certainty INTEGER, location_date TEXT);
CREATE INDEX idx_vendor_location_declared_vendor_id ON vendor_location_declared (vendor_id);

CREATE TABLE vendor (id INTEGER, username TEXT, type TEXT, mine_id INTEGER);
CREATE INDEX idx_vendor_id ON vendor (id);

-- Components whose contributors were too many to mine.
CREATE TABLE too_many_contributors (purl_id TEXT);

CREATE TABLE country (id INTEGER, name TEXT, code TEXT);
CREATE INDEX idx_country_id ON country (id);
`
