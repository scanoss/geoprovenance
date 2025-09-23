package service

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	common "github.com/scanoss/papi/api/commonv2"
	zlog "scanoss.com/provenance/pkg/logger"
	"scanoss.com/provenance/pkg/models"
)

type mockResponse struct {
	status *common.StatusResponse
}

func (m mockResponse) GetStatus() *common.StatusResponse {
	return m.status
}

func Test_buildErrorMessages(t *testing.T) {
	tests := []struct {
		name    string
		summary models.QuerySummary
		want    []string
	}{
		{
			name:    "Empty summary returns no messages",
			summary: models.QuerySummary{},
			want:    nil,
		},
		{
			name: "Failed to parse purls only",
			summary: models.QuerySummary{
				PurlsFailedToParse: []string{"invalid-purl1", "invalid-purl2"},
			},
			want: []string{"Failed to parse 2 purl(s):invalid-purl1,invalid-purl2"},
		},
		{
			name: "Not found purls only",
			summary: models.QuerySummary{
				PurlsNotFound: []string{"not-found1", "not-found2"},
			},
			want: []string{"Can't find 2 purl(s):not-found1,not-found2"},
		},
		{
			name: "Purls without info only",
			summary: models.QuerySummary{
				PurlsWOInfo: []string{"no-info1", "no-info2"},
			},
			want: []string{"Can't find information for 2 purl(s):no-info1,no-info2"},
		},
		{
			name: "Too much data purls only",
			summary: models.QuerySummary{
				PurlsTooMuchData: []string{"too-much1", "too-much2"},
			},
			want: []string{"Too many contributors for: too-much1, too-much2"},
		},
		{
			name: "All error types combined",
			summary: models.QuerySummary{
				PurlsFailedToParse: []string{"invalid1"},
				PurlsNotFound:      []string{"not-found1"},
				PurlsWOInfo:        []string{"no-info1"},
				PurlsTooMuchData:   []string{"too-much1"},
			},
			want: []string{
				"Failed to parse 1 purl(s):invalid1",
				"Can't find 1 purl(s):not-found1",
				"Can't find information for 1 purl(s):no-info1",
				"Too many contributors for: too-much1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildErrorMessages(tt.summary); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildErrorMessages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_determineStatusAndHTTPCode(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer zlog.SyncZap()

	tests := []struct {
		name         string
		summary      models.QuerySummary
		wantStatus   common.StatusCode
		wantHTTPCode int
		wantError    bool
		wantMessage  string
	}{
		{
			name: "All successful",
			summary: models.QuerySummary{
				TotalPurls: 2,
			},
			wantStatus:   common.StatusCode_SUCCESS,
			wantHTTPCode: http.StatusOK,
			wantError:    false,
			wantMessage:  ResponseMessageSuccess,
		},
		{
			name: "All failed to parse",
			summary: models.QuerySummary{
				PurlsFailedToParse: []string{"invalid1", "invalid2"},
				TotalPurls:         2,
			},
			wantStatus:   common.StatusCode_FAILED,
			wantHTTPCode: http.StatusBadRequest,
			wantError:    true,
		},
		{
			name: "All not found",
			summary: models.QuerySummary{
				PurlsNotFound: []string{"not-found1", "not-found2"},
				TotalPurls:    2,
			},
			wantStatus:   common.StatusCode_FAILED,
			wantHTTPCode: http.StatusNotFound,
			wantError:    true,
		},
		{
			name: "All without info",
			summary: models.QuerySummary{
				PurlsWOInfo: []string{"no-info1", "no-info2"},
				TotalPurls:  2,
			},
			wantStatus:   common.StatusCode_FAILED,
			wantHTTPCode: http.StatusNotFound,
			wantError:    true,
		},
		{
			name: "Mixed with too much data",
			summary: models.QuerySummary{
				PurlsTooMuchData: []string{"too-much1"},
				TotalPurls:       2,
			},
			wantStatus:   common.StatusCode_SUCCESS,
			wantHTTPCode: http.StatusOK,
			wantError:    false,
		},
		{
			name: "Mixed with some without info",
			summary: models.QuerySummary{
				PurlsWOInfo: []string{"no-info1"},
				TotalPurls:  2,
			},
			wantStatus:   common.StatusCode_SUCCEEDED_WITH_WARNINGS,
			wantHTTPCode: http.StatusOK,
			wantError:    false,
		},
		{
			name: "Mixed success and failure",
			summary: models.QuerySummary{
				PurlsNotFound: []string{"not-found1"},
				TotalPurls:    3,
			},
			wantStatus:   common.StatusCode_SUCCEEDED_WITH_WARNINGS,
			wantHTTPCode: http.StatusOK,
			wantError:    false,
			wantMessage:  "Can't find 1 purl(s):not-found1",
		},
		{
			name: "Mixed failures return not found status",
			summary: models.QuerySummary{
				PurlsFailedToParse: []string{"invalid1"},
				PurlsNotFound:      []string{"not-found1"},
				TotalPurls:         2,
			},
			wantStatus:   common.StatusCode_FAILED,
			wantHTTPCode: http.StatusNotFound,
			wantError:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStatus, gotHTTPCode, err := determineStatusAndHTTPCode(zlog.S, tt.summary)
			if (err != nil) != tt.wantError {
				t.Errorf("determineStatusAndHTTPCode() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError {
				if gotStatus.Status != tt.wantStatus {
					t.Errorf("determineStatusAndHTTPCode() gotStatus = %v, want %v", gotStatus.Status, tt.wantStatus)
				}
				if tt.wantMessage != "" && gotStatus.Message != tt.wantMessage {
					t.Errorf("determineStatusAndHTTPCode() gotMessage = %v, want %v", gotStatus.Message, tt.wantMessage)
				}
			}
			if gotHTTPCode != tt.wantHTTPCode {
				t.Errorf("determineStatusAndHTTPCode() gotHTTPCode = %v, want %v", gotHTTPCode, tt.wantHTTPCode)
			}
		})
	}
}

func Test_buildStatusResponse(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer zlog.SyncZap()

	ctx := context.Background()
	ctx = ctxzap.ToContext(ctx, zlog.L)

	tests := []struct {
		name       string
		summary    models.QuerySummary
		wantStatus common.StatusCode
		wantError  bool
	}{
		{
			name: "Successful response",
			summary: models.QuerySummary{
				TotalPurls: 2,
			},
			wantStatus: common.StatusCode_SUCCESS,
			wantError:  false,
		},
		{
			name: "Failed response",
			summary: models.QuerySummary{
				PurlsFailedToParse: []string{"invalid1"},
				TotalPurls:         1,
			},
			wantStatus: common.StatusCode_FAILED,
			wantError:  true,
		},
		{
			name: "Warning response",
			summary: models.QuerySummary{
				PurlsWOInfo: []string{"no-info1"},
				TotalPurls:  2,
			},
			wantStatus: common.StatusCode_SUCCEEDED_WITH_WARNINGS,
			wantError:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildStatusResponse(ctx, zlog.S, tt.summary)
			if (err != nil) != tt.wantError {
				t.Errorf("buildStatusResponse() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && got.Status != tt.wantStatus {
				t.Errorf("buildStatusResponse() = %v, want %v", got.Status, tt.wantStatus)
			}
		})
	}
}

func Test_resolveResponseStatus(t *testing.T) {
	tests := []struct {
		name     string
		response interface{}
		want     *common.StatusResponse
	}{
		{
			name:     "Nil response returns default failed status",
			response: nil,
			want: &common.StatusResponse{
				Status:  common.StatusCode_FAILED,
				Message: ResponseMessageError,
			},
		},
		{
			name: "Valid response with status",
			response: mockResponse{
				status: &common.StatusResponse{
					Status:  common.StatusCode_SUCCESS,
					Message: "Test success",
				},
			},
			want: &common.StatusResponse{
				Status:  common.StatusCode_SUCCESS,
				Message: "Test success",
			},
		},
		{
			name: "Response with nil status returns default",
			response: mockResponse{
				status: nil,
			},
			want: &common.StatusResponse{
				Status:  common.StatusCode_FAILED,
				Message: ResponseMessageError,
			},
		},
		{
			name:     "Response without GetStatus method returns default",
			response: "invalid response",
			want: &common.StatusResponse{
				Status:  common.StatusCode_FAILED,
				Message: ResponseMessageError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveResponseStatus(tt.response); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("resolveResponseStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}
