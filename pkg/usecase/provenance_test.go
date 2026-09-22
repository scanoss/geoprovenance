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
	"testing"

	_ "modernc.org/sqlite"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jmoiron/sqlx"
	"github.com/scanoss/go-component-helper/componenthelper"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
	myconfig "scanoss.com/provenance/pkg/config"
	zlog "scanoss.com/provenance/pkg/logger"
	"scanoss.com/provenance/pkg/models"
)

func TestProvenanceUseCase(t *testing.T) {
	for name, load := range models.TestDataSets {
		t.Run(name, func(t *testing.T) { testProvenanceUseCase(t, load) })
	}
}

func testProvenanceUseCase(t *testing.T, load models.TestDataLoader) {
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

	conn, err := db.Connx(ctx) // Get a connection from the pool
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer models.CloseConn(conn)
	err = load(db, nil, nil)
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
	if len(countries.Provenance[0].DeclaredLocations) != 2 {
		t.Fatalf("Expected to get 2 declared locations, got: %v", countries.Provenance[0].DeclaredLocations)
	}
	curated := countries.Provenance[0].CuratedLocations
	if len(curated) != 1 || curated[0].Country != "Argentina" || curated[0].Count != 4 {
		t.Fatalf("Expected 4 curated locations in Argentina, got: %v", curated)
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
