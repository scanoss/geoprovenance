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

// Serving-schema reads (the normalized export consumed as sqlite / postgres /
// duckdb): component -> contribution -> vendor_location -> country, keyed by
// purl_id. The legacy mining tables (github_contributors, vendors, ...) are
// handled in provenance.go; which path runs is decided by ServingModeEnabled.

package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Supported values of DB_DRIVER.
const (
	DriverSQLite   = "sqlite"
	DriverPostgres = "postgres"
	DriverDuckDB   = "duckdb"
)

// Values of DB_SERVING: pick the serving path, the legacy path, or probe the
// database for the serving tables at startup.
const (
	ServingAuto  = "auto"
	ServingOn    = "true"
	ServingOff   = "false"
	servingGate  = "contribution" // table that gates the serving path
	baseIdentity = "component"    // table every serving package carries
)

// ServingModel reads geoprovenance facts from the serving schema.
type ServingModel struct {
	db *sqlx.DB
}

// ServingComponent is one row of the serving `component` table.
type ServingComponent struct {
	PurlID       string `db:"purl_id"`
	Purl         string `db:"purl"`
	SourcePurlID string `db:"source_purl_id"`
}

// SchemaInfo is the serving database's self-declared contract (db_version).
// Present is false when the table is absent (older exports / legacy stores).
type SchemaInfo struct {
	PackageName   string `db:"package_name"`
	SchemaVersion string `db:"schema_version"`
	DBRelease     string `db:"db_release"`
	Present       bool   `db:"-"`
}

// NewServingModel creates a new instance of the serving model.
func NewServingModel(db *sqlx.DB) *ServingModel {
	return &ServingModel{db: db}
}

// TableExists reports whether a table is queryable, portably across the
// sqlite / postgres / duckdb drivers. `name` MUST come from an in-code
// constant, never from user input: it is interpolated into the SQL.
func TableExists(ctx context.Context, db *sqlx.DB, name string) bool {
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`SELECT 1 FROM "%s" LIMIT 0`, name))
	if err != nil {
		return false
	}
	_ = rows.Close()
	return true
}

// ServingModeEnabled decides whether to read the serving schema. `mode` is
// DB_SERVING: "true" forces serving, "false" forces legacy, anything else
// ("auto") probes the database for the serving gate table.
func ServingModeEnabled(ctx context.Context, db *sqlx.DB, mode string) bool {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case ServingOn:
		return true
	case ServingOff:
		return false
	default:
		return TableExists(ctx, db, servingGate) && TableExists(ctx, db, baseIdentity)
	}
}

// ReadSchemaInfo reads db_version when present. A missing table is not an
// error: it returns SchemaInfo{Present: false}.
func ReadSchemaInfo(ctx context.Context, db *sqlx.DB) (SchemaInfo, error) {
	if !TableExists(ctx, db, "db_version") {
		return SchemaInfo{Present: false}, nil
	}
	var info SchemaInfo
	q := db.Rebind(`SELECT package_name, schema_version, db_release FROM db_version LIMIT 1`)
	if err := db.GetContext(ctx, &info, q); err != nil {
		return SchemaInfo{}, fmt.Errorf("reading db_version: %w", err)
	}
	info.Present = true
	return info, nil
}

// ResolveComponent looks up a component by its serving purl key
// ("pkg:<type>/<namespace/name>", no version). The second return is false
// when the component is unknown.
func (m *ServingModel) ResolveComponent(ctx context.Context, s *zap.SugaredLogger, purlKey string) (ServingComponent, bool, error) {
	var comp ServingComponent
	// CAST the identity columns so a DuckDB native uuid scans into Go strings;
	// a no-op for sqlite (TEXT) and postgres.
	q := m.db.Rebind(`SELECT CAST(purl_id AS VARCHAR) AS purl_id, purl,
		COALESCE(CAST(source_purl_id AS VARCHAR), '') AS source_purl_id
		FROM component WHERE purl = ?`)
	err := m.db.GetContext(ctx, &comp, q, purlKey)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return ServingComponent{}, false, nil
	case err != nil:
		s.Errorf("Error: Failed to resolve serving component %v: %+v", purlKey, err)
		return ServingComponent{}, false, fmt.Errorf("failed to resolve component: %v", err)
	}
	return comp, true, nil
}

// GetCountryCountsByPurlID returns the component's contributors' distribution
// by country: distinct vendors per country, highest first, skipping empty names.
func (m *ServingModel) GetCountryCountsByPurlID(ctx context.Context, s *zap.SugaredLogger, purlID string) ([]Origin, error) {
	if purlID == "" {
		return nil, nil
	}
	var rows []Origin
	q := m.db.Rebind(`
		SELECT c.name AS country, COUNT(DISTINCT vl.vendor_id) AS vendor_count
		FROM contribution ct
		JOIN vendor_location vl ON vl.vendor_id = ct.contributor_id
		JOIN country c ON c.id = vl.country_id
		WHERE CAST(ct.purl_id AS VARCHAR) = ? AND c.name IS NOT NULL AND c.name <> ''
		GROUP BY c.name
		ORDER BY vendor_count DESC, country ASC`)
	if err := m.db.SelectContext(ctx, &rows, q, purlID); err != nil {
		s.Errorf("Error: Failed to query serving contributions for %v: %+v", purlID, err)
		return nil, fmt.Errorf("failed to query contributions: %v", err)
	}
	return rows, nil
}
