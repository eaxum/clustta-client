package repository

import (
	"bytes"
	"clustta/internal/compatibility"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type projectInfoJSON ProjectInfo

func (project *ProjectInfo) UnmarshalJSON(data []byte) error {
	wire := struct {
		*projectInfoJSON
		Version json.RawMessage `json:"version"`
	}{
		projectInfoJSON: (*projectInfoJSON)(project),
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	version, err := decodeProjectVersion(wire.Version)
	if err != nil {
		return err
	}
	project.Version = version
	return nil
}

func decodeProjectVersion(data []byte) (string, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		return "", nil
	}

	var version string
	if data[0] == '"' {
		if err := json.Unmarshal(data, &version); err != nil {
			return "", err
		}
	} else {
		var number json.Number
		if err := json.Unmarshal(data, &number); err != nil {
			return "", fmt.Errorf("invalid project version: %w", err)
		}
		version = number.String()
	}
	version = normalizeProjectVersionIdentifier(version)
	if version == "" {
		return "", nil
	}
	if _, err := compatibility.CompareVersions(version, version); err != nil {
		return "", fmt.Errorf("invalid project version %q: %w", version, err)
	}
	return version, nil
}

func normalizeProjectVersionIdentifier(version string) string {
	if strings.Contains(version, ".") {
		return version
	}
	number, err := strconv.Atoi(version)
	if err != nil || number < 0 || strconv.Itoa(number) != version {
		return version
	}
	return version + ".0"
}
