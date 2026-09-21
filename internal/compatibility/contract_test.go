package compatibility

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

func TestMissingHostContractRequiresServerUpdate(t *testing.T) {
	var rejection *Rejection
	if !errors.As(Check(nil), &rejection) || rejection.RequiredUpdate != "server" {
		t.Fatal("missing contract approved")
	}
}

func Respond(w http.ResponseWriter, schema string) {
	w.Header().Set(ProtocolHeader, Protocol)
	w.Header().Set(SchemaHeader, Schema)
	w.Header().Set(ProjectSchemaHeader, schema)
	w.Header().Set("Cache-Control", "no-store")
}

func WriteError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	if rejection, ok := err.(*Rejection); ok {
		w.WriteHeader(http.StatusUpgradeRequired)
		json.NewEncoder(w).Encode(rejection)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]string{"code": "invalid_compatibility_declaration", "message": err.Error()})
}

func TestContractUpdateDirection(t *testing.T) {
	for _, test := range []struct{ schema, update string }{{Schema, ""}, {"2.1", "server"}, {"2.10", "server"}} {
		err := Check(Current(test.schema))
		if test.update == "" {
			if err != nil {
				t.Fatal(err)
			}
			continue
		}
		var rejection *Rejection
		if !errors.As(err, &rejection) || rejection.RequiredUpdate != test.update {
			t.Fatalf("schema %s: %v", test.schema, err)
		}
	}
}

func TestExactSchemaOrdering(t *testing.T) {
	comparison, err := CompareVersions("2.10", "2.2")
	if err != nil || comparison <= 0 {
		t.Fatalf("unexpected comparison: %d %v", comparison, err)
	}
	if _, err := CompareVersions("2.10", "2.1"); err != nil {
		t.Fatal(err)
	}
	if _, err := CompareVersions("2.01", "2.1"); err == nil {
		t.Fatal("accepted a non-canonical schema identifier")
	}
	var rejection *Rejection
	if !errors.As(Check(Current("2.01")), &rejection) || rejection.RequiredUpdate != "server" {
		t.Fatal("malformed host schema did not require a host update")
	}
}
