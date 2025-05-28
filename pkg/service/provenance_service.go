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
	"errors"
	"fmt"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
	"strings"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jmoiron/sqlx"
	common "github.com/scanoss/papi/api/commonv2"
	pb "github.com/scanoss/papi/api/geoprovenancev2"
	myconfig "scanoss.com/provenance/pkg/config"
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
	// Make sure we have Provenance data to query
	reqPurls := request.GetPurls()
	if len(reqPurls) == 0 {
		statusResp := common.StatusResponse{Status: common.StatusCode_FAILED, Message: "No purls in request data supplied"}
		return &pb.ContributorResponse{Status: &statusResp}, errors.New("no purl data supplied")
	}
	dtoRequest, err := convertProvenanceInput(s, request) // Convert to internal DTO for processing
	if err != nil {
		statusResp := common.StatusResponse{Status: common.StatusCode_FAILED, Message: "Problem parsing Provenance input data"}
		return &pb.ContributorResponse{Status: &statusResp}, errors.New("problem parsing Provenance input data")
	}
	conn, err := p.db.Connx(ctx) // Get a connection from the pool
	if err != nil {
		s.Errorf("Failed to get a database connection from the pool: %v", err)
		statusResp := common.StatusResponse{Status: common.StatusCode_FAILED, Message: "Failed to get database pool connection"}
		return &pb.ContributorResponse{Status: &statusResp}, errors.New("problem getting database pool connection")
	}
	defer closeDbConnection(conn)
	// Search the KB for information about each Provenance
	provUc := usecase.NewProvenance(ctx, conn, s)
	dtoProv, summary, err := provUc.GetProvenance(dtoRequest)

	if err != nil {
		s.Errorf("Failed to get provenance: %v", err)
		statusResp := common.StatusResponse{Status: common.StatusCode_FAILED, Message: "Problems encountered extracting Provenance data"}
		return &pb.ContributorResponse{Status: &statusResp}, nil
	}
	provResponse, err := convertProvenanceOutput(s, dtoProv) // Convert the internal data into a response object
	if err != nil {
		s.Errorf("Failed to covnert parsed dependencies: %v", err)
		statusResp := common.StatusResponse{Status: common.StatusCode_FAILED, Message: "Problems encountered extracting Provenance data"}
		return &pb.ContributorResponse{Status: &statusResp}, nil
	}
	// Set the status and respond with the data

	statusResp := common.StatusResponse{Status: common.StatusCode_SUCCESS, Message: "Success"}
	var messages []string
	if len(summary.PurlsFailedToParse) > 0 {
		messages = append(messages, fmt.Sprintf("Failed to parse: %s", strings.Join(summary.PurlsFailedToParse, ", ")))
		statusResp.Status = common.StatusCode_SUCCEEDED_WITH_WARNINGS
	}
	if len(summary.PurlsWOInfo) > 0 {
		messages = append(messages, fmt.Sprintf("Can't find information for: %s", strings.Join(summary.PurlsWOInfo, ", ")))
		statusResp.Status = common.StatusCode_SUCCEEDED_WITH_WARNINGS
	}
	if len(summary.PurlsTooMuchData) > 0 {
		messages = append(messages, fmt.Sprintf("Too many contributors for: %s", strings.Join(summary.PurlsTooMuchData, ", ")))
		statusResp.Status = common.StatusCode_SUCCEEDED_WITH_WARNINGS
	}
	if len(messages) == 0 {
		statusResp.Message = "Success"
	} else {
		statusResp.Message = strings.Join(messages, ";")
	}

	return &pb.ContributorResponse{Purls: provResponse.Purls, Status: &statusResp}, nil
}

func (p provenanceServer) GetComponentOrigin(ctx context.Context, request *common.PurlRequest) (*pb.OriginResponse, error) {
	s := ctxzap.Extract(ctx).Sugar()
	// Make sure we have Provenance data to query
	reqPurls := request.GetPurls()
	if len(reqPurls) == 0 {
		statusResp := common.StatusResponse{Status: common.StatusCode_FAILED, Message: "No purls in request data supplied"}
		return &pb.OriginResponse{Status: &statusResp}, errors.New("no purl data supplied")
	}
	dtoRequest, err := convertProvenanceInput(s, request) // Convert to internal DTO for processing
	if err != nil {
		statusResp := common.StatusResponse{Status: common.StatusCode_FAILED, Message: "Problem parsing Provenance input data"}
		return &pb.OriginResponse{Status: &statusResp}, errors.New("problem parsing Provenance input data")
	}
	conn, err := p.db.Connx(ctx) // Get a connection from the pool
	if err != nil {
		s.Errorf("Failed to get a database connection from the pool: %v", err)
		statusResp := common.StatusResponse{Status: common.StatusCode_FAILED, Message: "Failed to get database pool connection"}
		return &pb.OriginResponse{Status: &statusResp}, errors.New("problem getting database pool connection")
	}
	defer closeDbConnection(conn)
	// Search the KB for information about each Provenance
	provUc := usecase.NewOrigin(ctx, conn)
	dtoProv, summary, err := provUc.GetOrigin(dtoRequest)

	if err != nil {
		s.Errorf("Failed to get provenance: %v", err)
		statusResp := common.StatusResponse{Status: common.StatusCode_FAILED, Message: "Problems encountered extracting Provenance data"}
		return &pb.OriginResponse{Status: &statusResp}, nil
	}
	provResponse, err := convertOriginOutput(s, dtoProv) // Convert the internal data into a response object
	if err != nil {
		s.Errorf("Failed to covnert parsed dependencies: %v", err)
		statusResp := common.StatusResponse{Status: common.StatusCode_FAILED, Message: "Problems encountered extracting Provenance data"}
		return &pb.OriginResponse{Status: &statusResp}, nil
	}
	_ = provResponse
	// Set the status and respond with the data

	statusResp := common.StatusResponse{Status: common.StatusCode_SUCCESS, Message: "Success"}
	var messages []string
	if len(summary.PurlsFailedToParse) > 0 {
		messages = append(messages, fmt.Sprintf("Failed to parse: %s", strings.Join(summary.PurlsFailedToParse, ", ")))
		statusResp.Status = common.StatusCode_SUCCEEDED_WITH_WARNINGS
	}
	if len(summary.PurlsTooMuchData) > 0 {
		messages = append(messages, fmt.Sprintf("Too many contributors for: %s", strings.Join(summary.PurlsTooMuchData, ", ")))
		statusResp.Status = common.StatusCode_SUCCEEDED_WITH_WARNINGS
	}
	if len(messages) == 0 {
		statusResp.Message = "Success"
	} else {
		statusResp.Message = strings.Join(messages, ";")
	}

	return &pb.OriginResponse{Purls: provResponse.Purls, Status: &statusResp}, nil
}

// closeDbConnection closes the specified database connection
func closeDbConnection(conn *sqlx.Conn) {
	err := conn.Close()
	if err != nil {
		zlog.S.Warnf("Warning: Problem closing database connection: %v", err)
	}
}
