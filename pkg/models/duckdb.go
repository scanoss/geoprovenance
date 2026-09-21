//go:build duckdb

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

// DuckDB driver registration, compiled only with the `duckdb` build tag
// (go-duckdb is cgo). Without the tag duckdb_stub.go is compiled instead.

package models

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/marcboeker/go-duckdb/v2"
)

// DuckDBSupported reports whether this binary can open DB_DRIVER=duckdb.
const DuckDBSupported = true

func init() {
	// go-duckdb accepts `?` positional parameters; keep Rebind a no-op for it.
	sqlx.BindDriver(DriverDuckDB, sqlx.QUESTION)
}
