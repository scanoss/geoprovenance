// Package domain defines provenance-specific status codes that extend the
// shared go-grpc-helper domain status codes.
package domain

import "github.com/scanoss/go-grpc-helper/pkg/grpc/domain"

const (
	TooManyContributors domain.StatusCode = "TOO_MANY_CONTRIBUTORS"
)
