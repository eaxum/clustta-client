package constants

import "fmt"

// const HOST string = "https://api.clustta.com"
var HOST = getHost()
var DASHBOARD_HOST = getDashboardHost()
var WEB_SITE = getWebsite()

var USER_AGENT = fmt.Sprintf("Clustta/%s", "0.2")

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

// host will be replaced by the value passed at build time using ldflags
var host string
var dashboard_host string
var website string
