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

	conn, err := db.Connx(ctx) // Get a connection from the pool
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer models.CloseConn(conn)
	err = models.LoadTestSqlData(db, ctx, conn)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when loading test data", err)
	}
	var provRequest = `{
		      "purls": [
		        {
		          "purl": "pkg:github/scanoss/engine",
		          "requirement": "5.2.4"
		        }
		      ]
		  	}`
	myConfig, err := myconfig.NewServerConfig(nil)
	_ = myConfig
	if err != nil {
		t.Fatalf("failed to load Config: %v", err)
	}
	provUc := NewProvenance(ctx, conn, s)

	requestDto, err := dtos.ParseProvenanceInput(s, []byte(provRequest))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when parsing input json", err)
	}
	countries, notFound, err := provUc.GetProvenance(requestDto)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when getting Provenance", err)
	}
	if len(countries.Provenance[0].DeclaredLocations) == 0 {
		t.Fatalf("Expected to get at least 1 declared location")

	}
	//fmt.Println(countries)
	fmt.Printf("Provenance response: %+v, %+v\n", countries, notFound)
	var provBadRequest = `{
	   		    "purls": [
	   		        {
	   		          "purl": "pkg:npm/"
	   		        }
	   		  ]
	   		}
	   		`

	requestDto, err = dtos.ParseProvenanceInput(s, []byte(provBadRequest))

	if err != nil {
		t.Fatalf("an error '%s' was not expected when parsing input json", err)
	}

	countries, _, err = provUc.GetProvenance(requestDto)

	if err == nil && len(countries.Provenance) > 0 {
		t.Fatalf("did not get an expected error: %v", countries)
	}

	emptyReq, err := dtos.ParseProvenanceInput(s, []byte(`{ "purls": [] }`))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when parsing input json", err)
	}
	fmt.Printf("requestDto: %+v\n", emptyReq)
	countries, _, err = provUc.GetProvenance(emptyReq)
	if err == nil {
		t.Fatalf("did not get an expected error: %v", countries)
	}

}
