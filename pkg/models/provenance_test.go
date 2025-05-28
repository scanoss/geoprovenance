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
	"testing"

	_ "modernc.org/sqlite"
	zlog "scanoss.com/provenance/pkg/logger"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jmoiron/sqlx"
)

func TestContributorProvenance(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := context.Background()
	ctx = ctxzap.ToContext(ctx, zlog.L)
	s := ctxzap.Extract(ctx).Sugar()
	_ = s
	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer CloseDB(db)
	ctx = ctxzap.ToContext(ctx, zlog.L)

	err = LoadTestSqlData(db, nil, nil)

	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	conn, err := db.Connx(ctx)

	if err != nil {
		t.Fatalf("an error '%s' was not expected when loading test data", err)
	}

	//CloseConn(conn)
	cProvModel := NewProvenanceModel(ctx, conn)
	purlsNames := []string{"torvalds/uemacs", "scanoss/engine"}
	list, errq := cProvModel.GetProvenanceByPurlNames(purlsNames)
	if errq != nil {
		t.Logf("unexpected error on model request  %+v\n", errq)
	} else {
		if len(list) == 0 {
			t.Log("Expected to get at least one result\n")
		}
	}
	t.Logf("%+v", list)

}
