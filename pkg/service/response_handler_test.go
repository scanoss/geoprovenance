package service

import (
	"reflect"
	"testing"

	common "github.com/scanoss/papi/api/commonv2"
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
