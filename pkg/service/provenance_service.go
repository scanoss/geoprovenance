// SPDX-License-Identifier: GPL-2.0-or-later
/*
 * Copyright (C) 2018-2025 SCANOSS.COM
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

// Package service implements the gRPC service endpoints for the geo provenance service.
package service

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jmoiron/sqlx"
	"github.com/scanoss/go-component-helper/componenthelper"
	common "github.com/scanoss/papi/api/commonv2"
	pb "github.com/scanoss/papi/api/geoprovenancev2"
	"go.uber.org/zap"
	myconfig "scanoss.com/provenance/pkg/config"
	se "scanoss.com/provenance/pkg/errors"
	"scanoss.com/provenance/pkg/usecase"
)

type provenanceServer struct {
	pb.GeoProvenanceServer
	config            *myconfig.ServerConfig
	provenanceUseCase *usecase.ProvenanceUseCase
	originUseCase     *usecase.OriginUseCase
}

func NewProvenanceServer(db *sqlx.DB, config *myconfig.ServerConfig) pb.GeoProvenanceServer {
	return &provenanceServer{
		config:            config,
		provenanceUseCase: usecase.NewProvenance(db),
		originUseCase:     usecase.NewOrigin(db),
	}
}

// runPipeline is the shared request-handling template for every gRPC endpoint in this
// service. Every endpoint does the same four things — extract the logger, translate the
// incoming request into component DTOs, invoke a use case, and attach a status response —
// so those steps live here once. The three callbacks capture what actually varies between
// endpoints:
//
//   - convert: how to pull []ComponentDTO out of this endpoint's request type (Req)
//   - handle:  which use case to run and how to shape its output (Data)
//   - build:   how to assemble the final protobuf response (Resp) from data + status
//
// Errors from either the use case or status-response assembly are funnelled through
// se.HandleServiceError, which maps them to a *common.StatusResponse; build is then
// invoked with a zero-value Data so each endpoint can still return a well-formed Resp
// carrying the error status.
func runPipeline[Req any, Data any, Resp any](
	ctx context.Context,
	req Req,
	convert func(Req) ([]componenthelper.ComponentDTO, error),
	handle func(context.Context, *zap.SugaredLogger, []componenthelper.ComponentDTO) (Data, error),
	build func(Data, *common.StatusResponse) Resp,
) Resp {
	s := ctxzap.Extract(ctx).Sugar()
	var zero Data
	dto, err := convert(req)
	if err != nil {
		return build(zero, se.HandleServiceError(ctx, s, err))
	}
	data, err := handle(ctx, s, dto)
	if err != nil {
		return build(zero, se.HandleServiceError(ctx, s, err))
	}
	status, err := buildStatusResponse(ctx, s)
	if err != nil {
		return build(zero, se.HandleServiceError(ctx, s, err))
	}
	return build(data, status)
}

func (p provenanceServer) Echo(ctx context.Context, request *common.EchoRequest) (*common.EchoResponse, error) {
	s := ctxzap.Extract(ctx).Sugar()
	s.Infof("Received (%v): %v", ctx, request.GetMessage())
	return &common.EchoResponse{Message: request.GetMessage()}, nil
}

func (p provenanceServer) GetComponentContributors(ctx context.Context, request *common.PurlRequest) (*pb.ContributorResponse, error) {
	return runPipeline(
		ctx,
		request,
		convertProvenanceInput,
		func(ctx context.Context, s *zap.SugaredLogger, dto []componenthelper.ComponentDTO) (*pb.ContributorResponse, error) {
			data, err := p.provenanceUseCase.GetProvenance(ctx, s, dto)
			if err != nil {
				return nil, err
			}
			return convertProvenanceOutput(s, data)
		},
		func(resp *pb.ContributorResponse, status *common.StatusResponse) *pb.ContributorResponse {
			if resp == nil {
				resp = &pb.ContributorResponse{}
			}
			resp.Status = status
			return resp
		},
	), nil
}

func (p provenanceServer) GetCountryContributorsByComponents(ctx context.Context, request *common.ComponentsRequest) (*pb.ComponentsContributorResponse, error) {
	return runPipeline(
		ctx,
		request,
		componentsRequestToDTO,
		func(ctx context.Context, s *zap.SugaredLogger, dto []componenthelper.ComponentDTO) (*pb.ComponentsContributorResponse, error) {
			data, err := p.provenanceUseCase.GetProvenance(ctx, s, dto)
			if err != nil {
				return nil, err
			}
			return toComponentsContributorResponse(data)
		},
		func(resp *pb.ComponentsContributorResponse, status *common.StatusResponse) *pb.ComponentsContributorResponse {
			if resp == nil {
				resp = &pb.ComponentsContributorResponse{}
			}
			resp.Status = status
			return resp
		},
	), nil
}

func (p provenanceServer) GetCountryContributorsByComponent(ctx context.Context, request *common.ComponentRequest) (*pb.ComponentContributorResponse, error) {
	return runPipeline(
		ctx,
		request,
		componentRequestToDTO,
		func(ctx context.Context, s *zap.SugaredLogger, dto []componenthelper.ComponentDTO) (*pb.ComponentContributorResponse, error) {
			data, err := p.provenanceUseCase.GetProvenance(ctx, s, dto)
			if err != nil {
				return nil, err
			}
			return toComponentContributorResponse(data)
		},
		func(resp *pb.ComponentContributorResponse, status *common.StatusResponse) *pb.ComponentContributorResponse {
			if resp == nil {
				resp = &pb.ComponentContributorResponse{}
			}
			resp.Status = status
			return resp
		},
	), nil
}

func (p provenanceServer) GetComponentOrigin(ctx context.Context, request *common.PurlRequest) (*pb.OriginResponse, error) {
	return runPipeline(
		ctx,
		request,
		convertProvenanceInput,
		func(ctx context.Context, s *zap.SugaredLogger, dto []componenthelper.ComponentDTO) (*pb.OriginResponse, error) {
			data, err := p.originUseCase.GetOrigin(ctx, s, dto)
			if err != nil {
				return nil, err
			}
			return convertOriginOutput(s, data)
		},
		func(resp *pb.OriginResponse, status *common.StatusResponse) *pb.OriginResponse {
			if resp == nil {
				resp = &pb.OriginResponse{}
			}
			resp.Status = status
			return resp
		},
	), nil
}

func (p provenanceServer) GetOriginByComponents(ctx context.Context, request *common.ComponentsRequest) (*pb.ComponentsOriginResponse, error) {
	return runPipeline(ctx, request, componentsRequestToDTO,
		func(ctx context.Context, s *zap.SugaredLogger, dto []componenthelper.ComponentDTO) (*pb.ComponentsOriginResponse, error) {
			data, err := p.originUseCase.GetOrigin(ctx, s, dto)
			if err != nil {
				return nil, err
			}
			return toComponentsOriginResponse(data)
		},
		func(resp *pb.ComponentsOriginResponse, status *common.StatusResponse) *pb.ComponentsOriginResponse {
			if resp == nil {
				resp = &pb.ComponentsOriginResponse{}
			}
			resp.Status = status
			return resp
		},
	), nil
}

func (p provenanceServer) GetOriginByComponent(ctx context.Context, request *common.ComponentRequest) (*pb.ComponentOriginResponse, error) {
	return runPipeline(ctx, request, componentRequestToDTO,
		func(ctx context.Context, s *zap.SugaredLogger, dto []componenthelper.ComponentDTO) (*pb.ComponentOriginResponse, error) {
			data, err := p.originUseCase.GetOrigin(ctx, s, dto)
			if err != nil {
				return nil, err
			}
			return toComponentOriginResponse(data)
		},
		func(resp *pb.ComponentOriginResponse, status *common.StatusResponse) *pb.ComponentOriginResponse {
			if resp == nil {
				resp = &pb.ComponentOriginResponse{}
			}
			resp.Status = status
			return resp
		},
	), nil
}
