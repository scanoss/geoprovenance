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

// Handle all interaction with the many tables to get provenance

package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type ProvenanceModel struct {
	ctx  context.Context
	s    *zap.SugaredLogger
	conn *sqlx.Conn
}

type Provenance struct {
	Type             string `db:"type"`
	PurlName         string `db:"purl_name"`
	VendorName       string `db:"vendor_name"`
	DeclaredLocation string `db:"declared_location"`
	CountriesId      string `db:"countries_id"`
}

type Origin struct {
	CountryName      string `db:"country"`
	ContributorCount int16  `db:"vendor_count"`
}

type LocationDistribution struct {
	CountryName           string
	ContributorPercentage float32
}

// NewProvenanceModel creates a new instance of a provenance Model
func NewProvenanceModel(ctx context.Context, conn *sqlx.Conn) *ProvenanceModel {
	return &ProvenanceModel{ctx: ctx, conn: conn}
}

// ProcessCuratedVendors assigns a list of country name to given set of id's of a set of provenance records
func (m *ProvenanceModel) ProcessCuratedVendors(vendors []Provenance) map[string]map[string]int {
	curatedCountries := make(map[string]map[string]int)
	for _, v := range vendors {
		if v.CountriesId != "" {
			listStr := strings.ReplaceAll(v.CountriesId, "{", "")
			listStr = strings.ReplaceAll(listStr, "}", "")
			list := strings.Split(listStr, ",")
			if len(list) == 0 {
				list = append(list, listStr)
			}
			_, exist := curatedCountries[v.PurlName]
			if !exist {
				curatedCountries[v.PurlName] = make(map[string]int)
			}
			curatedCountries[v.PurlName][list[0]]++
		}
	}
	return curatedCountries
}

// GetProvenanceByPurlNames get declared and curated locations for contributors and authors from a list of purlnames
func (m *ProvenanceModel) GetProvenanceByPurlNames(purlNames []string) ([]Provenance, error) {
	list := ""
	list = strings.Join(purlNames, "','")
	list = "('" + list + "')"
	var allSources []Provenance
	query := `
		    SELECT DISTINCT
		        gc.purl_name AS purl_name,
		        vd.type AS type,
		        vd.username AS vendor_name,
		        CASE
					WHEN vl.declared_location IS NULL THEN ''
					ELSE
						vl.declared_location
				END AS declared_location,
		        CASE
		            WHEN vl.curated_countries_ids IS NULL THEN ''
		            ELSE
		                CASE
		                    WHEN vl.curated_countries_ids = '{}' THEN ''
		                    ELSE concat(vl.curated_countries_ids)
		                END
		        END AS countries_id
		    FROM vendors vd
		    left JOIN github_contributors gc ON gc.contributor = vd.username
		    left JOIN vendor_locations vl ON vl.vendor_id = vd.id
		    WHERE gc.purl_name IN ` + list + `
		      AND vd.type IS NOT NULL
		      AND vd.mine_id = 5
		      AND vl.declared_location IS NOT NULL;`
	err := m.conn.SelectContext(m.ctx, &allSources, query)
	if err != nil {
		m.s.Errorf("Error: Failed to query %v: %+v", purlNames, err)
		return nil, fmt.Errorf("failed to query : %v", err)
	}
	return allSources, nil
}

// GetTooManyContributors get declared and curated locations for contributors and authors from a list of purlnames
func (m *ProvenanceModel) GetTooManyContributors(purlNames []string) ([]string, error) {
	list := ""
	list = strings.Join(purlNames, "','")
	list = "('" + list + "')"
	var purls []string
	query := ` 
			select tmc.purl_name 
			from too_many_contributors tmc 
			where tmc.purl_name in ` + list + `
		      AND tmc.mine_id = 5;`
	err := m.conn.SelectContext(m.ctx, &purls, query)
	if err != nil {
		m.s.Errorf("Error: Failed to query %v: %+v", purlNames, err)
		return nil, fmt.Errorf("failed to query : %v", err)
	}

	return purls, nil
}

func (m *ProvenanceModel) GetTimeZoneOriginByPurlName(purlName string) ([]Origin, error) {

	var allSources []Origin
	query := `
		SELECT
  vl.timezone_based_country AS country,
  COUNT(DISTINCT v.id) AS vendor_count
FROM
  github_contributors gc
JOIN
  vendors v ON gc.contributor = v.username
JOIN
  vendor_locations vl ON v.id = vl.vendor_id
WHERE
  gc.purl_name = $1
  AND vl.timezone_based_country IS NOT NULL 
GROUP BY
 country
ORDER BY
  vendor_count DESC;

`
	err := m.conn.SelectContext(m.ctx, &allSources, query, purlName)
	if err != nil {
		m.s.Errorf("Error: Failed to query %v: %+v", purlName, err)
		return nil, fmt.Errorf("failed to query : %v", err)
	}
	return allSources, nil
}
