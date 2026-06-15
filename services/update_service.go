package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"clustta/internal/constants"
)

// UpdateService checks for and surfaces application updates.
type UpdateService struct{}

// UpdateInfo is returned to the frontend after an update check.
type UpdateInfo struct {
	Available      bool   `json:"available"`
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	NotesUrl       string `json:"notes_url"`
	Required       bool   `json:"required"`
	Channel        string `json:"channel"`
	// Action: "store" or "manual".
	Action string `json:"action"`
	// TargetUrl is the URL to open.
	TargetUrl string `json:"target_url"`
}

// updateManifest is the JSON document published with each release.
type updateManifest struct {
	Version        string `json:"version"`
	MinimumVersion string `json:"minimum_version"`
	NotesUrl       string `json:"notes_url"`
}

// CheckForUpdate compares the published manifest against the running version.
// On any failure it returns an UpdateInfo with Available=false.
func (s *UpdateService) CheckForUpdate() (UpdateInfo, error) {
	current := constants.VERSION
	channel := (&AppService{}).GetChannel()

	info := UpdateInfo{
		Available:      false,
		CurrentVersion: current,
		Channel:        channel,
		Action:         "none",
	}

	// Dev builds never report updates.
	if current == "" || current == "dev" {
		return info, nil
	}

	manifest, err := fetchManifest(constants.UPDATE_MANIFEST_URL)
	if err != nil {
		return info, err
	}

	info.LatestVersion = manifest.Version
	info.NotesUrl = manifest.NotesUrl

	if compareVersions(manifest.Version, current) <= 0 {
		// Up to date (or manifest older than current).
		return info, nil
	}

	info.Available = true
	info.Required = manifest.MinimumVersion != "" &&
		compareVersions(current, manifest.MinimumVersion) < 0

	switch channel {
	case "msstore":
		info.Action = "store"
		info.TargetUrl = msStoreURL()
	case "mas":
		info.Action = "store"
		info.TargetUrl = masStoreURL()
	case "flathub":
		info.Action = "manual"
		info.TargetUrl = flathubURL()
	default: // "direct"
		info.Action = "manual"
		info.TargetUrl = downloadPageURL()
	}

	return info, nil
}

// fetchManifest downloads and decodes the update manifest.
func fetchManifest(url string) (updateManifest, error) {
	var manifest updateManifest

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return manifest, err
	}
	req.Header.Set("User-Agent", constants.USER_AGENT)

	resp, err := client.Do(req)
	if err != nil {
		return manifest, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return manifest, fmt.Errorf("update manifest request failed: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MiB cap
	if err != nil {
		return manifest, err
	}
	if err := json.Unmarshal(body, &manifest); err != nil {
		return manifest, err
	}
	return manifest, nil
}

// msStoreURL returns a Microsoft Store deep link for the published product.
func msStoreURL() string {
	if constants.MS_STORE_PRODUCT_ID != "" {
		return "ms-windows-store://pdp/?productid=" + constants.MS_STORE_PRODUCT_ID
	}
	return constants.RELEASES_PAGE_URL
}

// masStoreURL returns a Mac App Store deep link for the published app.
func masStoreURL() string {
	if constants.MAS_APP_ID != "" {
		return "macappstore://apps.apple.com/app/id" + constants.MAS_APP_ID
	}
	return constants.RELEASES_PAGE_URL
}

// flathubURL returns the Flathub page for the app.
func flathubURL() string {
	if constants.FLATHUB_APP_ID != "" {
		return "https://flathub.org/apps/" + constants.FLATHUB_APP_ID
	}
	return constants.RELEASES_PAGE_URL
}

// downloadPageURL returns the website download page for direct installs.
func downloadPageURL() string {
	return constants.WEB_SITE + "/download"
}

// compareVersions compares two dotted versions, ignoring a leading "v" and any
// pre-release/build suffix (after "-" or "+"). Returns -1, 0 or 1.
func compareVersions(a, b string) int {
	pa := parseVersion(a)
	pb := parseVersion(b)
	for i := 0; i < len(pa) || i < len(pb); i++ {
		var na, nb int
		if i < len(pa) {
			na = pa[i]
		}
		if i < len(pb) {
			nb = pb[i]
		}
		if na != nb {
			if na < nb {
				return -1
			}
			return 1
		}
	}
	return 0
}

// parseVersion extracts the numeric components of a semantic-ish version.
func parseVersion(v string) []int {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	if idx := strings.IndexAny(v, "-+"); idx != -1 {
		v = v[:idx]
	}
	parts := strings.Split(v, ".")
	nums := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			n = 0
		}
		nums = append(nums, n)
	}
	return nums
}
