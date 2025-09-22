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
	"scanoss.com/provenance/pkg/errors"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"scanoss.com/provenance/pkg/dtos"
	"scanoss.com/provenance/pkg/models"
	"scanoss.com/provenance/pkg/utils"
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

// GetProvenance takes the Provenance Input request, searches for Provenance data and returns a ProvenanceOutput struct
func (p ProvenanceUseCase) GetProvenance(ctx context.Context, s *zap.SugaredLogger, components []dtos.ComponentDTO) (dtos.ProvenanceOutput, models.QuerySummary, error) {

	summary := models.QuerySummary{}
	summary.TotalPurls = len(components)
	var purls []string
	//Prepare purls to query
	for _, component := range components {
		purlName, err := utils.PurlNameFromString(component.Purl) // Make sure we just have the bare minimum for a Purl Name
		if err == nil {
			// to avoid SQL Injection
			purlName = strings.ReplaceAll(purlName, "'", "")
			purlName = strings.ReplaceAll(purlName, "\"", "")
			purls = append(purls, purlName)
		} else {
			summary.PurlsFailedToParse = append(summary.PurlsFailedToParse, component.Purl)
		}
	}

	vendors, err := p.provenanceModel.GetProvenanceByPurlNames(ctx, s, purls)
	if err != nil {
		return dtos.ProvenanceOutput{}, models.QuerySummary{}, err
	}

	tooMany, err2many := p.provenanceModel.GetTooManyContributors(ctx, s, purls)
	if err2many != nil {
		return dtos.ProvenanceOutput{}, models.QuerySummary{}, err2many
	}

	curatedCountries := p.provenanceModel.ProcessCuratedVendors(vendors)
	vendorsMap := make(map[string][]models.Provenance)

	for _, v := range vendors {
		vendorsMap[v.PurlName] = append(vendorsMap[v.PurlName], v)
	}

	for _, component := range components {

		purlName, err := utils.PurlNameFromString(component.Purl) // Make sure we just have the bare minimum for a Purl Name
		if err == nil {
			if !(len(vendorsMap[purlName]) > 0) && !existPurl(summary.PurlsFailedToParse, component.Purl) {
				summary.PurlsWOInfo = append(summary.PurlsWOInfo, component.Purl)
			}
			if existPurl(tooMany, purlName) {
				summary.PurlsTooMuchData = append(summary.PurlsTooMuchData, component.Purl)
			}
		}
	}

	retV := dtos.ProvenanceOutput{}

	//Create the response

	for _, component := range components {
		purlName, err := utils.PurlNameFromString(component.Purl)
		if err != nil {
			continue
		}
		listOfVendors := vendorsMap[purlName]

		var provOutItem dtos.ProvenanceOutputItem

		provOutItem.Purl = component.Purl
		for _, vendor := range listOfVendors {
			if vendor.DeclaredLocation != "" {
				provOutItem.DeclaredLocations = append(provOutItem.DeclaredLocations, dtos.DeclaredProvenanceItem{Type: vendor.Type, Location: vendor.DeclaredLocation})
			}
		}

		//add curated values
		for k, v := range curatedCountries[purlName] {
			i, err := strconv.Atoi(k)
			if err == nil {
				countryName, err := p.countryMapModel.GetCountryById(ctx, s, i)
				if err == nil {
					provOutItem.CuratedLocations = append(provOutItem.CuratedLocations, dtos.CuratedProvenanceItem{Country: countryName, Count: v})
				}
			}
		}

		retV.Provenance = append(retV.Provenance, provOutItem)

	}
	if len(retV.Provenance) == 0 {
		return dtos.ProvenanceOutput{}, models.QuerySummary{}, errors.NewNotFoundError("No Provenance data found for the given Purl(s)")
	}
	return retV, summary, nil
}
