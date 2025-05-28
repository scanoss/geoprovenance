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
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jmoiron/sqlx"
	common "github.com/scanoss/papi/api/commonv2"
	pb "github.com/scanoss/papi/api/geoprovenancev2"
	_ "modernc.org/sqlite"
	myconfig "scanoss.com/provenance/pkg/config"
	"scanoss.com/provenance/pkg/dtos"
	zlog "scanoss.com/provenance/pkg/logger"
	"scanoss.com/provenance/pkg/models"
)

func TestCProvenanceServer_Echo(t *testing.T) {
	ctx := context.Background()
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer models.CloseDB(db)
	myConfig, err := myconfig.NewServerConfig(nil)
	if err != nil {
		t.Fatalf("failed to load Config: %v", err)
	}
	s := NewProvenanceServer(db, myConfig)

	type args struct {
		ctx context.Context
		req *common.EchoRequest
	}
	tests := []struct {
		name    string
		s       pb.GeoProvenanceServer
		args    args
		want    *common.EchoResponse
		wantErr bool
	}{
		{
			name: "Echo",
			s:    s,
			args: args{
				ctx: ctx,
				req: &common.EchoRequest{Message: "Hello there!"},
			},
			want: &common.EchoResponse{Message: "Hello there!"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.Echo(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.Echo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("service.Echo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCProvenanceServer_GetComponentContributors(t *testing.T) {
	ctx := context.Background()
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer models.CloseDB(db)
	ctx = ctxzap.ToContext(ctx, zlog.L)

	err = models.LoadTestSqlData(db, nil, nil)
	if err != nil {
		fmt.Println(err)
	}

	myConfig, err := myconfig.NewServerConfig(nil)
	if err != nil {
		t.Fatalf("failed to load Config: %v", err)
	}

	s := NewProvenanceServer(db, myConfig)

	tests := []struct {
		name             string
		request          string
		expectedResponse dtos.ProvenanceOutput
		expectError      bool
	}{
		{
			name:    "Should_Return_OneResult",
			request: `{"Purls":[ {"Purl":"pkg:github/scanoss/engine"},{"Purl":"pkg:github/torvalds/uemacs"}]}`,
			expectedResponse: dtos.ProvenanceOutput{
				Provenance: []dtos.ProvenanceOutputItem{
					{
						Purl: "pkg:github/scanoss/engine",
						DeclaredLocations: []dtos.DeclaredProvenanceItem{
							{
								Type:     "User",
								Location: "Tandil",
							},
							{
								Type:     "User",
								Location: "Argentina",
							},
						},
						CuratedLocations: []dtos.CuratedProvenanceItem{
							{
								Country: "Argentina",
								Count:   2,
							},
						},
					},
					{
						Purl:              "pkg:github/torvalds/uemacs",
						DeclaredLocations: []dtos.DeclaredProvenanceItem{},
						CuratedLocations:  []dtos.CuratedProvenanceItem{},
					},
				},
			},
			expectError: false,
		},
		{
			name:    "Should_ReturnError_NoDataSupplied",
			request: `{"Purls":[]}`,
			expectedResponse: dtos.ProvenanceOutput{
				Provenance: []dtos.ProvenanceOutputItem{},
			},
			expectError: true,
		},
		{
			name:    "Should_ReturnSucceedWithWarning_FailedToParse",
			request: `{"Purls":[ {"Purl":"pk:github/scanoss/engine"} ]}`,
			expectedResponse: dtos.ProvenanceOutput{
				Provenance: []dtos.ProvenanceOutputItem{},
			},
			expectError: true,
		},
		{
			name:    "Should_ReturnSucceed",
			request: `{"Purls":[ {"Purl":""} ]}`,
			expectedResponse: dtos.ProvenanceOutput{
				Provenance: []dtos.ProvenanceOutputItem{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var request common.PurlRequest
			err := json.Unmarshal([]byte(tt.request), &request)
			if err != nil {
				t.Errorf("an error '%s' was not expected when parsing input json", err)
			}
			r, errReq := s.GetComponentContributors(ctx, &request)
			if errReq != nil && !tt.expectError {
				t.Logf("unexpected error on request %+v", errReq)
			}
			var rcv dtos.ProvenanceOutput
			jsonOut, errResp := json.Marshal(r)
			if errResp != nil {
				t.Logf("unexpected error on unmarshalling response %+v", errResp)
			}
			err = json.Unmarshal(jsonOut, &rcv)
			if err != nil {
				t.Logf("unexpected error on unmarshalling to a dto %+v", err)
			}

			if len(rcv.Provenance) != len(tt.expectedResponse.Provenance) {
				t.Errorf("service.GetOrigin() = %v, want %v", rcv, tt.expectedResponse)
			}

			for i, item := range rcv.Provenance {
				if item.Purl != tt.expectedResponse.Provenance[i].Purl {
					t.Errorf("service.GetOrigin() = %v, want %v", rcv, tt.expectedResponse)
				}
				if len(item.DeclaredLocations) != len(tt.expectedResponse.Provenance[i].DeclaredLocations) {
					t.Errorf("service.GetOrigin() = %v, want %v", rcv, tt.expectedResponse)
				}
				if len(item.CuratedLocations) != len(tt.expectedResponse.Provenance[i].CuratedLocations) {
					t.Errorf("service.GetOrigin() = %v, want %v", rcv, tt.expectedResponse)
				}
				for j, declaredLocation := range item.DeclaredLocations {
					if declaredLocation.Type != tt.expectedResponse.Provenance[i].DeclaredLocations[j].Type {
						t.Errorf("service.GetOrigin() = %v, want %v", rcv, tt.expectedResponse)
					}
				}
				for j, curatedLocation := range item.CuratedLocations {
					if curatedLocation.Country != tt.expectedResponse.Provenance[i].CuratedLocations[j].Country {
						t.Errorf("service.GetOrigin() = %v, want %v", rcv, tt.expectedResponse)
					}
				}
			}
		})
	}

	request := common.PurlRequest{Purls: []*common.PurlRequest_Purls{{Purl: "pkg:github/scanoss/engine"}, {Purl: "pkg:github/torvalds/uemacs"}}}

	got, errReq := s.GetComponentContributors(ctx, &request)
	if errReq != nil {
		t.Logf("unexpected error on request %+v", errReq)
	}
	var rcv dtos.ProvenanceOutput
	jsonOut, errResp := json.Marshal(got)
	if errResp != nil {
		t.Logf("unexpected error on unmarshalling response %+v", errResp)
	}
	err = json.Unmarshal(jsonOut, &rcv)
	if err != nil {
		t.Logf("unexpected error on unmarshalling to a dto %+v", err)
	}
	if len(rcv.Provenance) == 0 {
		t.Error("expected to get 1 result")

	} else {
		fmt.Printf("%+v\n", rcv)
		firstPurl := rcv.Provenance[0]
		if len(firstPurl.DeclaredLocations) == 0 {
			t.Error("expected to get at least 1 declared location")
		} else if len(firstPurl.CuratedLocations) == 0 {
			t.Error("expected to get at least 1 curated location")
		} else {
			firstCuratedCountry := firstPurl.CuratedLocations[0]
			if firstCuratedCountry.Country != "Argentina" && firstCuratedCountry.Country != "Spain" && firstCuratedCountry.Country != "Afghanistan" {
				t.Errorf("Curated country (%s) was not expected", firstCuratedCountry.Country)
			}
		}

	}

}

func TestProvenanceServer_GetOrigin(t *testing.T) {
	ctx := context.Background()
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer models.CloseDB(db)
	ctx = ctxzap.ToContext(ctx, zlog.L)

	err = models.LoadTestSqlData(db, nil, nil)
	if err != nil {
		fmt.Println(err)
	}

	myConfig, err := myconfig.NewServerConfig(nil)
	if err != nil {
		t.Fatalf("failed to load Config: %v", err)
	}

	s := NewProvenanceServer(db, myConfig)

	tests := []struct {
		name             string
		request          string
		expectedResponse dtos.OriginOutput
		expectError      bool
	}{
		{
			name:    "Should_Return_OneResult",
			request: `{"Purls":[ {"Purl":"pkg:github/scanoss/engine"},{"Purl":"pkg:github/torvalds/uemacs"}]}`,
			expectedResponse: dtos.OriginOutput{
				Provenance: []dtos.OriginOutputItem{
					{
						Purl: "pkg:github/scanoss/engine",
						Countries: []dtos.CountryInfo{
							{Name: "BR", Percentage: 25, UserCount: 0},
							{Name: "AR", Percentage: 25, UserCount: 0},
							{Name: "?", Percentage: 25, UserCount: 0},
							{Name: "CO", Percentage: 25, UserCount: 0},
						},
					},
					{
						Purl:      "pkg:github/torvalds/uemacs",
						Countries: []dtos.CountryInfo{},
					},
				},
			},
			expectError: false,
		},
		{
			name:    "Should_ReturnError_NoDataSupplied",
			request: `{"Purls":[]}`,
			expectedResponse: dtos.OriginOutput{
				Provenance: []dtos.OriginOutputItem{},
			},
			expectError: true,
		},
		{
			name:    "Should_ReturnSucceedWithWarning_FailedToParse",
			request: `{"Purls":[ {"Purl":"pk:github/scanoss/engine"} ]}`,
			expectedResponse: dtos.OriginOutput{
				Provenance: []dtos.OriginOutputItem{},
			},
			expectError: true,
		},
		{
			name:    "Should_ReturnSucceed",
			request: `{"Purls":[ {"Purl":"pkg:github/scanoss/engines"} ]}`,
			expectedResponse: dtos.OriginOutput{
				Provenance: []dtos.OriginOutputItem{
					{
						Purl:      "pkg:github/scanoss/engine",
						Countries: nil,
					},
				},
			},
			expectError: false,
		},
		{
			name:    "Should_ReturnSucceed",
			request: `{"Purls":[ {"Purl":""} ]}`,
			expectedResponse: dtos.OriginOutput{
				Provenance: []dtos.OriginOutputItem{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var request common.PurlRequest
			err := json.Unmarshal([]byte(tt.request), &request)
			if err != nil {
				t.Errorf("an error '%s' was not expected when parsing input json", err)
			}
			r, errReq := s.GetComponentOrigin(ctx, &request)
			if errReq != nil && !tt.expectError {
				t.Logf("unexpected error on request %+v", errReq)
			}
			var rcv dtos.OriginOutput
			jsonOut, errResp := json.Marshal(r)
			if errResp != nil {
				t.Logf("unexpected error on unmarshalling response %+v", errResp)
			}
			err = json.Unmarshal(jsonOut, &rcv)
			if err != nil {
				t.Logf("unexpected error on unmarshalling to a dto %+v", err)
			}

			if len(rcv.Provenance) != len(tt.expectedResponse.Provenance) {
				t.Errorf("service.GetOrigin() = %v, want %v", rcv, tt.expectedResponse)
			}
		})
	}
}
