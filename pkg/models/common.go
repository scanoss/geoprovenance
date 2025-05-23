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
	"io/ioutil"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
	zlog "scanoss.com/provenance/pkg/logger"
)

// loadSqlData Load the specified SQL files into the supplied DB
func loadSqlData(db *sqlx.DB, ctx context.Context, conn *sqlx.Conn, filename string) error {
	fmt.Printf("Loading test data file: %v\n", filename)
	file, err := ioutil.ReadFile(filename)
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

// LoadTestSqlData loads all the required test SQL files
func LoadTestSqlData(db *sqlx.DB, ctx context.Context, conn *sqlx.Conn) error {
	files := []string{
		"../models/tests/countries.sql", "../models/tests/versions.sql", "../models/tests/golang_projects.sql", "../models/tests/vendor_locations.sql", "../models/tests/vendors.sql", "../models/tests/github_contributors.sql"}
	return loadTestSqlDataFiles(db, ctx, conn, files)
}

// loadTestSqlDataFiles loads a list of test SQL files
func loadTestSqlDataFiles(db *sqlx.DB, ctx context.Context, conn *sqlx.Conn, files []string) error {
	for _, file := range files {
		err := loadSqlData(db, ctx, conn, file)
		if err != nil {
			return err
		}
	}

	return nil
}

func Concat(args ...interface{}) (string, error) {
	var result string
	for _, arg := range args {
		if arg != nil {
			result += fmt.Sprint(arg)
		}
	}
	return result, nil
}
func RegisterConcat(db *sqlx.DB, ctx context.Context) {
	conn, err := db.Connx(ctx) // Get a connection from the pool
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	sqliteConn := conn.Raw(func(driverConn interface{}) error {
		if sqliteConn, ok := driverConn.(*sqlite3.SQLiteConn); ok {
			// Register CONCAT function
			err := sqliteConn.RegisterFunc("CONCAT", Concat, true)
			if err != nil {
				return fmt.Errorf("error al registrar la función CONCAT: %w", err)
			}
		} else {
			return fmt.Errorf("Could not connect to SQLite")
		}
		return nil
	})
	if sqliteConn != nil {
		log.Fatal("Error al registrar la función CONCAT:", err)
	}
	_ = sqliteConn
	CloseConn(conn)

}

// CloseDB closes the specified DB and logs any errors
func CloseDB(db *sqlx.DB) {
	if db != nil {
		zlog.S.Debugf("Closing DB...")
		err := db.Close()
		if err != nil {
			zlog.S.Warnf("Problem closing DB: %v", err)
		}
	}
}

// CloseConn closes the specified DB connection and logs any errors
func CloseConn(conn *sqlx.Conn) {
	if conn != nil {
		zlog.S.Debugf("Closing Connection...")
		err := conn.Close()
		if err != nil {
			zlog.S.Warnf("Problem closing DB connection: %v", err)
		}
	}
}

// CloseRows closes the specified DB query row and logs any errors
func CloseRows(rows *sqlx.Rows) {
	if rows != nil {
		zlog.S.Debugf("Closing Rows...")
		err := rows.Close()
		if err != nil {
			zlog.S.Warnf("Problem closing Rows: %v", err)
		}
	}
}
