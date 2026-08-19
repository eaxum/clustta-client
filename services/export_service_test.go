package services

import (
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
)

func TestNormalizeExportColumnsKeepsRequiredColumns(t *testing.T) {
	columns, err := normalizeExportColumns([]string{"tags", "status"})
	if err != nil {
		t.Fatal(err)
	}
	want := "name,extension,parent,status,tags"
	if got := strings.Join(columns, ","); got != want {
		t.Fatalf("columns = %q, want %q", got, want)
	}
}

func TestSerializeExportRowsCSVQuotesValues(t *testing.T) {
	columns := []string{"name", "extension", "parent", "tags"}
	rows := []map[string]interface{}{{
		"name": "House, exterior", "extension": ".blend", "parent": "/sets", "tags": []string{"hero", "day"},
	}}
	data, err := serializeExportRows(rows, columns, exportFormatCSV)
	if err != nil {
		t.Fatal(err)
	}
	records, err := csv.NewReader(strings.NewReader(string(data))).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if records[1][0] != "House, exterior" || records[1][3] != "hero, day" {
		t.Fatalf("unexpected CSV row: %#v", records[1])
	}
}

func TestSerializeExportRowsJSONUsesSelectedColumns(t *testing.T) {
	columns := []string{"name", "extension", "parent"}
	rows := []map[string]interface{}{{"name": "House", "extension": ".blend", "parent": "/sets"}}
	data, err := serializeExportRows(rows, columns, exportFormatJSON)
	if err != nil {
		t.Fatal(err)
	}
	var decoded []map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded[0]["name"] != "House" || len(decoded[0]) != len(columns) {
		t.Fatalf("unexpected JSON row: %#v", decoded[0])
	}
}

func TestEnsureExportExtension(t *testing.T) {
	if got := ensureExportExtension("report", exportFormatJSON); got != "report.json" {
		t.Fatalf("path = %q", got)
	}
	if got := ensureExportExtension("report.JSON", exportFormatJSON); got != "report.JSON" {
		t.Fatalf("path = %q", got)
	}
}

func TestValidateExportScopeRejectsUnknownScope(t *testing.T) {
	if err := validateExportScope("outside_project"); err == nil {
		t.Fatal("expected unsupported scope error")
	}
}
