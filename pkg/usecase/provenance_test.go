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

	_ "modernc.org/sqlite"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jmoiron/sqlx"
	"github.com/scanoss/go-component-helper/componenthelper"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
	myconfig "scanoss.com/provenance/pkg/config"
	"scanoss.com/provenance/pkg/dtos"
	zlog "scanoss.com/provenance/pkg/logger"
	"scanoss.com/provenance/pkg/models"
)

func TestProvenanceUseCase(t *testing.T) {

	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := context.Background()
	ctx = ctxzap.ToContext(ctx, zlog.L)
	s := ctxzap.Extract(ctx).Sugar()
	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer models.CloseDB(db)

	// No connection is checked out here. An in-memory SQLite database is
	// per-connection so the pool is pinned to one, and holding it would leave the
	// queries under test waiting for it forever.
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
	provUc := NewProvenance(db)

	countries, err := provUc.GetProvenance(ctx, s, componentDTOS)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when getting Provenance", err)
	}
	if len(countries.Provenance[0].CuratedLocations) == 0 {
		t.Fatalf("Expected to get at least 1 curated location")
	}
	// The declared locations come from vendor_location_declared. Pinned on known
	// values, not on "not empty": an export that moved them to another column
	// would otherwise still pass.
	wantDeclared := []dtos.DeclaredProvenanceItem{{Type: "User", Location: "Tandil"}, {Type: "User", Location: "Argentina"}}
	if !reflect.DeepEqual(countries.Provenance[0].DeclaredLocations, wantDeclared) {
		t.Errorf("declared locations = %+v, want %+v", countries.Provenance[0].DeclaredLocations, wantDeclared)
	}
	//fmt.Println(countries)
	fmt.Printf("Provenance response: %+v\n", countries)
	componentDTOS = []componenthelper.ComponentDTO{
		{
			Purl: "pkg:npm/",
		},
	}
	countries, err = provUc.GetProvenance(ctx, s, componentDTOS)

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

	componentDTOS = []componenthelper.ComponentDTO{}
	countries, err = provUc.GetProvenance(ctx, s, componentDTOS)
	if err == nil {
		t.Fatalf("Not found error was expected")
	}
}

// TestProvenanceTooManyContributors pins the status mining reported for a component
// listed in too_many_contributors.
func TestProvenanceTooManyContributors(t *testing.T) {
	ctx, s, db := setupUseCaseDB(t)
	got, err := NewProvenance(db).GetProvenance(ctx, s,
		[]componenthelper.ComponentDTO{{Purl: "pkg:github/lineageos/android_frameworks_base"}})
	if err != nil {
		t.Fatalf("GetProvenance() error = %v", err)
	}
	if code := got.Provenance[0].Status.StatusCode; code != domain.TooManyContributors {
		t.Errorf("status = %v, want %v", code, domain.TooManyContributors)
	}
}
