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
	"errors"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"scanoss.com/provenance/pkg/dtos"
	"scanoss.com/provenance/pkg/models"
	"scanoss.com/provenance/pkg/utils"
)

type ProvenanceUseCase struct {
	ctx  context.Context
	s    *zap.SugaredLogger
	conn *sqlx.Conn
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

func NewProvenance(ctx context.Context, conn *sqlx.Conn, s *zap.SugaredLogger) *ProvenanceUseCase {
	return &ProvenanceUseCase{ctx: ctx, conn: conn, s: s}
}

// GetProvenance takes the Provenance Input request, searches for Provenance data and returns a ProvenanceOutput struct
func (p ProvenanceUseCase) GetProvenance(request dtos.ProvenanceInput) (dtos.ProvenanceOutput, models.QuerySummary, error) {

	if len(request.Purls) == 0 {
		p.s.Info("Empty List of Purls supplied")
		return dtos.ProvenanceOutput{}, models.QuerySummary{}, errors.New("empty list of purls")
	}
	summary := models.QuerySummary{}
	var purls []string
	//Prepare purls to query
	for _, purl := range request.Purls {
		purlName, err := utils.PurlNameFromString(purl.Purl) // Make sure we just have the bare minimum for a Purl Name
		if err == nil {
			// to avoid SQL Injection
			purlName = strings.ReplaceAll(purlName, "'", "")
			purlName = strings.ReplaceAll(purlName, "\"", "")
			purls = append(purls, purlName)
		} else {
			summary.PurlsFailedToParse = append(summary.PurlsFailedToParse, purl.Purl)
		}
	}
	prov := models.NewProvenanceModel(p.ctx, p.conn)
	countries := models.NewCountryMapModel(p.ctx, p.conn)

	vendors, err := prov.GetProvenanceByPurlNames(purls)
	if err != nil {
		return dtos.ProvenanceOutput{}, models.QuerySummary{}, err
	}

	tooMany, err2many := prov.GetTooManyContributors(purls)
	if err2many != nil {
		return dtos.ProvenanceOutput{}, models.QuerySummary{}, err2many
	}

	curatedCountries := prov.ProcessCuratedVendors(vendors)
	vendorsMap := make(map[string][]models.Provenance)

	for _, v := range vendors {
		vendorsMap[v.PurlName] = append(vendorsMap[v.PurlName], v)
	}

	for _, purl := range request.Purls {

		purlName, err := utils.PurlNameFromString(purl.Purl) // Make sure we just have the bare minimum for a Purl Name
		if err == nil {
			if !(len(vendorsMap[purlName]) > 0) && !existPurl(summary.PurlsFailedToParse, purl.Purl) {
				summary.PurlsWOInfo = append(summary.PurlsWOInfo, purl.Purl)
			}
			if existPurl(tooMany, purlName) {
				summary.PurlsTooMuchData = append(summary.PurlsTooMuchData, purl.Purl)
			}
		}
	}

	retV := dtos.ProvenanceOutput{}

	//Create the response

	for _, purl := range request.Purls {
		purlName, err := utils.PurlNameFromString(purl.Purl)
		if err != nil {
			continue
		}
		listOfVendors := vendorsMap[purlName]

		var provOutItem dtos.ProvenanceOutputItem

		provOutItem.Purl = purl.Purl
		for _, vendor := range listOfVendors {
			if vendor.DeclaredLocation != "" {
				provOutItem.DeclaredLocations = append(provOutItem.DeclaredLocations, dtos.DeclaredProvenanceItem{Type: vendor.Type, Location: vendor.DeclaredLocation})
			}
		}

		//add curated values
		for k, v := range curatedCountries[purlName] {
			i, err := strconv.Atoi(k)
			if err == nil {
				countryName, err := countries.GetCountryById(i)
				if err == nil {
					provOutItem.CuratedLocations = append(provOutItem.CuratedLocations, dtos.CuratedProvenanceItem{Country: countryName, Count: v})
				}
			}
		}

		retV.Provenance = append(retV.Provenance, provOutItem)

	}
	return retV, summary, nil
}
