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
// It provides gRPC handlers for component contributor and origin information retrieval,
// including request validation, error handling, and response formatting.
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

// provenanceServer implements the GeoProvenanceServer interface and serves as the main
// gRPC service handler for geo provenance operations. It encapsulates the business logic
// use cases and server configuration needed to process provenance and origin requests.
type provenanceServer struct {
	pb.GeoProvenanceServer
	// config holds the server configuration including database settings and service parameters
	config *myconfig.ServerConfig
	// provenanceUseCase handles business logic for component contributor information
	provenanceUseCase *usecase.ProvenanceUseCase
	// originUseCase handles business logic for component origin information
	originUseCase *usecase.OriginUseCase
}

// NewProvenanceServer creates a new instance of the Provenance Server with the provided
// database connection and server configuration. It initializes both provenance and origin
// use cases that handle the core business logic for geo provenance operations.
//
// Parameters:
//   - db: Database connection pool for data access operations
//   - config: Server configuration containing service settings and parameters
//
// Returns:
//   - pb.GeoProvenanceServer: Configured gRPC service instance ready to handle requests
func NewProvenanceServer(db *sqlx.DB, config *myconfig.ServerConfig) pb.GeoProvenanceServer {
	return &provenanceServer{
		config:            config,
		provenanceUseCase: usecase.NewProvenance(db),
		originUseCase:     usecase.NewOrigin(db),
	}
}

// UseCaseHandler defines the core handler function type for processing component requests.
// It abstracts the common pattern of taking component DTOs and returning processed data
// along with a query summary for status reporting.
//
// Parameters:
//   - ctx: Request context for cancellation and timeout handling
//   - s: Structured logger for request tracing and debugging
//   - dto: Array of component DTOs containing PURL and requirement information
//
// Returns:
//   - interface{}: Processed data (type varies by specific handler implementation)
//   - models.QuerySummary: Summary of query execution including success/failure counts
//   - error: Any error encountered during processing
type UseCaseHandler func(ctx context.Context, s *zap.SugaredLogger, dto []componenthelper.ComponentDTO) (interface{}, error)

// ResponseBuilder defines a generic function type for building typed responses from
// processed data and status information. It provides type safety while avoiding
// the overhead of runtime type assertions in the response building process.
//
// Type Parameters:
//   - T: The specific response type to be built (e.g., *pb.ContributorResponse)
//
// Parameters:
//   - data: Processed data from the request handler (may be nil on errors)
//   - status: Status response containing success/failure information and messages
//
// Returns:
//   - T: Fully constructed response of the specified type
type ResponseBuilder[T any] func(data interface{}, status *common.StatusResponse) T

type RequestConverter[R any] func(R) []componenthelper.ComponentDTO

// executeRequestPipeline provides a unified abstraction for handling gRPC requests with
// common concerns like input validation, error handling, and response building.
// This function encapsulates the standard request processing pipeline used across
// different service endpoints.
//
// Type Parameters:
//   - T: The response type to be returned (e.g., *pb.ContributorResponse)
//   - R: The request type to be processed (e.g., *common.PurlRequest, *common.ComponentsRequest)
//
// Parameters:
//   - ctx: Request context for cancellation and timeout handling
//   - req: Request containing component information to process
//   - converter: Function that converts the request to internal DTOs
//   - useCaseHandler: Business logic handler that processes the validated input
//   - responseBuilder: Function that constructs the final typed response
//
// Returns:
//   - T: Fully constructed response with either success data or error status
//
// The function follows this processing pipeline:
// 1. Extract logger from context
// 2. Validate and convert input to internal DTOs using the provided converter
// 3. Execute business logic handler
// 4. Build status response from query summary
// 5. Construct and return typed response
func executeRequestPipeline[T any, R any](
	ctx context.Context,
	req R,
	converter RequestConverter[R],
	useCaseHandler UseCaseHandler,
	responseBuilder ResponseBuilder[T],
) T {
	s := ctxzap.Extract(ctx).Sugar()
	// Input validation
	dto := converter(req)

	// Use case call
	data, err := useCaseHandler(ctx, s, dto)
	if err != nil {
		return responseBuilder(nil, se.HandleServiceError(ctx, s, err))
	}
	status, err := buildStatusResponse(ctx, s)
	if err != nil {
		return responseBuilder(nil, se.HandleServiceError(ctx, s, err))
	}
	return responseBuilder(data, status)
}

// handleLegacyRequest provides a unified abstraction for handling gRPC requests with
// common concerns like input validation, error handling, and response building.
// This function encapsulates the standard request processing pipeline used across
// different service endpoints.
//
// Type Parameters:
//   - T: The response type to be returned (e.g., *pb.ContributorResponse)
//
// Parameters:
//   - ctx: Request context for cancellation and timeout handling
//   - req: PURL request containing component information to process
//   - useCaseHandler: Business logic handler that processes the validated input
//   - responseBuilder: Function that constructs the final typed response
//
// Returns:
//   - T: Fully constructed response with either success data or error status
//
// The function follows this processing pipeline:
// 1. Extract logger from context
// 2. Validate and convert input to internal DTOs
// 3. Execute business logic handler
// 4. Build status response from query summary
// 5. Construct and return typed response
func handleLegacyRequest[T any](
	ctx context.Context,
	req *common.PurlRequest, //nolint:staticcheck
	useCaseHandler UseCaseHandler,
	responseBuilder ResponseBuilder[T],
) T {
	return executeRequestPipeline(ctx, req, convertProvenanceInput, useCaseHandler, responseBuilder)
}

// handleComponentRequest provides a unified abstraction for handling gRPC requests with
// common concerns like input validation, error handling, and response building.
// This function encapsulates the standard request processing pipeline used across
// different service endpoints.
//
// Type Parameters:
//   - T: The response type to be returned (e.g., *pb.ContributorResponse)
//
// Parameters:
//   - ctx: Request context for cancellation and timeout handling
//   - req: components request containing component information to process
//   - useCaseHandler: Business logic handler that processes the validated input
//   - responseBuilder: Function that constructs the final typed response
//
// Returns:
//   - T: Fully constructed response with either success data or error status
//
// The function follows this processing pipeline:
// 1. Extract logger from context
// 2. Validate and convert input to internal DTOs
// 3. Execute business logic handler
// 4. Build status response from query summary
// 5. Construct and return typed response
func handleComponentsRequest[T any](
	ctx context.Context,
	req *common.ComponentsRequest,
	useCaseHandler UseCaseHandler,
	responseBuilder ResponseBuilder[T],
) T {
	return executeRequestPipeline(ctx, req, componentsRequestToDTO, useCaseHandler, responseBuilder)
}

// handleComponentRequest provides a unified abstraction for handling gRPC requests with
// common concerns like input validation, error handling, and response building.
// This function encapsulates a component request processing pipeline used across
// different service endpoints.
//
// Type Parameters:
//   - T: The response type to be returned (e.g., *pb.ComponentContributorRespons)
//
// Parameters:
//   - ctx: Request context for cancellation and timeout handling
//   - req: component request containing component information to process
//   - useCaseHandler: Business logic handler that processes the validated input
//   - responseBuilder: Function that constructs the final typed response
//
// Returns:
//   - T: Fully constructed response with either success data or error status
//
// The function follows this processing pipeline:
// 1. Extract logger from context
// 2. Validate and convert input to internal DTOs
// 3. Execute business logic handler
// 4. Build status response from query summary
// 5. Construct and return typed response
func handleComponentRequest[T any](
	ctx context.Context,
	req *common.ComponentRequest,
	useCaseHandler UseCaseHandler,
	responseBuilder ResponseBuilder[T],
) T {
	return executeRequestPipeline(ctx, req, componentRequestToDTO, useCaseHandler, responseBuilder)
}

// Echo implements a simple echo service for health checks and connectivity testing.
// It receives a message and returns the same message back to the client, along with
// logging the received message for debugging purposes.
//
// Parameters:
//   - ctx: Request context for cancellation and timeout handling
//   - request: Echo request containing the message to be echoed back
//
// Returns:
//   - *common.EchoResponse: Response containing the same message as received
//   - error: Always nil for this implementation
func (p provenanceServer) Echo(ctx context.Context, request *common.EchoRequest) (*common.EchoResponse, error) {
	s := ctxzap.Extract(ctx).Sugar()
	s.Infof("Received (%v): %v", ctx, request.GetMessage())
	return &common.EchoResponse{Message: request.GetMessage()}, nil
}

// GetComponentContributors retrieves contributor information for the specified components.
// This endpoint processes PURL (Package URL) requests to identify and return information
// about contributors associated with the requested software components.
//
// Parameters:
//   - ctx: Request context for cancellation and timeout handling
//   - request: PURL request containing components to analyze for contributor information
//
// Returns:
//   - *pb.ContributorResponse: Response containing contributor data and processing status
//   - error: Always nil; errors are encoded in the response status
//
// The function uses the handleLegacyRequest abstraction to:
// 1. Validate input PURLs and convert to internal DTOs
// 2. Execute provenance use case to retrieve contributor data
// 3. Convert output to protobuf format
// 4. Build appropriate status response based on processing results
func (p provenanceServer) GetComponentContributors(ctx context.Context, request *common.PurlRequest) (*pb.ContributorResponse, error) { //nolint:staticcheck
	result := handleLegacyRequest[*pb.ContributorResponse](ctx, request, //nolint:staticcheck
		// Component contributors use case call
		func(ctx context.Context, s *zap.SugaredLogger, dto []componenthelper.ComponentDTO) (interface{}, error) {

			data, err := p.provenanceUseCase.GetProvenance(ctx, s, dto)
			if err != nil {
				return nil, err
			}
			response, err := convertProvenanceOutput(s, data)
			return response, err
		},
		// Response mapping - type-safe and clear
		func(data interface{}, status *common.StatusResponse) *pb.ContributorResponse { //nolint:staticcheck
			resp := &pb.ContributorResponse{Status: status}                            //nolint:staticcheck
			if provData, ok := data.(*pb.ContributorResponse); ok && provData != nil { //nolint:staticcheck
				resp.Purls = provData.Purls
			}
			return resp
		},
	)
	return result, nil
}

func (p provenanceServer) GetCountryContributorsByComponents(ctx context.Context, request *common.ComponentsRequest) (*pb.ComponentsContributorResponse, error) {
	result := handleComponentsRequest[*pb.ComponentsContributorResponse](ctx, request,
		// Component contributors use case call
		func(ctx context.Context, s *zap.SugaredLogger, dto []componenthelper.ComponentDTO) (interface{}, error) {
			data, err := p.provenanceUseCase.GetProvenance(ctx, s, dto)
			response, err := toComponentsContributorResponse(data)
			return response, err
		},
		// Set status on response
		func(data interface{}, status *common.StatusResponse) *pb.ComponentsContributorResponse {
			if contributorResponse, ok := data.(*pb.ComponentsContributorResponse); ok && contributorResponse != nil {
				contributorResponse.Status = status
				return contributorResponse
			}
			resp := &pb.ComponentsContributorResponse{Status: status}
			return resp
		},
	)
	return result, nil
}

func (p provenanceServer) GetCountryContributorsByComponent(ctx context.Context, request *common.ComponentRequest) (*pb.ComponentContributorResponse, error) {
	result := handleComponentRequest[*pb.ComponentContributorResponse](ctx, request,
		// Component contributors use case call
		func(ctx context.Context, s *zap.SugaredLogger, dto []componenthelper.ComponentDTO) (interface{}, error) {
			data, err := p.provenanceUseCase.GetProvenance(ctx, s, dto)
			response, err := toComponentContributorResponse(data)
			return response, err
		},

		// Set status on response
		func(data interface{}, status *common.StatusResponse) *pb.ComponentContributorResponse {
			if contributorResponse, ok := data.(*pb.ComponentContributorResponse); ok && contributorResponse != nil {
				contributorResponse.Status = status
				return contributorResponse
			}
			resp := &pb.ComponentContributorResponse{Status: status}
			return resp
		},
	)
	return result, nil
}

// GetComponentOrigin retrieves origin information for the specified components.
// This endpoint processes PURL (Package URL) requests to identify and return information
// about the geographical and organizational origins of the requested software components.
func (p provenanceServer) GetComponentOrigin(ctx context.Context, request *common.PurlRequest) (*pb.OriginResponse, error) { //nolint:staticcheck
	result := handleLegacyRequest[*pb.OriginResponse](ctx, request, //nolint:staticcheck
		// Component contributors use case call
		func(ctx context.Context, s *zap.SugaredLogger, dto []componenthelper.ComponentDTO) (interface{}, error) {
			data, err := p.originUseCase.GetOrigin(ctx, s, dto)
			response, err := convertOriginOutput(s, data)
			return response, err
		},
		// Response mapping - type-safe and clear
		func(data interface{}, status *common.StatusResponse) *pb.OriginResponse { //nolint:staticcheck
			resp := &pb.OriginResponse{Status: status}                            //nolint:staticcheck
			if provData, ok := data.(*pb.OriginResponse); ok && provData != nil { //nolint:staticcheck
				resp.Purls = provData.Purls
			}
			return resp
		},
	)
	return result, nil
}

// GetOriginByComponents retrieves origin information for the specified components.
// This endpoint processes PURL (Package URL) requests to identify and return information
// about the geographical and organizational origins of the requested software components.
func (p provenanceServer) GetOriginByComponents(ctx context.Context, request *common.ComponentsRequest) (*pb.ComponentsOriginResponse, error) {
	result := handleComponentsRequest[*pb.ComponentsOriginResponse](ctx, request,
		// Component contributors use case call
		func(ctx context.Context, s *zap.SugaredLogger, dto []componenthelper.ComponentDTO) (interface{}, error) {
			data, err := p.originUseCase.GetOrigin(ctx, s, dto)
			response, err := toComponentsOriginResponse(data)
			return response, err
		},
		// Response mapping - type-safe and clear
		func(data interface{}, status *common.StatusResponse) *pb.ComponentsOriginResponse {
			if componentsOriginResponse, ok := data.(*pb.ComponentsOriginResponse); ok && componentsOriginResponse != nil {
				componentsOriginResponse.Status = status
				return componentsOriginResponse
			}
			resp := &pb.ComponentsOriginResponse{Status: status}
			return resp
		},
	)
	return result, nil
}

// GetOriginByComponent retrieves origin information for the specified components.
// This endpoint processes PURL (Package URL) requests to identify and return information
// about the geographical and organizational origins of the requested software components.
func (p provenanceServer) GetOriginByComponent(ctx context.Context, request *common.ComponentRequest) (*pb.ComponentOriginResponse, error) {
	result := handleComponentRequest[*pb.ComponentOriginResponse](ctx, request,
		// Component contributors use case call
		func(ctx context.Context, s *zap.SugaredLogger, dto []componenthelper.ComponentDTO) (interface{}, error) {
			data, err := p.originUseCase.GetOrigin(ctx, s, dto)
			response, err := toComponentOriginResponse(data)
			return response, err
		},
		// Response mapping - type-safe and clear
		func(data interface{}, status *common.StatusResponse) *pb.ComponentOriginResponse {
			if componentsOriginResponse, ok := data.(*pb.ComponentOriginResponse); ok && componentsOriginResponse != nil {
				componentsOriginResponse.Status = status
				return componentsOriginResponse
			}
			resp := &pb.ComponentOriginResponse{Status: status}
			return resp
		},
	)
	return result, nil
}
