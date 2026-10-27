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

package models

import (
	"testing"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

// sqliteSetup opens an in-memory SQLite database for a test.
func sqliteSetup(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	return db
}

func TestDbLoad(t *testing.T) {
	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	// The fixtures carry data only, so the schema has to exist first.
	if err = LoadTestSchema(db, nil, nil); err != nil {
		t.Fatalf("failed to load the test schema: %v", err)
	}
	err = loadSQLData(db, nil, nil, "./tests/country.sql")
	if err != nil {
		t.Errorf("failed to load SQL test data: %v", err)
	}
	err = LoadTestSQLData(db, nil, nil)
	if err != nil {
		t.Errorf("failed to load SQL test data: %v", err)
	}
	err = loadSQLData(db, nil, nil, "./tests/does-not-exist.sql")
	if err == nil {
		t.Errorf("did not fail to load SQL test data")
	}
	err = loadTestSQLDataFiles(db, nil, nil, []string{"./tests/does-not-exist.sql"})
	if err == nil {
		t.Errorf("did not fail to load SQL test data")
	}
}
