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

package models

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
)

// The fixture's pkg:github/scanoss/engine, whose contributors carry four located
// countries, and minr with one.
const (
	engineID = "b5ccd188-eb1c-5476-bfc4-b6a44d5bacc4"
	minrID   = "b3cc9409-ad24-556c-8bdc-57c4393ef3d0"
)

func setupProvenanceModel(t *testing.T) (context.Context, *sqlx.DB, *ProvenanceModel) {
	t.Helper()
	if err := zlog.NewSugaredDevLogger(); err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	ctx := context.Background()
	db := sqliteSetup(t)
	if err := LoadTestSQLData(db, ctx, nil); err != nil {
		t.Fatalf("failed to load SQL test data: %v", err)
	}
	return ctx, db, NewProvenanceModel(db)
}

func TestGetCountryCountsByPurlID(t *testing.T) {
	ctx, db, model := setupProvenanceModel(t)
	defer CloseDB(db)
	defer zlog.SyncZap()

	t.Run("component with located contributors", func(t *testing.T) {
		counts, err := model.GetCountryCountsByPurlID(ctx, zlog.S, engineID)
		if err != nil {
			t.Fatalf("GetCountryCountsByPurlID error = %v", err)
		}
		if len(counts) == 0 {
			t.Fatal("no countries returned for a component the fixture locates")
		}
		total := 0
		for _, c := range counts {
			if c.CountryName == "" {
				t.Error("a country came back with no name; the query filters those out")
			}
			if c.ContributorCount <= 0 {
				t.Errorf("country %q has a count of %d", c.CountryName, c.ContributorCount)
			}
			total += int(c.ContributorCount)
		}
		if total != 4 {
			t.Errorf("contributor total = %d, want 4", total)
		}
		// Highest first, so the caller can present the distribution as-is.
		for i := 1; i < len(counts); i++ {
			if counts[i-1].ContributorCount < counts[i].ContributorCount {
				t.Errorf("counts are not ordered highest-first: %+v", counts)
				break
			}
		}
	})

	t.Run("component with one located contributor", func(t *testing.T) {
		counts, err := model.GetCountryCountsByPurlID(ctx, zlog.S, minrID)
		if err != nil {
			t.Fatalf("GetCountryCountsByPurlID error = %v", err)
		}
		if len(counts) != 1 || counts[0].ContributorCount != 1 {
			t.Errorf("counts = %+v, want a single country with one contributor", counts)
		}
	})

	t.Run("unknown or empty component", func(t *testing.T) {
		for name, id := range map[string]string{
			"unknown": "00000000-0000-0000-0000-000000000000",
			"empty":   "",
		} {
			t.Run(name, func(t *testing.T) {
				counts, err := model.GetCountryCountsByPurlID(ctx, zlog.S, id)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(counts) != 0 {
					t.Errorf("got %d countries, want none", len(counts))
				}
			})
		}
	})
}

// A missing table is a query failure, not an empty result: a broken database must
// not read as "this component has no contributors anywhere".
func TestGetCountryCountsReportsQueryFailure(t *testing.T) {
	ctx, db, model := setupProvenanceModel(t)
	defer zlog.SyncZap()
	if _, err := db.Exec("DROP TABLE contribution;"); err != nil {
		t.Fatalf("failed to drop the table: %v", err)
	}
	defer CloseDB(db)

	if _, err := model.GetCountryCountsByPurlID(ctx, zlog.S, engineID); err == nil {
		t.Error("want an error once the table is gone")
	}
}
