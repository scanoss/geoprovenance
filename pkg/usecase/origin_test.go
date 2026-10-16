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

package usecase

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jmoiron/sqlx"
	"github.com/scanoss/go-component-helper/componenthelper"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
	"go.uber.org/zap"
	_ "modernc.org/sqlite/lib"
	myconfig "scanoss.com/provenance/pkg/config"
	"scanoss.com/provenance/pkg/dtos"
	zlog "scanoss.com/provenance/pkg/logger"
	"scanoss.com/provenance/pkg/models"
)

func TestOriginUseCase(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := context.Background()
	ctx = ctxzap.ToContext(ctx, zlog.L)
	s := ctxzap.Extract(ctx).Sugar()
	_ = s
	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer models.CloseDB(db)

	// No connection is checked out here, and no CONCAT function is registered: the
	// mining origin query used concat() over an array column, and the serving one
	// does not. Holding the pool's only connection - an in-memory SQLite database
	// is per-connection - would leave the queries under test waiting for it.
	err = models.LoadTestSQLData(db, nil, nil)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when loading test data", err)
	}
	componentDTOS := []componenthelper.ComponentDTO{
		{
			Purl:        "pkg:github/scanoss/engine",
			Requirement: "5.2.4",
		},
	}
	myConfig, err := myconfig.NewServerConfig(nil)
	_ = myConfig
	if err != nil {
		t.Fatalf("failed to load Config: %v", err)
	}
	provUc := NewOrigin(db)
	countries, err := provUc.GetOrigin(ctx, s, componentDTOS)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when getting Provenance", err)
	}
	if len(countries.Provenance[0].Countries) == 0 {
		t.Fatalf("Expected to get at least 1 country")

	}
	//fmt.Println(countries)
	fmt.Printf("Provenance response: %+v\n", countries)
	componentDTOS = []componenthelper.ComponentDTO{
		{
			Purl: "pkg:npm/",
		},
	}

	countries, err = provUc.GetOrigin(ctx, s, componentDTOS)
	if err == nil {
		if len(countries.Provenance) == 0 {
			t.Fatalf("expected at least one item with failure status, got: %v", countries)
		}
		for _, item := range countries.Provenance {
			if item.Status.StatusCode == "" || item.Status.StatusCode == domain.Success {
				t.Fatalf("expected non-success status for invalid purl, got: %v", item.Status)
			}
		}
	}

}

// setupUseCaseDB builds the fixture database the use cases run against.
func setupUseCaseDB(t *testing.T) (context.Context, *zap.SugaredLogger, *sqlx.DB) {
	t.Helper()
	if err := zlog.NewSugaredDevLogger(); err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	t.Cleanup(zlog.SyncZap)
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { models.CloseDB(db) })
	if err = models.LoadTestSQLData(db, nil, nil); err != nil {
		t.Fatalf("an error '%s' was not expected when loading test data", err)
	}
	return ctx, ctxzap.Extract(ctx).Sugar(), db
}

// TestOriginByTimezone pins the origin on known values. It is the timezone-inferred
// country as an ISO code, as mining returned it: of scanoss/engine's five
// contributors with a timezone, four point to AR and one to CO.
func TestOriginByTimezone(t *testing.T) {
	ctx, s, db := setupUseCaseDB(t)
	got, err := NewOrigin(db).GetOrigin(ctx, s, []componenthelper.ComponentDTO{{Purl: "pkg:github/scanoss/engine"}})
	if err != nil {
		t.Fatalf("GetOrigin() error = %v", err)
	}
	want := []dtos.CountryInfo{{Name: "AR", Percentage: 80}, {Name: "CO", Percentage: 20}}
	if !reflect.DeepEqual(got.Provenance[0].Countries, want) {
		t.Errorf("countries = %+v, want %+v", got.Provenance[0].Countries, want)
	}
	if code := got.Provenance[0].Status.StatusCode; code != domain.Success {
		t.Errorf("status = %v, want %v", code, domain.Success)
	}
}

// TestOriginTooManyContributors pins the status for a component listed in
// too_many_contributors: its distribution is still returned, flagged.
func TestOriginTooManyContributors(t *testing.T) {
	ctx, s, db := setupUseCaseDB(t)
	got, err := NewOrigin(db).GetOrigin(ctx, s,
		[]componenthelper.ComponentDTO{{Purl: "pkg:github/lineageos/android_frameworks_base"}})
	if err != nil {
		t.Fatalf("GetOrigin() error = %v", err)
	}
	if code := got.Provenance[0].Status.StatusCode; code != domain.TooManyContributors {
		t.Errorf("status = %v, want %v", code, domain.TooManyContributors)
	}
	if len(got.Provenance[0].Countries) != 1 || got.Provenance[0].Countries[0].Name != "DE" {
		t.Errorf("countries = %+v, want only DE", got.Provenance[0].Countries)
	}
}
