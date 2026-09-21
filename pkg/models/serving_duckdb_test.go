//go:build duckdb

// SPDX-License-Identifier: GPL-2.0-or-later
/*
 * Copyright (C) 2018-2022 SCANOSS.COM
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

// Runs only with the `duckdb` build tag: exercises the serving queries against a
// real DuckDB file with the export's native types (purl_id UUID), opened read-only.

package models

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	zlog "scanoss.com/provenance/pkg/logger"
)

const duckPurlID = "0f6b2a3e-4c1d-4b8a-9e2f-1234567890ab"

func TestDuckDB_ServingQueries(t *testing.T) {
	if err := zlog.NewSugaredDevLogger(); err != nil {
		t.Fatalf("logger: %v", err)
	}
	defer zlog.SyncZap()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "geoprovenance.duckdb")

	w, err := sqlx.Connect(DriverDuckDB, path)
	if err != nil {
		t.Fatalf("open duckdb for writing: %v", err)
	}
	for _, q := range []string{
		`CREATE TABLE component (purl VARCHAR, vendor VARCHAR, component VARCHAR, source_purl_id UUID, purl_id UUID)`,
		`CREATE TABLE component_version (purl_id UUID, version VARCHAR, version_name VARCHAR, version_semver VARCHAR)`,
		`CREATE TABLE contribution (purl_id UUID, contributor_id INTEGER)`,
		`CREATE TABLE vendor_location (vendor_id INTEGER, country_id INTEGER, location_date DATE)`,
		`CREATE TABLE country (id INTEGER, name VARCHAR, code VARCHAR)`,
		`CREATE TABLE db_version (package_name TEXT NOT NULL, schema_version TEXT NOT NULL, db_release TEXT NOT NULL)`,
		`INSERT INTO component VALUES ('pkg:npm/express', 'expressjs', 'express', NULL, '` + duckPurlID + `')`,
		`INSERT INTO contribution VALUES ('` + duckPurlID + `', 10), ('` + duckPurlID + `', 11), ('` + duckPurlID + `', 12)`,
		`INSERT INTO vendor_location VALUES (10, 1, DATE '2024-01-01'), (11, 1, DATE '2024-01-01'), (12, 2, DATE '2024-01-01')`,
		`INSERT INTO country VALUES (1, 'USA', 'US'), (2, 'Spain', 'ES')`,
		`INSERT INTO db_version VALUES ('geoprovenance', '1', '26.09.21.10.00')`,
	} {
		if _, err = w.Exec(q); err != nil {
			t.Fatalf("seed (%s): %v", q, err)
		}
	}
	CloseDB(w)

	db, err := sqlx.Connect(DriverDuckDB, path+"?access_mode=read_only")
	if err != nil {
		t.Fatalf("open duckdb read-only: %v", err)
	}
	defer CloseDB(db)

	if !ServingModeEnabled(ctx, db, ServingAuto) {
		t.Fatal("expected serving mode to be detected on the duckdb file")
	}
	info, err := ReadSchemaInfo(ctx, db)
	if err != nil || !info.Present || info.PackageName != "geoprovenance" {
		t.Errorf("schema info = %+v, err %v", info, err)
	}
	m := NewServingModel(db)
	comp, found, err := m.ResolveComponent(ctx, zlog.S, "pkg:npm/express")
	if err != nil || !found {
		t.Fatalf("resolve: found=%v err=%v", found, err)
	}
	if comp.PurlID != duckPurlID || comp.SourcePurlID != "" {
		t.Errorf("component = %+v, want purl_id %s and no source", comp, duckPurlID)
	}
	rows, err := m.GetCountryCountsByPurlID(ctx, zlog.S, comp.PurlID)
	if err != nil {
		t.Fatalf("counts: %v", err)
	}
	if len(rows) != 2 || rows[0].CountryName != "USA" || rows[0].ContributorCount != 2 || rows[1].CountryName != "Spain" || rows[1].ContributorCount != 1 {
		t.Errorf("distribution = %+v, want USA:2 Spain:1", rows)
	}
}
