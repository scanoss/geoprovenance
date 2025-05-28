package dtos

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
	"testing"
)

func TestProvenanceInput_ParseProvenanceInput(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := context.Background()
	ctx = ctxzap.ToContext(ctx, zlog.L)
	s := ctxzap.Extract(ctx).Sugar()
	tests := []struct {
		name        string
		input       []byte
		expectError bool
	}{
		{
			name:        "ShouldReturn_Error_When_Empty_Input",
			input:       []byte(""),
			expectError: true,
		},
		{
			name:        "ShouldReturn_Error_Invalid_Json_Input",
			input:       []byte(`{ "Purls": [ { "purl": "pkg:npm/foo"  ] }`),
			expectError: true,
		},
		{
			name:        "ShouldReturn_Valid_Provenance_Input",
			input:       []byte(`{ "Purls": [ { "Purl": "pkg:npm/foo" }  ] }`),
			expectError: false,
		},
		{
			name:        "ShouldReturn_Valid_Provenance_Input_With_Requirement",
			input:       []byte(`{ "Purls": [ { "Purl": "pkg:npm/foo", "Requirement": "1.0.0" }  ] }`),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseProvenanceInput(s, tt.input)

			if err == nil && tt.expectError {
				t.Errorf("TestProvenanceInput_ParseProvenanceInput failed: %v", err)
			}
		})
	}

}
