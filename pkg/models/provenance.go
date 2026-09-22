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
	db *sqlx.DB
}

type Provenance struct {
	Type             string `db:"type"`
	PurlName         string `db:"purl_name"`
	VendorName       string `db:"vendor_name"`
	DeclaredLocation string `db:"declared_location"`
	CountriesID      string `db:"countries_id"`
}

type Origin struct {
	CountryName      string `db:"country"`
	ContributorCount int    `db:"vendor_count"`
}

type LocationDistribution struct {
	CountryName           string
	ContributorPercentage float32
}

// NewProvenanceModel creates a new instance of a provenance Model.
func NewProvenanceModel(db *sqlx.DB) *ProvenanceModel {
	return &ProvenanceModel{db: db}
}

// ProcessCuratedVendors assigns a list of country name to given set of id's of a set of provenance records.
func (m *ProvenanceModel) ProcessCuratedVendors(vendors []Provenance) map[string]map[string]int {
	curatedCountries := make(map[string]map[string]int)
	for _, v := range vendors {
		if v.CountriesID != "" {
			listStr := strings.ReplaceAll(v.CountriesID, "{", "")
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

// githubMineID is the mine ID of github.com, the only mine with contributor data.
// It is inlined in the queries (not bound) so SQLite compares it against TEXT columns correctly.
const githubMineID = "5"

// GetProvenanceByPurlNames get declared and curated locations for contributors and authors from a list of purlnames.
func (m *ProvenanceModel) GetProvenanceByPurlNames(ctx context.Context, s *zap.SugaredLogger, purlNames []string) ([]Provenance, error) {
	if len(purlNames) == 0 {
		return []Provenance{}, nil
	}
	var allSources []Provenance
	// Missing values may be NULL or '' (both engines store empty declared locations as '').
	// Curated IDs are cast to TEXT because they are an integer array in PostgreSQL.
	// A location is kept if it has a declared location or curated countries (curated countries
	// can exist without a declared location).
	query := `
		    SELECT DISTINCT
		        gc.purl_name AS purl_name,
		        vd.type AS type,
		        vd.username AS vendor_name,
		        COALESCE(vl.declared_location, '') AS declared_location,
		        CASE
		            WHEN COALESCE(CAST(vl.curated_countries_ids AS TEXT), '') IN ('', '{}') THEN ''
		            ELSE CAST(vl.curated_countries_ids AS TEXT)
		        END AS countries_id
		    FROM vendors vd
		    JOIN github_contributors gc ON gc.contributor = vd.username
		    JOIN vendor_locations vl ON vl.vendor_id = vd.id
		    WHERE gc.purl_name IN (` + placeholders(len(purlNames)) + `)
		      AND COALESCE(vd.type, '') <> ''
		      AND vd.mine_id = ` + githubMineID + `
		      AND (COALESCE(vl.declared_location, '') <> ''
		           OR COALESCE(CAST(vl.curated_countries_ids AS TEXT), '') NOT IN ('', '{}'));`

	err := m.db.SelectContext(ctx, &allSources, query, toArgs(purlNames)...)
	if err != nil {
		s.Errorf("Error: Failed to query %v: %+v", purlNames, err)
		return nil, fmt.Errorf("failed to query : %v", err)
	}
	return allSources, nil
}

// GetTooManyContributors returns the subset of the given purl names flagged as having too many contributors.
func (m *ProvenanceModel) GetTooManyContributors(ctx context.Context, s *zap.SugaredLogger, purlNames []string) ([]string, error) {
	if len(purlNames) == 0 {
		return []string{}, nil
	}
	var purls []string
	query := `
			SELECT tmc.purl_name
			FROM too_many_contributors tmc
			WHERE tmc.purl_name IN (` + placeholders(len(purlNames)) + `)
		      AND tmc.mine_id = ` + githubMineID + `;`
	err := m.db.SelectContext(ctx, &purls, query, toArgs(purlNames)...)
	if err != nil {
		s.Errorf("Error: Failed to query %v: %+v", purlNames, err)
		return nil, fmt.Errorf("failed to query : %v", err)
	}

	return purls, nil
}

// GetTimeZoneOriginByPurlName returns the number of GitHub contributors per timezone based country for a purl name.
func (m *ProvenanceModel) GetTimeZoneOriginByPurlName(ctx context.Context, s *zap.SugaredLogger, purlName string) ([]Origin, error) {
	var allSources []Origin
	query := `
		SELECT
		  vl.timezone_based_country AS country,
		  COUNT(DISTINCT v.id) AS vendor_count
		FROM github_contributors gc
		JOIN vendors v ON gc.contributor = v.username
		JOIN vendor_locations vl ON v.id = vl.vendor_id
		WHERE gc.purl_name = $1
		  AND v.mine_id = ` + githubMineID + `
		  AND COALESCE(vl.timezone_based_country, '') <> ''
		GROUP BY vl.timezone_based_country
		ORDER BY vendor_count DESC;`
	err := m.db.SelectContext(ctx, &allSources, query, purlName)
	if err != nil {
		s.Errorf("Error: Failed to query %v: %+v", purlName, err)
		return nil, fmt.Errorf("failed to query : %v", err)
	}
	return allSources, nil
}
