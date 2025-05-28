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
	"math"
	"strings"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"scanoss.com/provenance/pkg/dtos"
	"scanoss.com/provenance/pkg/models"
	"scanoss.com/provenance/pkg/utils"
)

type OriginUseCase struct {
	ctx  context.Context
	s    *zap.SugaredLogger
	conn *sqlx.Conn
}

func NewOrigin(ctx context.Context, conn *sqlx.Conn) *OriginUseCase {
	return &OriginUseCase{ctx: ctx, conn: conn}
}

// GetOrigin takes the Provenance Input request, searches for Provenance data and returns a ProvenanceOutput struct
//
//goland:noinspection ALL
func (p OriginUseCase) GetOrigin(request dtos.ProvenanceInput) (dtos.OriginOutput, models.QuerySummary, error) {

	if len(request.Purls) == 0 {
		p.s.Info("Empty List of Purls supplied")
		return dtos.OriginOutput{}, models.QuerySummary{}, errors.New("empty list of purls")
	}
	summary := models.QuerySummary{}
	var purls []string
	resMaps := make(map[string][]models.LocationDistribution)

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
	// Query Origin for each purl and count amount of users per each
	mapTotal := make(map[string]int16)
	for _, p := range purls {
		mapOrigins := make(map[string]int16)
		tz, _ := prov.GetTimeZoneOriginByPurlName(p)
		for _, v := range tz {
			if count, exist := mapOrigins[v.CountryName]; !exist {
				mapOrigins[v.CountryName] = int16(v.ContributorCount)

			} else {
				mapOrigins[v.CountryName] = count + int16(v.ContributorCount)
			}
			mapTotal[p] += int16(v.ContributorCount)
		}

		for k, v := range mapOrigins {
			var percentage = float32(v*100) / float32(mapTotal[p])
			resMaps[p] = append(resMaps[p], models.LocationDistribution{CountryName: k, ContributorPercentage: float32(math.Round(float64(percentage*100)) / 100)})
		}
	}

	retV := dtos.OriginOutput{}
	tooMany, err2many := prov.GetTooManyContributors(purls)
	if err2many != nil {
		return dtos.OriginOutput{}, models.QuerySummary{}, err2many
	}

	//Create the response
	for _, purl := range request.Purls {
		purlName, err := utils.PurlNameFromString(purl.Purl)
		if err != nil {
			continue
		}
		origins := resMaps[purlName]
		var origOutItem dtos.OriginOutputItem
		origOutItem.Purl = purl.Purl
		for _, origin := range origins {
			origOutItem.Countries = append(origOutItem.Countries, dtos.CountryInfo{Name: origin.CountryName /*, Developers: origin.UserCount*/, Percentage: origin.ContributorPercentage})
		}
		retV.Provenance = append(retV.Provenance, origOutItem)

	}

	// Check if results should have a "too many contributors Warning"
	for _, purl := range request.Purls {
		purlName, err := utils.PurlNameFromString(purl.Purl)
		if err == nil {
			if existPurl(tooMany, purlName) {
				summary.PurlsTooMuchData = append(summary.PurlsTooMuchData, purl.Purl)
			}
		}
	}

	return retV, summary, nil
}
