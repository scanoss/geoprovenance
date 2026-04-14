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

// Handle all interaction with the mines table

package models

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type CountriesModel struct {
	db *sqlx.DB
}
type countryRow struct {
	ID      int    `db:"id"`
	Country string `db:"country_name"`
}

// NewCountryMapModel creates a new instance of the Mine Model.
func NewCountryMapModel(db *sqlx.DB) *CountriesModel {
	return &CountriesModel{db: db}
}

var countryMap map[int]string

// GetCountryByID Gets a list of the Purl Type IDs associated with the given Purl Type (string).
func (m *CountriesModel) GetCountryByID(ctx context.Context, s *zap.SugaredLogger, id int) (string, error) {
	if countryMap == nil {
		countryMap = make(map[int]string)

		var countries []countryRow
		err := m.db.SelectContext(ctx, &countries, "SELECT id,country_name FROM countries")
		if err != nil {
			// zlog.S.Errorf("Error: Failed to query country table for %v: %v", purlType, err)
			return "", fmt.Errorf("failed to query the mines table: %v", err)
		}

		for _, country := range countries {
			countryMap[country.ID] = country.Country
		}
	}
	c, exist := countryMap[id]
	if !exist {
		s.Error("coutry not found")
		return "N/A", fmt.Errorf("coutry not found")
	} else {
		return c, nil
	}
}
