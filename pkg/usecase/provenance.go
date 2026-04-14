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
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/scanoss/go-component-helper/componenthelper"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
	"go.uber.org/zap"
	"scanoss.com/provenance/pkg/dtos"
	"scanoss.com/provenance/pkg/errors"
	"scanoss.com/provenance/pkg/models"
)

type ProvenanceUseCase struct {
	db              *sqlx.DB
	provenanceModel *models.ProvenanceModel
	countryMapModel *models.CountriesModel
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

func existPurl(purls []string, purl string) bool {
	for _, r := range purls {
		if purl == r {
			return true
		}
	}
	return false
}

func NewProvenance(db *sqlx.DB) *ProvenanceUseCase {
	return &ProvenanceUseCase{
		db:              db,
		provenanceModel: models.NewProvenanceModel(db),
		countryMapModel: models.NewCountryMapModel(db),
	}
}

// GetProvenance takes the Provenance Input request, searches for Provenance data and returns a ProvenanceOutput struct.
func (p ProvenanceUseCase) GetProvenance(ctx context.Context, s *zap.SugaredLogger, components []componenthelper.ComponentDTO) (dtos.ProvenanceOutput, error) {
	validComponents := make([]componenthelper.Component, 0)
	purlNames := make([]string, 0)
	retV := dtos.ProvenanceOutput{}
	sanitizedComponents := componenthelper.GetComponentsVersion(componenthelper.ComponentVersionCfg{
		MaxWorkers: 5,
		DB:         p.db,
		Ctx:        ctx,
		S:          s,
		Input:      components,
	})

	for _, component := range sanitizedComponents {
		// Keep searching for components with  SUCCESS AND VERSION_NOT_FOUND status.
		if component.Status.StatusCode == domain.VersionNotFound || component.Status.StatusCode == domain.Success {
			validComponents = append(validComponents, component)
			purlNames = append(purlNames, component.Name)
		} else {
			retV.Provenance = append(retV.Provenance, dtos.ProvenanceOutputItem{
				Purl:   component.Purl,
				Status: component.Status,
			})
		}
	}

	tooMany, err2many := p.provenanceModel.GetTooManyContributors(ctx, s, purlNames)
	if err2many != nil {
		return dtos.ProvenanceOutput{}, err2many
	}

	vendors, err := p.provenanceModel.GetProvenanceByPurlNames(ctx, s, purlNames)
	if err != nil {
		return dtos.ProvenanceOutput{}, err
	}

	curatedCountries := p.provenanceModel.ProcessCuratedVendors(vendors)
	vendorsMap := make(map[string][]models.Provenance)

	for _, v := range vendors {
		vendorsMap[v.PurlName] = append(vendorsMap[v.PurlName], v)
	}

	// Create the response
	for _, c := range validComponents {
		var provOutItem dtos.ProvenanceOutputItem
		provOutItem.Purl = c.OriginalPurl
		if len(vendorsMap[c.Name]) == 0 {
			provOutItem.Status = domain.ComponentStatus{
				StatusCode: domain.ComponentWithoutInfo,
				Message:    "No Provenance data found for the given Purl",
			}
			retV.Provenance = append(retV.Provenance, provOutItem)
			continue
		}

		listOfVendors := vendorsMap[c.Name]

		for _, vendor := range listOfVendors {
			if vendor.DeclaredLocation != "" {
				provOutItem.DeclaredLocations = append(provOutItem.DeclaredLocations, dtos.DeclaredProvenanceItem{Type: vendor.Type, Location: vendor.DeclaredLocation})
			}
		}

		// add curated values
		for k, v := range curatedCountries[c.Name] {
			i, errAtoi := strconv.Atoi(k)
			if errAtoi == nil {
				countryName, errCountry := p.countryMapModel.GetCountryByID(ctx, s, i)
				if errCountry == nil {
					provOutItem.CuratedLocations = append(provOutItem.CuratedLocations, dtos.CuratedProvenanceItem{Country: countryName, Count: v})
				}
			}
		}

		if existPurl(tooMany, c.Name) {
			msg := "Too many contributors for " + c.OriginalPurl
			provOutItem.Status.Message = msg
			provOutItem.Status.StatusCode = domain.TooManyContributors
		}

		retV.Provenance = append(retV.Provenance, provOutItem)
	}
	if len(retV.Provenance) == 0 {
		return dtos.ProvenanceOutput{}, errors.NewNotFoundError("No Provenance data found for the given Purl(s)")
	}
	return retV, nil
}
