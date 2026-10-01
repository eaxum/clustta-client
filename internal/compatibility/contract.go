package compatibility

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

const (
	LegacyAPIVersion     = "1"
	CurrentAPIVersion    = "2"
	CurrentProjectSchema = "2.2"
	APIVersionHeader     = "Clustta-API-Version"
	UnsupportedCode      = "api_version_unsupported"

	VersionedDependenciesCapability = "versioned_dependencies"
	ProjectPermissionsCapability    = "project_permissions"
)

var supportedAPIVersions = []string{LegacyAPIVersion, CurrentAPIVersion}

var capabilitiesByVersion = map[string][]string{
	LegacyAPIVersion:  {},
	CurrentAPIVersion: {VersionedDependenciesCapability, ProjectPermissionsCapability},
}

type APIInfo struct {
	DefaultVersion        string              `json:"default_version"`
	SupportedVersions     []string            `json:"supported_versions"`
	CapabilitiesByVersion map[string][]string `json:"capabilities_by_version"`
}

type APIContext struct {
	Version      string   `json:"version"`
	Capabilities []string `json:"capabilities"`
}

type UnsupportedAPIError struct {
	Code              string   `json:"code"`
	Message           string   `json:"message"`
	RequestedVersion  string   `json:"requested_version,omitempty"`
	SupportedVersions []string `json:"supported_versions"`
}

func (e *UnsupportedAPIError) Error() string {
	data, err := json.Marshal(e)
	if err != nil {
		return e.Message
	}
	return string(data)
}

func Info() APIInfo {
	versions := append([]string(nil), supportedAPIVersions...)
	capabilities := make(map[string][]string, len(capabilitiesByVersion))
	for version, values := range capabilitiesByVersion {
		capabilities[version] = append([]string(nil), values...)
	}
	return APIInfo{
		DefaultVersion:        LegacyAPIVersion,
		SupportedVersions:     versions,
		CapabilitiesByVersion: capabilities,
	}
}

func SelectHighestMutual(info APIInfo) APIContext {
	supported := make(map[string]bool, len(info.SupportedVersions))
	for _, version := range info.SupportedVersions {
		supported[version] = true
	}
	for index := len(supportedAPIVersions) - 1; index >= 0; index-- {
		version := supportedAPIVersions[index]
		if supported[version] {
			return APIContext{
				Version:      version,
				Capabilities: append([]string(nil), info.CapabilitiesByVersion[version]...),
			}
		}
	}
	return APIContext{Version: LegacyAPIVersion}
}

func Negotiate(headers http.Header) (APIContext, error) {
	version := headers.Get(APIVersionHeader)
	if version == "" {
		version = LegacyAPIVersion
	}
	if !supportsAPI(version) {
		return APIContext{}, &UnsupportedAPIError{
			Code:              UnsupportedCode,
			Message:           "Update Clustta or the Studio to use a supported API version.",
			RequestedVersion:  version,
			SupportedVersions: append([]string(nil), supportedAPIVersions...),
		}
	}
	return APIContext{
		Version:      version,
		Capabilities: append([]string(nil), capabilitiesByVersion[version]...),
	}, nil
}

type apiContextKey struct{}

func WithAPIContext(ctx context.Context, api APIContext) context.Context {
	return context.WithValue(ctx, apiContextKey{}, api)
}

func FromContext(ctx context.Context) APIContext {
	api, ok := ctx.Value(apiContextKey{}).(APIContext)
	if ok {
		return api
	}
	return APIContext{Version: LegacyAPIVersion}
}

func Respond(w http.ResponseWriter, api APIContext) {
	w.Header().Set(APIVersionHeader, api.Version)
	w.Header().Set("Cache-Control", "no-store")
}

func WriteError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	if unsupported, ok := err.(*UnsupportedAPIError); ok {
		w.WriteHeader(http.StatusUpgradeRequired)
		json.NewEncoder(w).Encode(unsupported)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]string{"code": "invalid_api_version", "message": err.Error()})
}

func ReadSchema(q sqlx.Queryer) (string, error) {
	var schema string
	if err := sqlx.Get(q, &schema, "SELECT value FROM config WHERE name = 'version'"); err != nil {
		return "", err
	}
	if !validVersion(schema) {
		return "", fmt.Errorf("invalid project schema %q", schema)
	}
	return schema, nil
}

func ValidateDatabase(q sqlx.Queryer) error {
	_, err := ReadSchema(q)
	return err
}

func CompareVersions(left, right string) (int, error) {
	leftParts, err := versionParts(left)
	if err != nil {
		return 0, err
	}
	rightParts, err := versionParts(right)
	if err != nil {
		return 0, err
	}
	if len(leftParts) != len(rightParts) {
		return 0, fmt.Errorf("incomparable schema identifiers %q and %q", left, right)
	}
	for index := range leftParts {
		if leftParts[index] < rightParts[index] {
			return -1, nil
		}
		if leftParts[index] > rightParts[index] {
			return 1, nil
		}
	}
	return 0, nil
}

func supportsAPI(version string) bool {
	for _, supported := range supportedAPIVersions {
		if version == supported {
			return true
		}
	}
	return false
}

func validVersion(value string) bool {
	_, err := versionParts(value)
	return err == nil
}

func versionParts(value string) ([]int, error) {
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid schema identifier %q", value)
	}
	numbers := make([]int, len(parts))
	for index, part := range parts {
		number, err := strconv.Atoi(part)
		if err != nil || number < 0 || strconv.Itoa(number) != part {
			return nil, fmt.Errorf("invalid schema identifier %q", value)
		}
		numbers[index] = number
	}
	return numbers, nil
}
