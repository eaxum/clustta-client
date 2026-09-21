package services

import (
	"clustta/internal/compatibility"
	"testing"
)

func TestCompatibilityRejectionDefersToLocalMutation(t *testing.T) {
	problem := compatibility.Reject("2.3", "client")
	if IsMetadataTransportFailure(problem) {
		t.Fatal("compatibility rejection treated as a transport failure")
	}
	allowed, err := metadataMutationAllowsLocalFallback("unused.clst", "asset", []string{"pending-asset"}, problem)
	if !allowed || err != nil {
		t.Fatalf("unexpected fallback: allowed=%v err=%v", allowed, err)
	}
}
