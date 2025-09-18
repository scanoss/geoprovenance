package service

import (
	"context"
	"fmt"
	common "github.com/scanoss/papi/api/commonv2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"net/http"
	"scanoss.com/provenance/pkg/errors"
	"scanoss.com/provenance/pkg/models"
	"strings"
)

const (
	ResponseMessageSuccess = "Success"
	ResponseMessageError   = "Internal error occurred"
)

// buildErrorMessages creates error messages for each type of PURL failure.
func buildErrorMessages(summary models.QuerySummary) []string {
	var messages []string

	if len(summary.PurlsFailedToParse) > 0 {
		messages = append(messages, fmt.Sprintf("Failed to parse %d purl(s):%s",
			len(summary.PurlsFailedToParse), strings.Join(summary.PurlsFailedToParse, ",")))
	}

	if len(summary.PurlsNotFound) > 0 {
		messages = append(messages, fmt.Sprintf("Can't find %d purl(s):%s",
			len(summary.PurlsNotFound), strings.Join(summary.PurlsNotFound, ",")))
	}

	if len(summary.PurlsWOInfo) > 0 {
		messages = append(messages, fmt.Sprintf("Can't find information for %d purl(s):%s",
			len(summary.PurlsWOInfo), strings.Join(summary.PurlsWOInfo, ",")))
	}

	if len(summary.PurlsTooMuchData) > 0 {
		messages = append(messages, fmt.Sprintf("Too many contributors for: %s", strings.Join(summary.PurlsTooMuchData, ", ")))
	}
	return messages
}

// determineStatusAndHTTPCode analyzes the PURL processing results and determines the appropriate
// status code and HTTP code based on the success/failure ratios.
// Returns common.StatusCode, HTTP status code, and error.
func determineStatusAndHTTPCode(s *zap.SugaredLogger, summary models.QuerySummary) (common.StatusResponse, int, error) {
	var messages = buildErrorMessages(summary)
	responseMessage := ResponseMessageSuccess
	if len(messages) > 0 {
		responseMessage = strings.Join(messages, " | ")
	}
	// Calculate failure statistics
	totalFailedToParse := len(summary.PurlsFailedToParse)
	totalNotFound := len(summary.PurlsNotFound)
	totalWOInfo := len(summary.PurlsWOInfo)
	totalFailed := totalFailedToParse + totalNotFound + totalWOInfo
	totalSuccessful := summary.TotalPurls - totalFailed
	totalPurls := summary.TotalPurls

	// Log processing summary
	s.Debugf("PURL Summary - Total: %d, Successful: %d, Failed to parse: %d, Not found: %d, No info: %d",
		summary.TotalPurls, totalSuccessful, totalFailedToParse, totalNotFound, totalWOInfo)

	switch {
	case totalFailed == 0:
		// All PURLs succeeded
		return common.StatusResponse{
			Message: responseMessage,
			Status:  common.StatusCode_SUCCESS,
		}, http.StatusOK, nil

	case totalSuccessful == 0:
		// All PURLs failed - determine HTTP code by failure type priority
		if totalFailedToParse > 0 && totalFailedToParse >= totalPurls {
			return common.StatusResponse{}, http.StatusBadRequest, errors.NewBadRequestError(responseMessage, nil)
		}
		return common.StatusResponse{}, http.StatusNotFound, errors.NewNotFoundError(responseMessage)

	case len(summary.PurlsTooMuchData) > 0:
		return common.StatusResponse{
			Message: responseMessage,
			Status:  common.StatusCode_SUCCEEDED_WITH_WARNINGS,
		}, http.StatusOK, nil

	default:
		// Mixed results: some succeeded, some failed
		return common.StatusResponse{
			Message: ResponseMessageSuccess,
			Status:  common.StatusCode_SUCCESS,
		}, http.StatusOK, nil
	}
}

// buildStatusResponse constructs a StatusResponse based on PURL processing results and sets appropriate HTTP status codes.
func buildStatusResponse(ctx context.Context, s *zap.SugaredLogger, summary models.QuerySummary) (*common.StatusResponse, error) {
	statusResponse, httpStatusCode, err := determineStatusAndHTTPCode(s, summary)
	if err != nil {
		return &common.StatusResponse{}, err
	}
	setHTTPCodeOnTrailer(ctx, s, httpStatusCode)
	return &statusResponse, err
}

// setHTTPCodeOnTrailer sets the HTTP status code in the gRPC trailer metadata.
// This allows clients to determine the appropriate HTTP response code for the request.
func setHTTPCodeOnTrailer(ctx context.Context, s *zap.SugaredLogger, code int) {
	err := grpc.SetTrailer(ctx, metadata.Pairs("x-http-code", fmt.Sprintf("%d", code)))
	if err != nil {
		s.Errorf("error setting x-http-code to trailer: %v\n", err)
	}
}

// resolveResponseStatus safely extracts status from a response interface, providing a fallback
// when the response is nil. This prevents nil pointer dereferences in error handling.
// This method is mainly used by single component calls that delegate to batch calls and need
// to safely handle potentially nil responses from the underlying batch operations.
func resolveResponseStatus(response interface{}) *common.StatusResponse {
	defaultStatus := &common.StatusResponse{
		Status:  common.StatusCode_FAILED,
		Message: ResponseMessageError,
	}
	if response == nil {
		return defaultStatus
	}

	if statusResp, ok := response.(interface{ GetStatus() *common.StatusResponse }); ok {
		if status := statusResp.GetStatus(); status != nil {
			return status
		}
	}
	// Fallback if status cannot be extracted
	return defaultStatus
}
