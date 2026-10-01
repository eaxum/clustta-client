package services

import (
	"clustta/internal/compatibility"
	"testing"
)

func TestUnsupportedAPIDefersToLocalMutation(t *testing.T) {
	problem := &compatibility.UnsupportedAPIError{
		Code:              compatibility.UnsupportedCode,
		Message:           "unsupported API",
		RequestedVersion:  "3",
		SupportedVersions: []string{"1", "2"},
	}
	if IsMetadataTransportFailure(problem) {
		t.Fatal("compatibility rejection treated as a transport failure")
	}
	allowed, err := metadataMutationAllowsLocalFallback("unused.clst", "asset", []string{"pending-asset"}, problem)
	if !allowed || err != nil {
		t.Fatalf("unexpected fallback: allowed=%v err=%v", allowed, err)
	}
}
