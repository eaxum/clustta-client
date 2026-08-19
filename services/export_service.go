package services

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"clustta/internal/repository/models"
)

const (
	exportFormatCSV  = "csv"
	exportFormatJSON = "json"
	exportFormatText = "txt"
	exportPageSize   = 20
)

var requiredExportColumns = []string{"name", "extension", "parent"}

var (
	nameBoundaryPattern  = regexp.MustCompile(`([\p{Ll}\p{N}])([\p{Lu}])`)
	nameSeparatorPattern = regexp.MustCompile(`[^\p{L}\p{N}]+`)
)

type ExportService struct{}

type ExportRequest struct {
	AssetIDs   []string `json:"asset_ids"`
	Scope      string   `json:"scope"`
	Columns    []string `json:"columns"`
	NameFormat string   `json:"name_format"`
	Page       int      `json:"page"`
}

type ExportColumn struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Required bool   `json:"required"`
}

type ExportPreview struct {
	Columns  []ExportColumn           `json:"columns"`
	Rows     []map[string]interface{} `json:"rows"`
	Total    int                      `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"page_size"`
}

var availableExportColumns = []ExportColumn{
	{Key: "name", Label: "Name", Required: true},
	{Key: "extension", Label: "Extension", Required: true},
	{Key: "parent", Label: "Parent", Required: true},
	{Key: "status", Label: "Status"},
	{Key: "assignee", Label: "Assignee"},
	{Key: "tags", Label: "Tags"},
	{Key: "asset_type", Label: "Asset type"},
	{Key: "created_at", Label: "Created at"},
	{Key: "updated_at", Label: "Updated at"},
	{Key: "description", Label: "Description"},
	{Key: "kind", Label: "Kind"},
	{Key: "path", Label: "Path"},
}

func (s *ExportService) Preview(projectPath string, request ExportRequest) (ExportPreview, error) {
	assets, err := s.resolveAssets(projectPath, request)
	if err != nil {
		return ExportPreview{}, err
	}
	columns, err := normalizeExportColumns(request.Columns)
	if err != nil {
		return ExportPreview{}, err
	}
	if err := validateNameFormat(request.NameFormat); err != nil {
		return ExportPreview{}, err
	}
	page := request.Page
	if page < 1 {
		page = 1
	}
	start := (page - 1) * exportPageSize
	if start > len(assets) {
		start = len(assets)
	}
	end := min(start+exportPageSize, len(assets))
	rows := exportRows(assets[start:end], columns, request.NameFormat)

	return ExportPreview{
		Columns:  availableExportColumns,
		Rows:     rows,
		Total:    len(assets),
		Page:     page,
		PageSize: exportPageSize,
	}, nil
}

func (s *ExportService) Export(projectPath, destinationPath, format string, request ExportRequest) (string, error) {
	assets, err := s.resolveAssets(projectPath, request)
	if err != nil {
		return "", err
	}
	columns, err := normalizeExportColumns(request.Columns)
	if err != nil {
		return "", err
	}
	if err := validateNameFormat(request.NameFormat); err != nil {
		return "", err
	}
	format = strings.ToLower(strings.TrimSpace(format))
	if format != exportFormatCSV && format != exportFormatJSON && format != exportFormatText {
		return "", fmt.Errorf("unsupported export format %q", format)
	}
	if strings.TrimSpace(destinationPath) == "" {
		return "", errors.New("export destination is required")
	}

	outputPath := ensureExportExtension(destinationPath, format)
	data, err := serializeExportRows(exportRows(assets, columns, request.NameFormat), columns, format)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(outputPath, data, 0o600); err != nil {
		return "", fmt.Errorf("write export: %w", err)
	}
	return outputPath, nil
}

func (s *ExportService) resolveAssets(projectPath string, request ExportRequest) ([]models.Asset, error) {
	if err := validateExportScope(request.Scope); err != nil {
		return nil, err
	}
	assets, err := (&AssetService{}).GetAssets(projectPath)
	if err != nil {
		return nil, err
	}
	wanted := make(map[string]struct{}, len(request.AssetIDs))
	for _, id := range request.AssetIDs {
		if id != "" {
			wanted[id] = struct{}{}
		}
	}
	resolved := make([]models.Asset, 0, len(wanted))
	for _, asset := range assets {
		if asset.Trashed {
			continue
		}
		if exportScopeMatches(asset, request.Scope, wanted) {
			resolved = append(resolved, asset)
		}
	}
	sort.Slice(resolved, func(i, j int) bool {
		return strings.ToLower(resolved[i].AssetPath) < strings.ToLower(resolved[j].AssetPath)
	})
	return resolved, nil
}

func validateExportScope(scope string) error {
	switch scope {
	case "", "selection", "all_assets", "all_tasks", "blender_tasks":
		return nil
	default:
		return fmt.Errorf("unsupported export scope %q", scope)
	}
}

func exportScopeMatches(asset models.Asset, scope string, wanted map[string]struct{}) bool {
	switch scope {
	case "all_assets":
		return true
	case "all_tasks":
		return !asset.IsResource
	case "blender_tasks":
		return !asset.IsResource && strings.EqualFold(asset.Extension, ".blend")
	case "selection", "":
		_, ok := wanted[asset.Id]
		return ok
	default:
		return false
	}
}

func normalizeExportColumns(columns []string) ([]string, error) {
	available := make(map[string]struct{}, len(availableExportColumns))
	for _, column := range availableExportColumns {
		available[column.Key] = struct{}{}
	}
	selected := make(map[string]struct{}, len(columns)+len(requiredExportColumns))
	for _, column := range append(requiredExportColumns, columns...) {
		column = strings.TrimSpace(column)
		if _, ok := available[column]; !ok {
			return nil, fmt.Errorf("unsupported export column %q", column)
		}
		selected[column] = struct{}{}
	}
	ordered := make([]string, 0, len(selected))
	for _, column := range availableExportColumns {
		if _, ok := selected[column.Key]; ok {
			ordered = append(ordered, column.Key)
		}
	}
	return ordered, nil
}

func exportRows(assets []models.Asset, columns []string, nameFormat string) []map[string]interface{} {
	rows := make([]map[string]interface{}, 0, len(assets))
	for _, asset := range assets {
		allValues := map[string]interface{}{
			"name":        formatExportName(asset.Name, nameFormat),
			"extension":   asset.Extension,
			"parent":      asset.CollectionPath,
			"status":      asset.StatusShortName,
			"assignee":    asset.AssigneeName,
			"tags":        asset.Tags,
			"asset_type":  asset.AssetTypeName,
			"created_at":  asset.CreatedAt,
			"updated_at":  asset.MTime,
			"description": asset.Description,
			"kind":        map[bool]string{true: "resource", false: "task"}[asset.IsResource],
			"path":        asset.AssetPath + asset.Extension,
		}
		row := make(map[string]interface{}, len(columns))
		for _, column := range columns {
			row[column] = allValues[column]
		}
		rows = append(rows, row)
	}
	return rows
}

func validateNameFormat(nameFormat string) error {
	switch nameFormat {
	case "", "original", "kebab", "snake", "camel", "pascal", "uppercase", "lowercase", "title":
		return nil
	default:
		return fmt.Errorf("unsupported name format %q", nameFormat)
	}
}

func formatExportName(name, nameFormat string) string {
	if nameFormat == "" || nameFormat == "original" {
		return name
	}
	words := exportNameWords(name)
	if len(words) == 0 {
		return name
	}
	switch nameFormat {
	case "kebab":
		return strings.Join(lowercaseWords(words), "-")
	case "snake":
		return strings.Join(lowercaseWords(words), "_")
	case "camel":
		formatted := lowercaseWords(words)
		for index := 1; index < len(formatted); index++ {
			formatted[index] = uppercaseFirst(formatted[index])
		}
		return strings.Join(formatted, "")
	case "pascal":
		formatted := lowercaseWords(words)
		for index := range formatted {
			formatted[index] = uppercaseFirst(formatted[index])
		}
		return strings.Join(formatted, "")
	case "uppercase":
		return strings.ToUpper(strings.Join(words, " "))
	case "lowercase":
		return strings.Join(lowercaseWords(words), " ")
	case "title":
		formatted := lowercaseWords(words)
		for index := range formatted {
			formatted[index] = uppercaseFirst(formatted[index])
		}
		return strings.Join(formatted, " ")
	default:
		return name
	}
}

func exportNameWords(name string) []string {
	withBoundaries := nameBoundaryPattern.ReplaceAllString(name, `${1} ${2}`)
	return strings.Fields(nameSeparatorPattern.ReplaceAllString(withBoundaries, " "))
}

func lowercaseWords(words []string) []string {
	formatted := make([]string, len(words))
	for index, word := range words {
		formatted[index] = strings.ToLower(word)
	}
	return formatted
}

func uppercaseFirst(value string) string {
	runes := []rune(value)
	if len(runes) == 0 {
		return value
	}
	runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
	return string(runes)
}

func serializeExportRows(rows []map[string]interface{}, columns []string, format string) ([]byte, error) {
	switch format {
	case exportFormatJSON:
		return json.MarshalIndent(rows, "", "  ")
	case exportFormatCSV:
		var output bytes.Buffer
		writer := csv.NewWriter(&output)
		if err := writer.Write(columns); err != nil {
			return nil, err
		}
		for _, row := range rows {
			values := make([]string, 0, len(columns))
			for _, column := range columns {
				values = append(values, exportString(row[column]))
			}
			if err := writer.Write(values); err != nil {
				return nil, err
			}
		}
		writer.Flush()
		return output.Bytes(), writer.Error()
	case exportFormatText:
		var output strings.Builder
		output.WriteString(strings.Join(columns, "\t"))
		output.WriteByte('\n')
		for _, row := range rows {
			values := make([]string, 0, len(columns))
			for _, column := range columns {
				value := strings.NewReplacer("\t", " ", "\r", " ", "\n", " ").Replace(exportString(row[column]))
				values = append(values, value)
			}
			output.WriteString(strings.Join(values, "\t"))
			output.WriteByte('\n')
		}
		return []byte(output.String()), nil
	default:
		return nil, fmt.Errorf("unsupported export format %q", format)
	}
}

func exportString(value interface{}) string {
	switch typed := value.(type) {
	case []string:
		return strings.Join(typed, ", ")
	case int:
		return strconv.Itoa(typed)
	case nil:
		return ""
	default:
		return fmt.Sprint(typed)
	}
}

func ensureExportExtension(path, format string) string {
	extension := "." + format
	if strings.EqualFold(filepath.Ext(path), extension) {
		return path
	}
	return path + extension
}
