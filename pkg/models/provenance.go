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
// (author or contributor), and `vendor_locations` held the vendor's declared
// location, a timezone-inferred country and a curated country id. Serving reduces
// it to `contribution(purl_id → contributor_id)` → `vendor_location(vendor_id →
// country_id)` → `country`, so:
//
//   - lookups key on the component's purl_id rather than its purl name;
//   - the curated country is the only geographic signal, and both the curated
//     locations and the origin distribution derive from it;
//   - there is no declared location, no vendor name and no vendor type, so the
//     declared-locations part of the response has no source;
//   - there is no too_many_contributors table.
type ProvenanceModel struct {
	db *sqlx.DB
}

// Provenance is one contributor's declared location. Serving records none, so
// nothing populates this any more; it is kept because the response DTO is built
// from it.
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

// countryCountsQuery counts a component's contributors per country.
//
// The mining origin query read `vendor_locations.timezone_based_country`, a
// country inferred from the contributor's commit timezone. Serving has no such
// column, so this counts the curated country instead. Where mining recorded both,
// the two agree on 119,906 of 120,988 vendors; the curated set is the smaller one,
// so fewer components resolve to an origin at all.
const countryCountsQuery = `SELECT c.name AS country, COUNT(DISTINCT vl.vendor_id) AS vendor_count
FROM contribution ct
INNER JOIN vendor_location vl ON vl.vendor_id = ct.contributor_id
INNER JOIN country c ON c.id = vl.country_id
WHERE ct.purl_id = ? AND c.name IS NOT NULL AND c.name <> ''
GROUP BY c.name
ORDER BY vendor_count DESC, country ASC`

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
