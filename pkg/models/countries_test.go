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
	"context"
	"fmt"
	_ "modernc.org/sqlite"
	zlog "scanoss.com/provenance/pkg/logger"
	"testing"

	"github.com/jmoiron/sqlx"
	"math/rand/v2"
)

func TestCountryLookUp(t *testing.T) {
	ctx := context.Background()
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer CloseDB(db)
	conn, err := db.Connx(ctx) // Get a connection from the pool
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer CloseConn(conn)
	err = loadTestSqlDataFiles(db, ctx, conn, []string{"../models/tests/countries.sql"})
	if err != nil {
		t.Fatalf("failed to load SQL test data: %v", err)
	}
	countryModel := NewCountryMapModel(ctx, conn)
	countriesToPick := []string{"Afghanistan", "Albania", "Algeria", "Andorra", "Angola", "Antigua and Barbuda", "Argentina", "Armenia", "Australia", "Austria"}

	randomIndex := rand.IntN(len(countriesToPick))
	dbPK := randomIndex + 1

	randomElement := countriesToPick[randomIndex]

	var countryName = randomElement
	fmt.Printf("Searching for Country: %v\n", countryName)
	gotName, err := countryModel.GetCountryById(dbPK)

	if err != nil {
		t.Errorf("Countries model error = %v", err)
	}
	if gotName != randomElement {
		t.Errorf("Country does not match. Expected %s - Received %s\n", randomElement, gotName)
	}
	fmt.Printf("Country: %#v\n", randomElement)
}
