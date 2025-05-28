package models

import (
	"context"
	"testing"

	_ "modernc.org/sqlite"
	zlog "scanoss.com/provenance/pkg/logger"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jmoiron/sqlx"
)

func TestContributorOrigin(t *testing.T) {
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
	list, errq := cProvModel.GetTimeZoneOriginByPurlName(purlsNames[1])
	if errq != nil {
		t.Logf("unexpected error on model request  %+v\n", errq)
	} else {
		if len(list) == 0 {
			t.Log("Expected to get at least one result\n")
		}
	}
	t.Logf("%+v", list)

}
