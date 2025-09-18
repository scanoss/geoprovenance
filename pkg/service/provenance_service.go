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

// Package service implements the gRPC service endpoints
package service

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jmoiron/sqlx"
	common "github.com/scanoss/papi/api/commonv2"
	pb "github.com/scanoss/papi/api/geoprovenancev2"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
	myconfig "scanoss.com/provenance/pkg/config"
	se "scanoss.com/provenance/pkg/errors"
	"scanoss.com/provenance/pkg/usecase"
)

type provenanceServer struct {
	pb.GeoProvenanceServer
	db     *sqlx.DB
	config *myconfig.ServerConfig
}

// NewProvenanceServer creates a new instance of Provenance Server
func NewProvenanceServer(db *sqlx.DB, config *myconfig.ServerConfig) pb.GeoProvenanceServer {
	return &provenanceServer{db: db, config: config}
}

// Echo sends back the same message received
func (p provenanceServer) Echo(ctx context.Context, request *common.EchoRequest) (*common.EchoResponse, error) {
	s := ctxzap.Extract(ctx).Sugar()
	s.Infof("Received (%v): %v", ctx, request.GetMessage())
	return &common.EchoResponse{Message: request.GetMessage()}, nil
}

func (p provenanceServer) GetComponentContributors(ctx context.Context, request *common.PurlRequest) (*pb.ContributorResponse, error) {
	s := ctxzap.Extract(ctx).Sugar()
	dtoRequest, err := convertProvenanceInput(request) // Convert to internal DTO for processing
	if err != nil {
		return &pb.ContributorResponse{Status: se.HandleServiceError(ctx, s, err)}, nil
	}
	conn, err := p.db.Connx(ctx) // Get a connection from the pool
	if err != nil {
		s.Errorf("Failed to get a database connection from the pool: %v", err)
		return &pb.ContributorResponse{Status: se.HandleServiceError(ctx, s, se.NewInternalError("problem getting database pool connection", err))}, nil
	}
	defer closeDbConnection(conn)
	// Search the KB for information about each Provenance
	provUc := usecase.NewProvenance(ctx, conn, s)
	dtoProv, summary, err := provUc.GetProvenance(dtoRequest)
	if err != nil {
		s.Errorf("Failed to get provenance: %v", err)
		return &pb.ContributorResponse{Status: se.HandleServiceError(ctx, s, se.NewNotFoundError("Problems encountered extracting Provenance data"))}, nil
	}

	provResponse, err := convertProvenanceOutput(s, dtoProv) // Convert the internal data into a response object
	if err != nil {
		return &pb.ContributorResponse{Status: se.HandleServiceError(ctx, s, se.NewInternalError("Problems encountered extracting Provenance data", err))}, nil
	}

	statusResp, err := buildStatusResponse(ctx, s, summary)
	// Set the status and respond with the data
	if err != nil {
		return &pb.ContributorResponse{Status: se.HandleServiceError(ctx, s, err)}, nil
	}
	return &pb.ContributorResponse{Purls: provResponse.Purls, Status: statusResp}, nil
}

func (p provenanceServer) GetComponentOrigin(ctx context.Context, request *common.PurlRequest) (*pb.OriginResponse, error) {
	s := ctxzap.Extract(ctx).Sugar()
	// Make sure we have Provenance data to query
	dtoRequest, err := convertProvenanceInput(request) // Convert to internal DTO for processing
	if err != nil {
		return &pb.OriginResponse{Status: se.HandleServiceError(ctx, s, err)}, nil
	}

	conn, err := p.db.Connx(ctx) // Get a connection from the pool
	if err != nil {
		s.Errorf("Failed to get a database connection from the pool: %v", err)
		return &pb.OriginResponse{Status: se.HandleServiceError(ctx, s, se.NewInternalError("problem getting database pool connection", err))}, nil
	}
	defer closeDbConnection(conn)
	// Search the KB for information about each Provenance
	provUc := usecase.NewOrigin(ctx, conn)
	dtoProv, summary, err := provUc.GetOrigin(dtoRequest)

	if err != nil {
		s.Errorf("Failed to get provenance: %v", err)
		return &pb.OriginResponse{Status: se.HandleServiceError(ctx, s, se.NewNotFoundError("Problems encountered extracting Provenance data"))}, nil
	}
	provResponse, err := convertOriginOutput(s, dtoProv) // Convert the internal data into a response object
	if err != nil {
		return &pb.OriginResponse{Status: se.HandleServiceError(ctx, s, se.NewInternalError("Problems encountered extracting Provenance data", err))}, nil
	}
	_ = provResponse
	// Set the status and respond with the data

	statusResp, err := buildStatusResponse(ctx, s, summary)
	// Set the status and respond with the data
	if err != nil {
		return &pb.OriginResponse{Status: se.HandleServiceError(ctx, s, err)}, nil
	}
	return &pb.OriginResponse{Purls: provResponse.Purls, Status: statusResp}, nil
}

// closeDbConnection closes the specified database connection
func closeDbConnection(conn *sqlx.Conn) {
	err := conn.Close()
	if err != nil {
		zlog.S.Warnf("Warning: Problem closing database connection: %v", err)
	}
}
