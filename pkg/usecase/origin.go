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
	"fmt"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
	"math"
	_ "strings"

	"github.com/jmoiron/sqlx"
	"github.com/scanoss/go-component-helper/componenthelper"
	"go.uber.org/zap"
	"scanoss.com/provenance/pkg/dtos"
	"scanoss.com/provenance/pkg/models"
)

type OriginUseCase struct {
	provenanceModel *models.ProvenanceModel
	db              *sqlx.DB
}

func NewOrigin(db *sqlx.DB) *OriginUseCase {
	return &OriginUseCase{
		db:              db,
		provenanceModel: models.NewProvenanceModel(db)}
}

// GetOrigin takes the Provenance Input request, searches for Provenance data and returns a ProvenanceOutput struct
//
//goland:noinspection ALL
func (p OriginUseCase) GetOrigin(ctx context.Context, s *zap.SugaredLogger, components []componenthelper.ComponentDTO) (dtos.OriginOutput, error) {
	sanitizedComponents := componenthelper.GetComponentsVersion(componenthelper.ComponentVersionCfg{
		MaxWorkers: 5,
		DB:         p.db,
		Ctx:        ctx,
		S:          s,
		Input:      components,
	})
	resMaps := make(map[string][]models.LocationDistribution)
	retV := dtos.OriginOutput{}
	validComponents := make([]componenthelper.Component, 0)
	purlNames := make([]string, 0)
	for _, component := range sanitizedComponents {
		if component.Status.StatusCode == domain.Success || component.Status.StatusCode != domain.VersionNotFound {
			validComponents = append(validComponents, component)
			purlNames = append(purlNames, component.Name)
		} else {
			retV.Provenance = append(retV.Provenance, dtos.OriginOutputItem{
				Purl:   component.Purl,
				Status: component.Status,
			})
		}
	}

	tooMany, err2many := p.provenanceModel.GetTooManyContributors(ctx, s, purlNames)
	if err2many != nil {
		return dtos.OriginOutput{}, err2many
	}

	// Query Origin for each purl and count amount of users per each
	mapTotal := make(map[string]int16)
	for _, c := range validComponents {
		mapOrigins := make(map[string]int16)
		tz, _ := p.provenanceModel.GetTimeZoneOriginByPurlName(ctx, s, c.Name)
		for _, v := range tz {
			if count, exist := mapOrigins[v.CountryName]; !exist {
				mapOrigins[v.CountryName] = int16(v.ContributorCount)

			} else {
				mapOrigins[v.CountryName] = count + int16(v.ContributorCount)
			}
			mapTotal[c.Name] += int16(v.ContributorCount)
		}

		for k, v := range mapOrigins {
			var percentage = float32(v*100) / float32(mapTotal[c.Name])
			resMaps[c.Name] = append(resMaps[c.Name], models.LocationDistribution{CountryName: k, ContributorPercentage: float32(math.Round(float64(percentage*100)) / 100)})
		}
	}

	//Create the response
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

		if existPurl(tooMany, c.Name) {
			msg := fmt.Sprintf("Too many contributors for %s", c.OriginalPurl)
			origOutItem.Status.Message = msg
			origOutItem.Status.StatusCode = domain.TooManyContributors
		} else {
			origOutItem.Status.StatusCode = c.Status.StatusCode
		}
		retV.Provenance = append(retV.Provenance, origOutItem)

	}

	return retV, nil
}
