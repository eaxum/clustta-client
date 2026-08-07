package constants

import "fmt"

// const HOST string = "https://api.clustta.com"
var HOST = getHost()
var DASHBOARD_HOST = getDashboardHost()
var WEB_SITE = getWebsite()

var USER_AGENT = fmt.Sprintf("Clustta/%s", "0.2")

// fallbackVersion is used when no build-time version is injected.
// Keep in sync with build/config.yml; release CI may override via ldflags.
const fallbackVersion = "0.4.38"

// VERSION is the resolved application version (build-time injected or fallback).
var VERSION = getVersion()

// CHANNEL is an optional build-time channel override; when empty it is detected
// at runtime. Values: "msstore", "mas", "flathub", "direct".
var CHANNEL = channel

// UPDATE_MANIFEST_URL points to the JSON manifest describing the latest release.
const UPDATE_MANIFEST_URL = "https://eaxum.github.io/clustta-client/latest.json"

// Store / distribution destinations used when the client cannot self-update.
const (
	MS_STORE_PRODUCT_ID = "9PNRGHGP3LGX"
	MAS_APP_ID          = "6748349288"
	FLATHUB_APP_ID      = "com.clustta.clustta"
	RELEASES_PAGE_URL   = "https://github.com/eaxum/clustta-client/releases/latest"
)

// version and channel are replaced at build time using ldflags.
var version string

// getHost determines the HOST based on the build-time variable
func getHost() string {
	if host != "" {
		return host
	}
	// Fallback if not set at build time
	return "https://api.clustta.com"
}
func getDashboardHost() string {
	if dashboard_host != "" {
		return dashboard_host
	}
	// Fallback if not set at build time
	return "https://app.clustta.com"
}

func getWebsite() string {
	if website != "" {
		return website
	}
	// Fallback if not set at build time
	return "https://clustta.com"
}

// getVersion returns the build-time injected version or the fallback.
func getVersion() string {
	if version != "" {
		return version
	}
	return fallbackVersion
}

// host will be replaced by the value passed at build time using ldflags
var host string
var dashboard_host string
var website string
var channel string
