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

package models

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
	zlog "scanoss.com/provenance/pkg/logger"
)

func newServingTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	t.Cleanup(func() { CloseDB(db); zlog.SyncZap() })
	if err = LoadServingTestSQLData(db, nil, nil); err != nil {
		t.Fatalf("failed to load serving fixture: %v", err)
	}
	return db
}

func TestServingModeEnabled(t *testing.T) {
	ctx := context.Background()
	db := newServingTestDB(t)
	if !ServingModeEnabled(ctx, db, ServingAuto) {
		t.Error("auto: expected serving mode on a serving fixture")
	}
	if ServingModeEnabled(ctx, db, ServingOff) {
		t.Error("false: expected legacy mode")
	}
	legacy, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open legacy db: %v", err)
	}
	defer CloseDB(legacy)
	if err = LoadTestSQLData(legacy, nil, nil); err != nil {
		t.Fatalf("failed to load legacy fixture: %v", err)
	}
	if ServingModeEnabled(ctx, legacy, ServingAuto) {
		t.Error("auto: expected legacy mode on the mining fixture")
	}
	if !ServingModeEnabled(ctx, legacy, ServingOn) {
		t.Error("true: expected serving mode to be forced")
	}
}

func TestReadSchemaInfo(t *testing.T) {
	ctx := context.Background()
	db := newServingTestDB(t)
	info, err := ReadSchemaInfo(ctx, db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.Present || info.PackageName != "geoprovenance" || info.SchemaVersion != "1" {
		t.Errorf("unexpected schema info: %+v", info)
	}
	legacy, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open legacy db: %v", err)
	}
	defer CloseDB(legacy)
	info, err = ReadSchemaInfo(ctx, legacy)
	if err != nil || info.Present {
		t.Errorf("expected absent db_version without error, got %+v / %v", info, err)
	}
}

func TestServingModel_ResolveComponent(t *testing.T) {
	ctx := context.Background()
	db := newServingTestDB(t)
	m := NewServingModel(db)
	comp, found, err := m.ResolveComponent(ctx, zlog.S, "pkg:npm/express")
	if err != nil || !found {
		t.Fatalf("expected express to resolve, got found=%v err=%v", found, err)
	}
	if comp.PurlID != "u-express" || comp.SourcePurlID != "u-src" {
		t.Errorf("unexpected component: %+v", comp)
	}
	comp, found, err = m.ResolveComponent(ctx, zlog.S, "pkg:github/expressjs/express")
	if err != nil || !found || comp.SourcePurlID != "" {
		t.Errorf("expected source component without source link, got %+v found=%v err=%v", comp, found, err)
	}
	_, found, err = m.ResolveComponent(ctx, zlog.S, "pkg:npm/does-not-exist")
	if err != nil || found {
		t.Errorf("expected unknown component, got found=%v err=%v", found, err)
	}
}

func TestServingModel_GetCountryCountsByPurlID(t *testing.T) {
	ctx := context.Background()
	db := newServingTestDB(t)
	m := NewServingModel(db)
	rows, err := m.GetCountryCountsByPurlID(ctx, zlog.S, "u-express")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 || rows[0].CountryName != "USA" || rows[0].ContributorCount != 2 || rows[1].CountryName != "Spain" || rows[1].ContributorCount != 1 {
		t.Errorf("unexpected distribution: %+v", rows)
	}
	rows, err = m.GetCountryCountsByPurlID(ctx, zlog.S, "u-blank")
	if err != nil || len(rows) != 0 {
		t.Errorf("expected empty-named country to be skipped, got %+v / %v", rows, err)
	}
	rows, err = m.GetCountryCountsByPurlID(ctx, zlog.S, "")
	if err != nil || rows != nil {
		t.Errorf("expected nil for empty purl id, got %+v / %v", rows, err)
	}
}
