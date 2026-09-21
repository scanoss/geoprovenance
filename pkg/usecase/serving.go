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

// Serving-schema path shared by the provenance and origin use cases: each purl
// is resolved to its purl_id in `component` and the contributors' country
// distribution is read from contribution -> vendor_location -> country. When
// the purl has no contributors of its own, the linked source component
// (component.source_purl_id) is tried.

package usecase

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
	"go.uber.org/zap"
	"scanoss.com/provenance/pkg/dtos"
	"scanoss.com/provenance/pkg/models"
	"scanoss.com/provenance/pkg/utils"
)

// servingResult is the outcome of resolving one purl against the serving schema.
type servingResult struct {
	Counts []models.CountryCount
	Status domain.ComponentStatus
}

// ServingPurlKey converts a request purl into the serving `component.purl`
// key. The export builds that column as 'pkg:' || purl_type || '/' || purl_name
// from the mining tables, so the key is derived exactly like the legacy path
// derives purl_name (utils.PurlNameFromString): version, qualifiers and subpath
// stripped, lower-cased except for npm/nuget, percent-encoding kept as stored
// (e.g. pkg:npm/%40angular/core), after the golang-GitHub rewrite.
func ServingPurlKey(purl string) (string, error) {
	purl = utils.ConvertPurlString(strings.TrimSpace(purl))
	name, err := utils.PurlNameFromString(purl)
	if err != nil {
		return "", err
	}
	purlType, _, ok := strings.Cut(strings.TrimPrefix(purl, "pkg:"), "/")
	if !ok || purlType == "" {
		return "", fmt.Errorf("no purl type found in '%v'", purl)
	}
	return "pkg:" + strings.ToLower(purlType) + "/" + name, nil
}

// resolveServing resolves one purl and returns its contributors' country
// counts, falling back to the linked source component when it has none.
func resolveServing(ctx context.Context, s *zap.SugaredLogger, m *models.ServingModel, purl, noInfoMsg string) (servingResult, error) {
	key, err := ServingPurlKey(purl)
	if err != nil {
		s.Warnf("Failed to parse PURL %q: %v", purl, err)
		return servingResult{Status: domain.ComponentStatus{StatusCode: domain.InvalidPurl, Message: "Invalid Purl"}}, nil
	}
	comp, found, err := m.ResolveComponent(ctx, s, key)
	if err != nil {
		return servingResult{}, err
	}
	noInfo := servingResult{Status: domain.ComponentStatus{StatusCode: domain.ComponentWithoutInfo, Message: noInfoMsg}}
	if !found {
		return noInfo, nil
	}
	counts, err := m.GetCountryCountsByPurlID(ctx, s, comp.PurlID)
	if err != nil {
		return servingResult{}, err
	}
	if len(counts) > 0 {
		return servingResult{Counts: counts, Status: domain.ComponentStatus{StatusCode: domain.Success}}, nil
	}
	if comp.SourcePurlID == "" || comp.SourcePurlID == comp.PurlID {
		return noInfo, nil
	}
	// Source-purl fallback: a package purl with no contributors linked to its
	// source repository (e.g. pkg:npm/express -> pkg:github/expressjs/express).
	counts, err = m.GetCountryCountsByPurlID(ctx, s, comp.SourcePurlID)
	if err != nil {
		return servingResult{}, err
	}
	if len(counts) == 0 {
		return noInfo, nil
	}
	return servingResult{Counts: counts, Status: domain.ComponentStatus{StatusCode: domain.Success}}, nil
}

// getProvenanceServing builds the contributors output (curated locations
// only: the serving schema carries no declared-location text).
func getProvenanceServing(ctx context.Context, s *zap.SugaredLogger, m *models.ServingModel, purls []string) (dtos.ProvenanceOutput, error) {
	retV := dtos.ProvenanceOutput{}
	for _, purl := range purls {
		res, err := resolveServing(ctx, s, m, purl, "No Provenance data found for the given Purl")
		if err != nil {
			return dtos.ProvenanceOutput{}, err
		}
		item := dtos.ProvenanceOutputItem{Purl: purl, Status: res.Status}
		for _, c := range res.Counts {
			item.CuratedLocations = append(item.CuratedLocations,
				dtos.CuratedProvenanceItem{Country: c.Country, Count: c.Count})
		}
		retV.Provenance = append(retV.Provenance, item)
	}
	return retV, nil
}

// getOriginServing builds the origin output: the country distribution as
// percentages (2 decimals), same shape as the legacy timezone-based path.
func getOriginServing(ctx context.Context, s *zap.SugaredLogger, m *models.ServingModel, purls []string) (dtos.OriginOutput, error) {
	retV := dtos.OriginOutput{}
	for _, purl := range purls {
		res, err := resolveServing(ctx, s, m, purl, "No Origin data found for the given Purl")
		if err != nil {
			return dtos.OriginOutput{}, err
		}
		item := dtos.OriginOutputItem{Purl: purl, Status: res.Status}
		total := 0
		for _, c := range res.Counts {
			total += c.Count
		}
		for _, c := range res.Counts {
			percentage := float64(c.Count) * 100 / float64(total)
			item.Countries = append(item.Countries, dtos.CountryInfo{
				Name:       c.Country,
				Percentage: float32(math.Round(percentage*100) / 100),
			})
		}
		retV.Provenance = append(retV.Provenance, item)
	}
	return retV, nil
}
