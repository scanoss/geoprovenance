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

package service

import (
	"context"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jmoiron/sqlx"
	common "github.com/scanoss/papi/api/commonv2"
	pb "github.com/scanoss/papi/api/geoprovenancev2"
	_ "modernc.org/sqlite"
	myconfig "scanoss.com/provenance/pkg/config"
	zlog "scanoss.com/provenance/pkg/logger"
	"scanoss.com/provenance/pkg/models"
)

// newServingServer builds the gRPC server over the serving-schema SQLite
// fixture. DB_SERVING is left on "auto" so the test also covers detection.
func newServingServer(t *testing.T) (context.Context, pb.GeoProvenanceServer) {
	t.Helper()
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	t.Cleanup(func() { models.CloseDB(db); zlog.SyncZap() })
	if err = models.LoadServingTestSQLData(db, nil, nil); err != nil {
		t.Fatalf("failed to load serving fixture: %v", err)
	}
	myConfig, err := myconfig.NewServerConfig(nil)
	if err != nil {
		t.Fatalf("failed to load Config: %v", err)
	}
	ctx := withFakeStream(ctxzap.ToContext(context.Background(), zlog.L))
	return ctx, NewProvenanceServer(db, myConfig)
}

func infoCode(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func TestServing_GetComponentContributors(t *testing.T) {
	ctx, s := newServingServer(t)
	request := common.PurlRequest{Purls: []*common.PurlRequest_Purls{ //nolint:staticcheck
		{Purl: "pkg:npm/express@4.18.2"},
		{Purl: "pkg:npm/nogeo"},
		{Purl: "pkg:github/scanoss/orphan"},
		{Purl: "pkg:npm/does-not-exist"},
		{Purl: "not-a-purl"},
	}}
	got, err := s.GetComponentContributors(ctx, &request) //nolint:staticcheck
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Purls) != 5 {
		t.Fatalf("expected 5 items, got %d: %+v", len(got.Purls), got.Purls)
	}
	byPurl := map[string]*pb.ContributorResponse_Purls{}
	for _, p := range got.Purls {
		byPurl[p.Purl] = p
	}
	// express: curated from its own contributors, declared empty in serving mode.
	express := byPurl["pkg:npm/express@4.18.2"]
	if express == nil {
		t.Fatalf("express missing (purls are echoed as requested): %+v", got.Purls)
	}
	if len(express.DeclaredLocations) != 0 {
		t.Errorf("declared_locations = %+v, want empty (serving mode)", express.DeclaredLocations)
	}
	curated := map[string]int32{}
	for _, c := range express.CuratedLocations {
		curated[c.Country] = c.Count
	}
	if curated["USA"] != 2 || curated["Spain"] != 1 || len(curated) != 2 {
		t.Errorf("curated = %+v, want USA:2 Spain:1", express.CuratedLocations)
	}
	if infoCode(express.InfoCode) != "" {
		t.Errorf("express info_code = %q, want none", infoCode(express.InfoCode))
	}
	// nogeo: no contributors of its own, resolved through its source component.
	nogeo := byPurl["pkg:npm/nogeo"]
	if len(nogeo.CuratedLocations) != 1 || nogeo.CuratedLocations[0].Country != "Germany" {
		t.Errorf("nogeo curated = %+v, want Germany via source fallback", nogeo.CuratedLocations)
	}
	// orphan / unknown: NO data.
	for _, purl := range []string{"pkg:github/scanoss/orphan", "pkg:npm/does-not-exist"} {
		if infoCode(byPurl[purl].InfoCode) != "COMPONENT_WITHOUT_INFO" {
			t.Errorf("%s info_code = %q, want COMPONENT_WITHOUT_INFO", purl, infoCode(byPurl[purl].InfoCode))
		}
	}
	if infoCode(byPurl["not-a-purl"].InfoCode) != "INVALID_PURL" {
		t.Errorf("invalid purl info_code = %q, want INVALID_PURL", infoCode(byPurl["not-a-purl"].InfoCode))
	}
}

func TestServing_GetComponentOrigin(t *testing.T) {
	ctx, s := newServingServer(t)
	request := common.PurlRequest{Purls: []*common.PurlRequest_Purls{ //nolint:staticcheck
		{Purl: "pkg:npm/express"},
		{Purl: "pkg:github/scanoss/blank"},
	}}
	got, err := s.GetComponentOrigin(ctx, &request) //nolint:staticcheck
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Purls) != 2 {
		t.Fatalf("expected 2 items, got %d: %+v", len(got.Purls), got.Purls)
	}
	pct := map[string]float32{}
	for _, l := range got.Purls[0].Locations {
		pct[l.Name] = l.Percentage
	}
	if pct["USA"] != 66.67 || pct["Spain"] != 33.33 {
		t.Errorf("percentages = %+v, want USA 66.67 / Spain 33.33", got.Purls[0].Locations)
	}
	// blank: its only contributor's country has an empty name -> no data.
	if infoCode(got.Purls[1].InfoCode) != "COMPONENT_WITHOUT_INFO" || len(got.Purls[1].Locations) != 0 {
		t.Errorf("blank = %+v, want COMPONENT_WITHOUT_INFO and no locations", got.Purls[1])
	}
}

func TestServing_GetCountryContributorsByComponent(t *testing.T) {
	ctx, s := newServingServer(t)
	got, err := s.GetCountryContributorsByComponent(ctx, &common.ComponentRequest{Purl: "pkg:golang/github.com/expressjs/express"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c := got.ComponentLocations
	if c == nil || len(c.CuratedLocations) != 1 || c.CuratedLocations[0].Country != "Germany" {
		t.Errorf("golang purl should be rewritten to github and resolve: %+v", got)
	}
}
