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

package models

import (
	"context"
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// CountriesModel resolves a country id to its name. Serving names the table
// `country` and the column `name`, where mining had `countries.country_name`.
type CountriesModel struct {
	db *sqlx.DB
}

type countryRow struct {
	ID      int    `db:"id"`
	Country string `db:"name"`
}

// NewCountryMapModel creates a new instance of the Countries Model.
func NewCountryMapModel(db *sqlx.DB) *CountriesModel {
	return &CountriesModel{db: db}
}

// The country table is small (284 rows) and immutable, so it is read once and
// cached. The mutex replaces a bare package-level map that several request
// goroutines populated concurrently.
var (
	countryMapMu sync.RWMutex
	countryMap   map[int]string
)

// GetCountryByID returns the country name for an id, loading the table on first use.
func (m *CountriesModel) GetCountryByID(ctx context.Context, s *zap.SugaredLogger, id int) (string, error) {
	countryMapMu.RLock()
	loaded := countryMap != nil
	countryMapMu.RUnlock()
	if !loaded {
		if err := m.loadCountries(ctx); err != nil {
			return "", err
		}
	}
	countryMapMu.RLock()
	name, exists := countryMap[id]
	countryMapMu.RUnlock()
	if !exists {
		s.Errorf("country %d not found", id)
		return "N/A", fmt.Errorf("country %d not found", id)
	}
	return name, nil
}

func (m *CountriesModel) loadCountries(ctx context.Context) error {
	var countries []countryRow
	if err := m.db.SelectContext(ctx, &countries, "SELECT id, name FROM country"); err != nil {
		return fmt.Errorf("failed to query the country table: %w", err)
	}
	loaded := make(map[int]string, len(countries))
	for _, c := range countries {
		loaded[c.ID] = c.Country
	}
	countryMapMu.Lock()
	countryMap = loaded
	countryMapMu.Unlock()
	return nil
}
