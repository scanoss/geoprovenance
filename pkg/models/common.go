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

// This file common tasks for the models package

package models

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
	zlog "scanoss.com/provenance/pkg/logger"
)

// loadSQLData Load the specified SQL files into the supplied DB.
func loadSQLData(db *sqlx.DB, ctx context.Context, conn *sqlx.Conn, filename string) error {
	fmt.Printf("Loading test data file: %v\n", filename)
	file, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	if conn != nil {
		_, err = conn.ExecContext(ctx, string(file))
	} else {
		_, err = db.Exec(string(file))
	}
	if err != nil {
		return err
	}
	return nil
}

// LoadTestSQLData loads all the required test SQL files.
func LoadTestSQLData(db *sqlx.DB, ctx context.Context, conn *sqlx.Conn) error {
	files := []string{
		"../models/tests/countries.sql",
		"../models/tests/versions.sql",
		"../models/tests/golang_projects.sql",
		"../models/tests/vendor_locations.sql",
		"../models/tests/vendors.sql",
		"../models/tests/github_contributors.sql",
		"../models/tests/mines.sql",
		"../models/tests/licenses.sql",
		"../models/tests/all_urls.sql"}
	return loadTestSQLDataFiles(db, ctx, conn, files)
}

// LoadTestSQLiteSchemaData loads the same test data as LoadTestSQLData, but using the schema of the
// exported SQLite provenance database (all columns TEXT, empty strings instead of NULL). See tests/sqlite/generate.sh.
func LoadTestSQLiteSchemaData(db *sqlx.DB, ctx context.Context, conn *sqlx.Conn) error {
	files := []string{
		"../models/tests/sqlite/db_version.sql",
		"../models/tests/sqlite/countries.sql",
		"../models/tests/sqlite/versions.sql",
		"../models/tests/sqlite/vendor_locations.sql",
		"../models/tests/sqlite/vendors.sql",
		"../models/tests/sqlite/github_contributors.sql",
		"../models/tests/sqlite/too_many_contributors.sql",
		"../models/tests/sqlite/mines.sql",
		"../models/tests/sqlite/licenses.sql",
		"../models/tests/sqlite/all_urls.sql"}
	return loadTestSQLDataFiles(db, ctx, conn, files)
}

// TestDataLoader loads a set of test fixtures into the supplied DB.
type TestDataLoader func(db *sqlx.DB, ctx context.Context, conn *sqlx.Conn) error

// TestDataSets lists the fixture sets that DB backed tests should run against: the PostgreSQL
// style schema and the schema of the exported SQLite provenance database.
var TestDataSets = map[string]TestDataLoader{
	"postgres-schema": LoadTestSQLData,
	"sqlite-schema":   LoadTestSQLiteSchemaData,
}

// loadTestSQLDataFiles loads a list of test SQL files.
func loadTestSQLDataFiles(db *sqlx.DB, ctx context.Context, conn *sqlx.Conn, files []string) error {
	for _, file := range files {
		err := loadSQLData(db, ctx, conn, file)
		if err != nil {
			return err
		}
	}

	return nil
}

// placeholders returns a bind variable list ($1, $2, ..., $n) for an IN clause.
// This numbered style is accepted by both the PostgreSQL and SQLite drivers.
func placeholders(n int) string {
	vars := make([]string, n)
	for i := range vars {
		vars[i] = "$" + strconv.Itoa(i+1)
	}
	return strings.Join(vars, ",")
}

// toArgs converts a list of strings into query arguments.
func toArgs(values []string) []any {
	args := make([]any, len(values))
	for i, v := range values {
		args[i] = v
	}
	return args
}

// CloseDB closes the specified DB and logs any errors.
func CloseDB(db *sqlx.DB) {
	if db != nil {
		zlog.S.Debugf("Closing DB...")
		err := db.Close()
		if err != nil {
			zlog.S.Warnf("Problem closing DB: %v", err)
		}
	}
}

// CloseConn closes the specified DB connection and logs any errors.
func CloseConn(conn *sqlx.Conn) {
	if conn != nil {
		zlog.S.Debugf("Closing Connection...")
		err := conn.Close()
		if err != nil {
			zlog.S.Warnf("Problem closing DB connection: %v", err)
		}
	}
}
