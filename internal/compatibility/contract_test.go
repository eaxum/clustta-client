package compatibility

import (
	"net/http"
	"testing"
)

func TestNegotiateDefaultsLegacyClientsToAPI1(t *testing.T) {
	api, err := Negotiate(http.Header{})
	if err != nil {
		t.Fatal(err)
	}
	if api.Version != LegacyAPIVersion {
		t.Fatalf("expected API %s, got %s", LegacyAPIVersion, api.Version)
	}
	if len(api.Capabilities) != 0 {
		t.Fatalf("legacy API exposed capabilities: %v", api.Capabilities)
	}
}

func TestNegotiateCurrentAPI(t *testing.T) {
	headers := http.Header{}
	headers.Set(APIVersionHeader, CurrentAPIVersion)

	api, err := Negotiate(headers)
	if err != nil {
		t.Fatal(err)
	}
	if api.Version != CurrentAPIVersion {
		t.Fatalf("expected API %s, got %s", CurrentAPIVersion, api.Version)
	}
	if len(api.Capabilities) != 2 {
		t.Fatalf("expected current capabilities, got %v", api.Capabilities)
	}
}

func TestNegotiateRejectsUnsupportedAPI(t *testing.T) {
	headers := http.Header{}
	headers.Set(APIVersionHeader, "3")

	_, err := Negotiate(headers)
	unsupported, ok := err.(*UnsupportedAPIError)
	if !ok {
		t.Fatalf("expected UnsupportedAPIError, got %T", err)
	}
	if len(unsupported.SupportedVersions) != 2 {
		t.Fatalf("unexpected supported versions: %v", unsupported.SupportedVersions)
	}
}

func TestInfoAdvertisesLegacyDefault(t *testing.T) {
	info := Info()
	if info.DefaultVersion != LegacyAPIVersion {
		t.Fatalf("expected legacy default, got %s", info.DefaultVersion)
	}
	if len(info.SupportedVersions) != 2 {
		t.Fatalf("unexpected supported versions: %v", info.SupportedVersions)
	}
}
