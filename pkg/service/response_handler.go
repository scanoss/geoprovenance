package service

import (
	"context"
	"fmt"
	"net/http"

	common "github.com/scanoss/papi/api/commonv2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	ResponseMessageSuccess = "Success"
	ResponseMessageError   = "Internal error occurred"
)

// buildStatusResponse constructs a StatusResponse based on PURL processing results and sets appropriate HTTP status codes.
func buildStatusResponse(ctx context.Context, s *zap.SugaredLogger) (*common.StatusResponse, error) {
	err := setHTTPCodeOnTrailer(ctx, s, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &common.StatusResponse{
		Message: ResponseMessageSuccess,
		Status:  common.StatusCode_SUCCESS,
	}, nil
}

// setHTTPCodeOnTrailer sets the HTTP status code in the gRPC trailer metadata.
// This allows clients to determine the appropriate HTTP response code for the request.
func setHTTPCodeOnTrailer(ctx context.Context, s *zap.SugaredLogger, code int) error {
	err := grpc.SetTrailer(ctx, metadata.Pairs("x-http-code", fmt.Sprintf("%d", code)))
	if err != nil {
		s.Errorf("error setting x-http-code to trailer: %v\n", err)
		return err
	}
	return nil
}
