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

// Package usecase contains the business logic for the geo provenance service,
// orchestrating provenance and origin data for components.
package usecase

import (
	"context"
	"math"
	_ "strings"

	"github.com/jmoiron/sqlx"
	"github.com/scanoss/go-component-helper/componenthelper"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
	"go.uber.org/zap"
	"scanoss.com/provenance/pkg/dtos"
	"scanoss.com/provenance/pkg/models"
	"scanoss.com/provenance/pkg/resolver"
)

// originMaxWorkers caps the concurrency used when resolving a batch.
const originMaxWorkers = 5

type OriginUseCase struct {
	provenanceModel *models.ProvenanceModel
	resolver        *resolver.Resolver
	db              *sqlx.DB
}

func NewOrigin(db *sqlx.DB) *OriginUseCase {
	return &OriginUseCase{
		db:              db,
		provenanceModel: models.NewProvenanceModel(db),
		resolver:        resolver.New(nil, db)}
}

// GetOrigin takes the Provenance Input request, searches for Provenance data and returns a ProvenanceOutput struct
//
//goland:noinspection ALL
func (p OriginUseCase) GetOrigin(ctx context.Context, s *zap.SugaredLogger, components []componenthelper.ComponentDTO) (dtos.OriginOutput, error) {
	sanitizedComponents := resolver.ResolveComponents(ctx, s, p.resolver, components, originMaxWorkers)
	resMaps := make(map[string][]models.LocationDistribution)
	retV := dtos.OriginOutput{}
	validComponents := make([]resolver.Resolution, 0)
	for _, component := range sanitizedComponents {
		// Keep searching for components with  SUCCESS AND VERSION_NOT_FOUND status.
		if component.Status.StatusCode == domain.Success || component.Status.StatusCode == domain.VersionNotFound {
			validComponents = append(validComponents, component)
		} else {
			retV.Provenance = append(retV.Provenance, dtos.OriginOutputItem{
				Purl:   component.Purl,
				Status: component.Status,
			})
		}
	}

	// The distribution now comes from the curated country per contributor. Mining
	// used vendor_locations.timezone_based_country, a country inferred from commit
	// timezones, which serving does not carry; where mining recorded both they
	// agree on 119,906 of 120,988 vendors, but the curated set is much smaller, so
	// fewer components resolve to an origin at all.
	for _, c := range validComponents {
		counts, err := p.provenanceModel.GetCountryCountsByPurlID(ctx, s, c.PurlID)
		if err != nil {
			return dtos.OriginOutput{}, err
		}
		var total int16
		for _, v := range counts {
			total += v.ContributorCount
		}
		if total == 0 {
			continue
		}
		for _, v := range counts {
			percentage := float32(v.ContributorCount*100) / float32(total)
			resMaps[c.Name] = append(resMaps[c.Name], models.LocationDistribution{
				CountryName:           v.CountryName,
				ContributorPercentage: float32(math.Round(float64(percentage*100)) / 100),
			})
		}
	}

	// Create the response
	for _, c := range validComponents {
		origins := resMaps[c.Name]
		var origOutItem dtos.OriginOutputItem
		origOutItem.Purl = c.Purl
		if len(origins) == 0 {
			origOutItem.Status = domain.ComponentStatus{
				StatusCode: domain.ComponentWithoutInfo,
				Message:    "No Origin data found for the given Purl",
			}
			retV.Provenance = append(retV.Provenance, origOutItem)
			continue
		}
		for _, origin := range origins {
			origOutItem.Countries = append(origOutItem.Countries, dtos.CountryInfo{Name: origin.CountryName /*, Developers: origin.UserCount*/, Percentage: origin.ContributorPercentage})
		}

		{
			origOutItem.Status.StatusCode = c.Status.StatusCode
		}
		retV.Provenance = append(retV.Provenance, origOutItem)
	}

	return retV, nil
}
