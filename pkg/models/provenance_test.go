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
	"sort"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	_ "modernc.org/sqlite"
	zlog "scanoss.com/provenance/pkg/logger"
)

// setupTestModel opens an in-memory SQLite DB, loads the given fixture set and returns a provenance model.
func setupTestModel(t *testing.T, load TestDataLoader) (context.Context, *zap.SugaredLogger, *ProvenanceModel) {
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
	t.Cleanup(func() { CloseDB(db) })
	if err = load(db, nil, nil); err != nil {
		t.Fatalf("failed to load SQL test data: %v", err)
	}
	return ctx, ctxzap.Extract(ctx).Sugar(), NewProvenanceModel(db)
}

func TestContributorProvenance(t *testing.T) {
	for name, load := range TestDataSets {
		t.Run(name, func(t *testing.T) {
			ctx, s, model := setupTestModel(t, load)
			list, err := model.GetProvenanceByPurlNames(ctx, s, []string{"torvalds/uemacs", "scanoss/engine"})
			if err != nil {
				t.Fatalf("unexpected error on model request: %v", err)
			}
			// Locations without a declared location nor curated countries (NULL or '') must be skipped
			// on both schemas, while curated countries without a declared location are kept
			var got []string
			for _, p := range list {
				if p.DeclaredLocation == "" && p.CountriesID == "" {
					t.Errorf("unexpected empty location: %+v", p)
				}
				got = append(got, p.VendorName+"|"+p.DeclaredLocation+"|"+p.CountriesID)
			}
			sort.Strings(got)
			want := []string{"mscasso-scanoss||{{7}}", "perezale|Tandil|{{7}}", "scanoss-qg||{{7}}", "scanossjeronimo|Argentina|{{7}}"}
			if len(got) != len(want) {
				t.Fatalf("expected %v, got %v", want, got)
			}
			for i := range want {
				if got[i] != want[i] {
					t.Errorf("expected %v, got %v", want, got)
				}
			}
		})
	}
}

func TestContributorProvenanceEmptyAndQuoted(t *testing.T) {
	for name, load := range TestDataSets {
		t.Run(name, func(t *testing.T) {
			ctx, s, model := setupTestModel(t, load)
			list, err := model.GetProvenanceByPurlNames(ctx, s, nil)
			if err != nil || len(list) != 0 {
				t.Errorf("expected no results and no error for an empty list, got %v, %v", list, err)
			}
			// Purl names are bound as parameters, so quotes cannot break (or inject into) the query
			list, err = model.GetProvenanceByPurlNames(ctx, s, []string{"scanoss/engine') OR ('1'='1"})
			if err != nil || len(list) != 0 {
				t.Errorf("expected no results and no error for a quoted purl name, got %v, %v", list, err)
			}
		})
	}
}

func TestTooManyContributors(t *testing.T) {
	for name, load := range TestDataSets {
		t.Run(name, func(t *testing.T) {
			ctx, s, model := setupTestModel(t, load)
			purls, err := model.GetTooManyContributors(ctx, s, []string{"torvalds/linux", "scanoss/engine"})
			if err != nil {
				t.Fatalf("unexpected error on model request: %v", err)
			}
			if len(purls) != 1 || purls[0] != "torvalds/linux" {
				t.Errorf("expected [torvalds/linux], got %v", purls)
			}
			purls, err = model.GetTooManyContributors(ctx, s, []string{})
			if err != nil || len(purls) != 0 {
				t.Errorf("expected no results and no error for an empty list, got %v, %v", purls, err)
			}
		})
	}
}
