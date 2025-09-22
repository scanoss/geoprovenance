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
	"fmt"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	common "github.com/scanoss/papi/api/commonv2"
	"scanoss.com/provenance/pkg/dtos"
	zlog "scanoss.com/provenance/pkg/logger"
)

func TestOutputConvert(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := context.Background()
	ctx = ctxzap.ToContext(ctx, zlog.L)
	s := ctxzap.Extract(ctx).Sugar()

	var outputDto = dtos.ProvenanceOutput{}

	output, err := convertProvenanceOutput(s, outputDto)
	if err != nil {
		t.Errorf("TestOutputConvert failed: %v", err)
	}
	//assert.NotNilf(t, output, "Output Provenance empty")
	fmt.Printf("Output: %v\n", output)
}

func TestInputConvert(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	var provIn = &common.PurlRequest{ //nolint:staticcheck
		Purls: []*common.PurlRequest_Purls{
			{
				Purl: "pkg:github/scanoss/scanoss.js",
			},
		},
	}
	input, err := convertProvenanceInput(provIn)
	if err != nil {
		t.Errorf("TestInputConvert failed: %v", err)
	}
	fmt.Printf("Input: %v\n", input)
}

func Test_convertProvenanceInput(t *testing.T) {
	tests := []struct {
		name        string
		request     *common.PurlRequest //nolint:staticcheck
		wantLen     int
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid request with single purl",
			request: &common.PurlRequest{ //nolint:staticcheck
				Purls: []*common.PurlRequest_Purls{
					{Purl: "pkg:github/scanoss/engine", Requirement: "latest"},
				},
			},
			wantLen:     1,
			expectError: false,
		},
		{
			name: "Valid request with multiple purls",
			request: &common.PurlRequest{ //nolint:staticcheck
				Purls: []*common.PurlRequest_Purls{
					{Purl: "pkg:github/scanoss/engine"},
					{Purl: "pkg:npm/lodash@4.17.21"},
				},
			},
			wantLen:     2,
			expectError: false,
		},
		{
			name: "Empty purls slice",
			request: &common.PurlRequest{ //nolint:staticcheck
				Purls: []*common.PurlRequest_Purls{},
			},
			expectError: true,
			errorMsg:    "No components supplied. At least one component should be supplied",
		},
		{
			name: "Nil purls slice",
			request: &common.PurlRequest{ //nolint:staticcheck
				Purls: nil,
			},
			expectError: true,
			errorMsg:    "No components supplied. At least one component should be supplied",
		},
		{
			name: "Empty purl string",
			request: &common.PurlRequest{ //nolint:staticcheck
				Purls: []*common.PurlRequest_Purls{
					{Purl: ""},
				},
			},
			expectError: true,
			errorMsg:    "Empty purl supplied. At least one component should be supplied",
		},
		{
			name: "Mix of valid and empty purls - valid wins",
			request: &common.PurlRequest{ //nolint:staticcheck
				Purls: []*common.PurlRequest_Purls{
					{Purl: "pkg:github/scanoss/engine"},
					{Purl: ""},
					{Purl: "pkg:npm/lodash"},
				},
			},
			wantLen:     2,
			expectError: false,
		},
		{
			name: "Mix of valid and empty purls - empty wins",
			request: &common.PurlRequest{ //nolint:staticcheck
				Purls: []*common.PurlRequest_Purls{
					{Purl: ""},
					{Purl: ""},
					{Purl: "pkg:npm/lodash"},
				},
			},
			expectError: true,
			errorMsg:    "Empty purl supplied. At least one component should be supplied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertProvenanceInput(tt.request)
			if tt.expectError {
				if err == nil {
					t.Errorf("convertProvenanceInput() expected error but got none")
					return
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("convertProvenanceInput() error message = %v, want %v", err.Error(), tt.errorMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("convertProvenanceInput() unexpected error = %v", err)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("convertProvenanceInput() len = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}

func Test_convertOriginOutput(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer zlog.SyncZap()

	tests := []struct {
		name        string
		output      dtos.OriginOutput
		expectError bool
	}{
		{
			name: "Valid origin output with multiple components",
			output: dtos.OriginOutput{
				Provenance: []dtos.OriginOutputItem{
					{
						Purl: "pkg:github/scanoss/engine",
						Countries: []dtos.CountryInfo{
							{Name: "US", Percentage: 60.5, UserCount: 10},
							{Name: "BR", Percentage: 39.5, UserCount: 5},
						},
					},
					{
						Purl: "pkg:npm/lodash",
						Countries: []dtos.CountryInfo{
							{Name: "FR", Percentage: 100, UserCount: 1},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "Empty origin output",
			output: dtos.OriginOutput{
				Provenance: []dtos.OriginOutputItem{},
			},
			expectError: false,
		},
		{
			name: "Component with no countries",
			output: dtos.OriginOutput{
				Provenance: []dtos.OriginOutputItem{
					{
						Purl:      "pkg:github/scanoss/engine",
						Countries: []dtos.CountryInfo{},
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertOriginOutput(zlog.S, tt.output)
			if tt.expectError {
				if err == nil {
					t.Errorf("convertOriginOutput() expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("convertOriginOutput() unexpected error = %v", err)
				return
			}
			if got == nil {
				t.Errorf("convertOriginOutput() returned nil response")
			}
		})
	}
}

func Test_componentsRequestToDTO(t *testing.T) {
	tests := []struct {
		name        string
		request     *common.ComponentsRequest
		wantLen     int
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid components request",
			request: &common.ComponentsRequest{
				Components: []*common.ComponentRequest{
					{Purl: "pkg:github/scanoss/engine", Requirement: "latest"},
					{Purl: "pkg:npm/lodash@4.17.21"},
				},
			},
			wantLen:     2,
			expectError: false,
		},
		{
			name: "Empty components slice",
			request: &common.ComponentsRequest{
				Components: []*common.ComponentRequest{},
			},
			expectError: true,
			errorMsg:    "No components supplied. At least one component should be supplied",
		},
		{
			name: "Nil components slice",
			request: &common.ComponentsRequest{
				Components: nil,
			},
			expectError: true,
			errorMsg:    "No components supplied. At least one component should be supplied",
		},
		{
			name: "Component with empty purl",
			request: &common.ComponentsRequest{
				Components: []*common.ComponentRequest{
					{Purl: ""},
				},
			},
			expectError: true,
			errorMsg:    "Empty purl supplied. At least one component should be supplied",
		},
		{
			name: "Mix of valid and empty purls",
			request: &common.ComponentsRequest{
				Components: []*common.ComponentRequest{
					{Purl: "pkg:github/scanoss/engine"},
					{Purl: ""},
					{Purl: "pkg:npm/lodash"},
				},
			},
			wantLen:     2,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := componentsRequestToDTO(tt.request)
			if tt.expectError {
				if err == nil {
					t.Errorf("componentsRequestToDTO() expected error but got none")
					return
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("componentsRequestToDTO() error message = %v, want %v", err.Error(), tt.errorMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("componentsRequestToDTO() unexpected error = %v", err)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("componentsRequestToDTO() len = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}

func Test_componentRequestToDTO(t *testing.T) {
	tests := []struct {
		name        string
		request     *common.ComponentRequest
		wantLen     int
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid component request",
			request: &common.ComponentRequest{
				Purl:        "pkg:github/scanoss/engine",
				Requirement: "latest",
			},
			wantLen:     1,
			expectError: false,
		},
		{
			name:        "Nil request",
			request:     nil,
			expectError: true,
			errorMsg:    "No component supplied. A component needs to be supplied",
		},
		{
			name: "Empty purl",
			request: &common.ComponentRequest{
				Purl: "",
			},
			expectError: true,
			errorMsg:    "No component supplied. A component needs to be supplied",
		},
		{
			name: "Valid purl without requirement",
			request: &common.ComponentRequest{
				Purl: "pkg:npm/lodash@4.17.21",
			},
			wantLen:     1,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := componentRequestToDTO(tt.request)
			if tt.expectError {
				if err == nil {
					t.Errorf("componentRequestToDTO() expected error but got none")
					return
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("componentRequestToDTO() error message = %v, want %v", err.Error(), tt.errorMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("componentRequestToDTO() unexpected error = %v", err)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("componentRequestToDTO() len = %v, want %v", len(got), tt.wantLen)
			}
			if got[0].Purl != tt.request.Purl {
				t.Errorf("componentRequestToDTO() purl = %v, want %v", got[0].Purl, tt.request.Purl)
			}
		})
	}
}

func Test_toComponentsContributorResponse(t *testing.T) {
	tests := []struct {
		name   string
		output dtos.ProvenanceOutput
		want   int // expected number of components
	}{
		{
			name: "Valid provenance output with multiple components",
			output: dtos.ProvenanceOutput{
				Provenance: []dtos.ProvenanceOutputItem{
					{
						Purl: "pkg:github/scanoss/engine",
						DeclaredLocations: []dtos.DeclaredProvenanceItem{
							{Type: "User", Location: "Argentina"},
							{Type: "Organization", Location: "Brazil"},
						},
						CuratedLocations: []dtos.CuratedProvenanceItem{
							{Country: "Argentina", Count: 5},
							{Country: "Brazil", Count: 3},
						},
					},
					{
						Purl: "pkg:npm/lodash",
						DeclaredLocations: []dtos.DeclaredProvenanceItem{
							{Type: "User", Location: "USA"},
						},
						CuratedLocations: []dtos.CuratedProvenanceItem{
							{Country: "USA", Count: 10},
						},
					},
				},
			},
			want: 2,
		},
		{
			name: "Empty provenance output",
			output: dtos.ProvenanceOutput{
				Provenance: []dtos.ProvenanceOutputItem{},
			},
			want: 0,
		},
		{
			name: "Component with empty locations",
			output: dtos.ProvenanceOutput{
				Provenance: []dtos.ProvenanceOutputItem{
					{
						Purl:              "pkg:github/scanoss/engine",
						DeclaredLocations: []dtos.DeclaredProvenanceItem{},
						CuratedLocations:  []dtos.CuratedProvenanceItem{},
					},
				},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toComponentsContributorResponse(tt.output)
			if err != nil {
				t.Errorf("toComponentsContributorResponse() unexpected error = %v", err)
				return
			}
			if len(got.ComponentsLocations) != tt.want {
				t.Errorf("toComponentsContributorResponse() components count = %v, want %v", len(got.ComponentsLocations), tt.want)
			}

			// Validate structure for non-empty results
			for i, component := range got.ComponentsLocations {
				if i < len(tt.output.Provenance) {
					expectedPurl := tt.output.Provenance[i].Purl
					if component.Purl != expectedPurl {
						t.Errorf("toComponentsContributorResponse() component[%d].Purl = %v, want %v", i, component.Purl, expectedPurl)
					}
				}
			}
		})
	}
}

func Test_toComponentContributorResponse(t *testing.T) {
	tests := []struct {
		name        string
		output      dtos.ProvenanceOutput
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid provenance output with single component",
			output: dtos.ProvenanceOutput{
				Provenance: []dtos.ProvenanceOutputItem{
					{
						Purl: "pkg:github/scanoss/engine",
						DeclaredLocations: []dtos.DeclaredProvenanceItem{
							{Type: "User", Location: "Argentina"},
						},
						CuratedLocations: []dtos.CuratedProvenanceItem{
							{Country: "Argentina", Count: 5},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "Empty provenance output",
			output: dtos.ProvenanceOutput{
				Provenance: []dtos.ProvenanceOutputItem{},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toComponentContributorResponse(tt.output)
			if tt.expectError {
				if err == nil {
					t.Errorf("toComponentContributorResponse() expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("toComponentContributorResponse() unexpected error = %v", err)
				return
			}
			if got.ComponentLocations == nil {
				t.Errorf("toComponentContributorResponse() returned nil ComponentLocations")
			}
		})
	}
}

func Test_toComponentsOriginResponse(t *testing.T) {
	tests := []struct {
		name   string
		output dtos.OriginOutput
		want   int // expected number of components
	}{
		{
			name: "Valid origin output with multiple components",
			output: dtos.OriginOutput{
				Provenance: []dtos.OriginOutputItem{
					{
						Purl: "pkg:github/scanoss/engine",
						Countries: []dtos.CountryInfo{
							{Name: "US", Percentage: 60.5, UserCount: 10},
							{Name: "BR", Percentage: 39.5, UserCount: 5},
						},
					},
					{
						Purl: "pkg:npm/lodash",
						Countries: []dtos.CountryInfo{
							{Name: "FR", Percentage: 100, UserCount: 1},
						},
					},
				},
			},
			want: 2,
		},
		{
			name: "Empty origin output",
			output: dtos.OriginOutput{
				Provenance: []dtos.OriginOutputItem{},
			},
			want: 0,
		},
		{
			name: "Component with no countries",
			output: dtos.OriginOutput{
				Provenance: []dtos.OriginOutputItem{
					{
						Purl:      "pkg:github/scanoss/engine",
						Countries: []dtos.CountryInfo{},
					},
				},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toComponentsOriginResponse(tt.output)
			if err != nil {
				t.Errorf("toComponentsOriginResponse() unexpected error = %v", err)
				return
			}
			if len(got.ComponentsLocations) != tt.want {
				t.Errorf("toComponentsOriginResponse() components count = %v, want %v", len(got.ComponentsLocations), tt.want)
			}

			// Validate structure for non-empty results
			for i, component := range got.ComponentsLocations {
				if i < len(tt.output.Provenance) {
					expectedPurl := tt.output.Provenance[i].Purl
					if component.Purl != expectedPurl {
						t.Errorf("toComponentsOriginResponse() component[%d].Purl = %v, want %v", i, component.Purl, expectedPurl)
					}
					// Check percentage conversion
					for j, location := range component.Locations {
						if j < len(tt.output.Provenance[i].Countries) {
							expectedPercentage := float32(tt.output.Provenance[i].Countries[j].Percentage)
							if location.Percentage != expectedPercentage {
								t.Errorf("toComponentsOriginResponse() location[%d].Percentage = %v, want %v", j, location.Percentage, expectedPercentage)
							}
						}
					}
				}
			}
		})
	}
}

func Test_toComponentOriginResponse(t *testing.T) {
	tests := []struct {
		name        string
		output      dtos.OriginOutput
		expectError bool
	}{
		{
			name: "Valid origin output with single component",
			output: dtos.OriginOutput{
				Provenance: []dtos.OriginOutputItem{
					{
						Purl: "pkg:github/scanoss/engine",
						Countries: []dtos.CountryInfo{
							{Name: "US", Percentage: 60.5, UserCount: 10},
							{Name: "BR", Percentage: 39.5, UserCount: 5},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "Empty origin output",
			output: dtos.OriginOutput{
				Provenance: []dtos.OriginOutputItem{},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toComponentOriginResponse(tt.output)
			if tt.expectError {
				if err == nil {
					t.Errorf("toComponentOriginResponse() expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("toComponentOriginResponse() unexpected error = %v", err)
				return
			}
			if got.ComponentLocations == nil {
				t.Errorf("toComponentOriginResponse() returned nil ComponentLocations")
			}
		})
	}
}
