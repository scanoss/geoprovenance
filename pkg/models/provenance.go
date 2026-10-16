// SPDX-License-Identifier: GPL-2.0-or-later
/*
 * Copyright (C) 2018-2025 SCANOSS.COM
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

// Handle all interaction with the serving tables that carry geoprovenance.

package models

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// ProvenanceModel reads where a component's contributors are from.
//
// Mining spread this across four tables: `github_contributors` linked a purl name
// to a contributor login, `vendors` turned that login into a vendor id and a type
// (User or Organization), and `vendor_locations` held the vendor's declared
// location, a timezone-inferred country and a curated country id. Serving keys
// contributions by the component's purl_id and splits the locations in two:
//
//   - `vendor_location(vendor_id, country_id)` is the curated country;
//   - `vendor_location_declared(vendor_id, declared_location, timezone_country,
//     timezone_certainty)` is the declared location and the timezone-inferred
//     country, which the origin endpoint has always been based on.
//
// `vendor` carries the login, type and mine, and `too_many_contributors` lists
// the components whose contributors were not mined.
type ProvenanceModel struct {
	db *sqlx.DB
}

// Provenance is one contributor's declared location.
type Provenance struct {
	Type             string `db:"type"`
	PurlName         string `db:"purl_name"`
	VendorName       string `db:"vendor_name"`
	DeclaredLocation string `db:"declared_location"`
	CountriesID      string `db:"countries_id"`
}

// Origin is one country and how many of the component's contributors it accounts
// for. It backs both the curated locations and the origin distribution.
type Origin struct {
	CountryName      string `db:"country"`
	ContributorCount int16  `db:"vendor_count"`
}

type LocationDistribution struct {
	CountryName           string
	ContributorPercentage float32
}

// NewProvenanceModel creates a new instance of a provenance Model.
func NewProvenanceModel(db *sqlx.DB) *ProvenanceModel {
	return &ProvenanceModel{db: db}
}

// countryCountsQuery counts a component's contributors per curated country.
const countryCountsQuery = `SELECT c.name AS country, COUNT(DISTINCT vl.vendor_id) AS vendor_count
FROM contribution ct
INNER JOIN vendor_location vl ON vl.vendor_id = ct.contributor_id
INNER JOIN country c ON c.id = vl.country_id
WHERE ct.purl_id = ? AND c.name IS NOT NULL AND c.name <> ''
GROUP BY c.name
ORDER BY vendor_count DESC, country ASC`

// timezoneOriginQuery counts a component's contributors per timezone-inferred
// country, as mining's origin query did over vendor_locations.timezone_based_country.
// The country is an ISO code (DE, US), not a name, exactly as mining returned it.
// A vendor recorded in two countries counts in both, as it did there.
const timezoneOriginQuery = `SELECT vld.timezone_country AS country, COUNT(DISTINCT vld.vendor_id) AS vendor_count
FROM contribution ct
INNER JOIN vendor_location_declared vld ON vld.vendor_id = ct.contributor_id
WHERE ct.purl_id = ? AND vld.timezone_country IS NOT NULL AND vld.timezone_country <> ''
GROUP BY vld.timezone_country
ORDER BY vendor_count DESC, country ASC`

// declaredLocationsQuery lists the locations a component's GitHub contributors
// declare on their profiles. The filters are mining's: a vendor with a type, from
// the GitHub mine (5), and a declared location that is not empty.
const declaredLocationsQuery = `SELECT DISTINCT v.type AS type, v.username AS vendor_name,
	vld.declared_location AS declared_location
FROM contribution ct
INNER JOIN vendor v ON v.id = ct.contributor_id
INNER JOIN vendor_location_declared vld ON vld.vendor_id = v.id
WHERE ct.purl_id = ? AND v.type IS NOT NULL AND v.mine_id = 5
  AND vld.declared_location IS NOT NULL AND vld.declared_location <> ''
ORDER BY vendor_name, declared_location`

// tooManyContributorsQuery tells whether a component is one whose contributors
// were too many to mine.
const tooManyContributorsQuery = `SELECT COUNT(*) FROM too_many_contributors WHERE purl_id = ?`

// GetCountryCountsByPurlID returns the component's contributor distribution by
// country, highest first.
func (m *ProvenanceModel) GetCountryCountsByPurlID(ctx context.Context, s *zap.SugaredLogger, purlID string) ([]Origin, error) {
	if len(purlID) == 0 {
		return []Origin{}, nil
	}
	var counts []Origin
	if err := m.db.SelectContext(ctx, &counts, m.db.Rebind(countryCountsQuery), purlID); err != nil {
		s.Errorf("Error: Failed to query the contributor countries for %v: %v", purlID, err)
		return nil, fmt.Errorf("failed to query the contributor countries: %w", err)
	}
	return counts, nil
}

// GetTimezoneOriginByPurlID returns the component's contributor distribution by
// timezone-inferred country, highest first.
func (m *ProvenanceModel) GetTimezoneOriginByPurlID(ctx context.Context, s *zap.SugaredLogger, purlID string) ([]Origin, error) {
	if len(purlID) == 0 {
		return []Origin{}, nil
	}
	var counts []Origin
	if err := m.db.SelectContext(ctx, &counts, m.db.Rebind(timezoneOriginQuery), purlID); err != nil {
		s.Errorf("Error: Failed to query the contributor timezones for %v: %v", purlID, err)
		return nil, fmt.Errorf("failed to query the contributor timezones: %w", err)
	}
	return counts, nil
}

// GetDeclaredLocationsByPurlID returns the locations the component's contributors
// declare, one per (contributor, location).
func (m *ProvenanceModel) GetDeclaredLocationsByPurlID(ctx context.Context, s *zap.SugaredLogger, purlID string) ([]Provenance, error) {
	if len(purlID) == 0 {
		return []Provenance{}, nil
	}
	var declared []Provenance
	if err := m.db.SelectContext(ctx, &declared, m.db.Rebind(declaredLocationsQuery), purlID); err != nil {
		s.Errorf("Error: Failed to query the declared locations for %v: %v", purlID, err)
		return nil, fmt.Errorf("failed to query the declared locations: %w", err)
	}
	return declared, nil
}

// HasTooManyContributors reports whether the component is listed in
// too_many_contributors.
func (m *ProvenanceModel) HasTooManyContributors(ctx context.Context, s *zap.SugaredLogger, purlID string) (bool, error) {
	if len(purlID) == 0 {
		return false, nil
	}
	var n int
	if err := m.db.GetContext(ctx, &n, m.db.Rebind(tooManyContributorsQuery), purlID); err != nil {
		s.Errorf("Error: Failed to query too_many_contributors for %v: %v", purlID, err)
		return false, fmt.Errorf("failed to query too_many_contributors: %w", err)
	}
	return n > 0, nil
}
