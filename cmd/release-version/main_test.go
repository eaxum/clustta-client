package main

import "testing"

func TestParseReleaseVersion(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		version        string
		packageVersion string
	}{
		{name: "stable", input: "0.4.42", version: "0.4.42", packageVersion: "0.4.42"},
		{name: "prefixed", input: "v0.4.42", version: "0.4.42", packageVersion: "0.4.42"},
		{name: "prerelease", input: "0.4.42-beta.1", version: "0.4.42-beta.1", packageVersion: "0.4.42"},
		{name: "trimmed", input: " v0.4.42-rc.2 ", version: "0.4.42-rc.2", packageVersion: "0.4.42"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			version, packageVersion, err := parseReleaseVersion(test.input)
			if err != nil {
				t.Fatalf("parse version: %v", err)
			}
			if version != test.version {
				t.Fatalf("expected version %q, got %q", test.version, version)
			}
			if packageVersion != test.packageVersion {
				t.Fatalf("expected package version %q, got %q", test.packageVersion, packageVersion)
			}
		})
	}
}

func TestParseReleaseVersionRejectsInvalidValues(t *testing.T) {
	invalidVersions := []string{"", "0.4", "0.4.42-", "0.4.42-beta..1", "0.4.42+build.1"}

	for _, version := range invalidVersions {
		t.Run(version, func(t *testing.T) {
			if _, _, err := parseReleaseVersion(version); err == nil {
				t.Fatalf("expected %q to be rejected", version)
			}
		})
	}
}

func TestPlatformVersionsUseNumericPackageVersion(t *testing.T) {
	_, packageVersion, err := parseReleaseVersion("0.4.42-beta.1")
	if err != nil {
		t.Fatalf("parse version: %v", err)
	}
	if version := macOSBundleVersion(packageVersion); version != "0.4.420" {
		t.Fatalf("expected macOS version %q, got %q", "0.4.420", version)
	}
	if version := msixPackageVersion(packageVersion); version != "0.4.42.0" {
		t.Fatalf("expected MSIX version %q, got %q", "0.4.42.0", version)
	}
}
