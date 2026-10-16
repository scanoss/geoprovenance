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

	"github.com/jmoiron/sqlx"
	"github.com/scanoss/go-component-helper/componenthelper"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
	"go.uber.org/zap"
	"scanoss.com/provenance/pkg/dtos"
	"scanoss.com/provenance/pkg/errors"
	"scanoss.com/provenance/pkg/models"
	"scanoss.com/provenance/pkg/resolver"
)

// provenanceMaxWorkers caps the concurrency used when resolving a batch.
const provenanceMaxWorkers = 5

type ProvenanceUseCase struct {
	db              *sqlx.DB
	provenanceModel *models.ProvenanceModel
	countryMapModel *models.CountriesModel
	resolver        *resolver.Resolver
}
type ProvenanceWorkerStruct struct {
	URLMd5  string
	Purl    string
	Version string
}
type InternalQuery struct {
	CompletePurl    string
	PurlName        string
	Requirement     string
	SelectedVersion string
}

func NewProvenance(db *sqlx.DB) *ProvenanceUseCase {
	return &ProvenanceUseCase{
		db:              db,
		provenanceModel: models.NewProvenanceModel(db),
		countryMapModel: models.NewCountryMapModel(db),
		resolver:        resolver.New(nil, db),
	}
}

// GetProvenance takes the Provenance Input request, searches for Provenance data and returns a ProvenanceOutput struct.
func (p ProvenanceUseCase) GetProvenance(ctx context.Context, s *zap.SugaredLogger, components []componenthelper.ComponentDTO) (dtos.ProvenanceOutput, error) {
	validComponents := make([]resolver.Resolution, 0)
	retV := dtos.ProvenanceOutput{}
	sanitizedComponents := resolver.ResolveComponents(ctx, s, p.resolver, components, provenanceMaxWorkers)

	for _, component := range sanitizedComponents {
		// Keep searching for components with  SUCCESS AND VERSION_NOT_FOUND status.
		if component.Status.StatusCode == domain.VersionNotFound || component.Status.StatusCode == domain.Success {
			validComponents = append(validComponents, component)
		} else {
			retV.Provenance = append(retV.Provenance, dtos.ProvenanceOutputItem{
				Purl:   component.Purl,
				Status: component.Status,
			})
		}
	}

	for _, c := range validComponents {
		var provOutItem dtos.ProvenanceOutputItem
		provOutItem.Purl = c.OriginalPurl

		declared, err := p.provenanceModel.GetDeclaredLocationsByPurlID(ctx, s, c.PurlID)
		if err != nil {
			return dtos.ProvenanceOutput{}, err
		}
		counts, err := p.provenanceModel.GetCountryCountsByPurlID(ctx, s, c.PurlID)
		if err != nil {
			return dtos.ProvenanceOutput{}, err
		}
		if len(declared) == 0 && len(counts) == 0 {
			provOutItem.Status = domain.ComponentStatus{
				StatusCode: domain.ComponentWithoutInfo,
				Message:    "No Provenance data found for the given Purl",
			}
			retV.Provenance = append(retV.Provenance, provOutItem)
			continue
		}
		for _, d := range declared {
			provOutItem.DeclaredLocations = append(provOutItem.DeclaredLocations,
				dtos.DeclaredProvenanceItem{Type: d.Type, Location: d.DeclaredLocation})
		}
		for _, count := range counts {
			provOutItem.CuratedLocations = append(provOutItem.CuratedLocations,
				dtos.CuratedProvenanceItem{Country: count.CountryName, Count: int(count.ContributorCount)})
		}
		provOutItem.Status = domain.ComponentStatus{StatusCode: domain.Success}
		tooMany, err := p.provenanceModel.HasTooManyContributors(ctx, s, c.PurlID)
		if err != nil {
			return dtos.ProvenanceOutput{}, err
		}
		if tooMany {
			provOutItem.Status = domain.ComponentStatus{
				StatusCode: domain.TooManyContributors,
				Message:    "Too many contributors for " + c.OriginalPurl,
			}
		}
		retV.Provenance = append(retV.Provenance, provOutItem)
	}
	if len(retV.Provenance) == 0 {
		return dtos.ProvenanceOutput{}, errors.NewNotFoundError("No Provenance data found for the given Purl(s)")
	}
	return retV, nil
}
