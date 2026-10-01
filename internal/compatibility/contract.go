// Package compatibility defines the project compatibility contract and version policy.
package compatibility

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

const (
	Protocol            = "1"
	Schema              = "2.2"
	LegacySchema        = "2.1"
	ProtocolHeader      = "Protocol"
	SchemaHeader        = "Schema"
	ProjectSchemaHeader = "Project-Schema"
	UnsupportedCode     = "project_schema_unsupported"
)

type Contract struct {
	Protocol      string `json:"protocol"`
	Schema        string `json:"schema"`
	ProjectSchema string `json:"project_schema"`
}

type Rejection struct {
	Code           string `json:"code"`
	Message        string `json:"message"`
	ProjectSchema  string `json:"project_schema"`
	RequiredUpdate string `json:"required_update"`
}

func (r *Rejection) Error() string {
	data, err := json.Marshal(r)
	if err != nil {
		return r.Message
	}
	return string(data)
}

func Current(projectSchema string) *Contract {
	return &Contract{Protocol: Protocol, Schema: Schema, ProjectSchema: projectSchema}
}

func Reject(schema, update string) *Rejection {
	message := "Update Clustta to access this project."
	if update == "server" {
		message = "Update the Studio/server to access this project."
	}
	if update == "replica" {
		message = "This local replica uses a different project schema. Local changes have been preserved; update the replica before syncing."
	}
	return &Rejection{Code: UnsupportedCode, Message: message, ProjectSchema: schema, RequiredUpdate: update}
}

func Check(contract *Contract) error {
	if contract == nil || contract.Protocol == "" || contract.Schema == "" || contract.ProjectSchema == "" {
		return Reject("", "server")
	}
	if !validProtocol(contract.Protocol) || !validVersion(contract.Schema) || !validVersion(contract.ProjectSchema) {
		return Reject(contract.ProjectSchema, "server")
	}
	if contract.Protocol != Protocol {
		return Reject(contract.ProjectSchema, protocolUpdateDirection(Protocol, contract.Protocol))
	}
	if contract.ProjectSchema != contract.Schema {
		return Reject(contract.ProjectSchema, "server")
	}
	if contract.Schema != Schema {
		comparison, err := CompareVersions(Schema, contract.Schema)
		if err != nil {
			return Reject(contract.ProjectSchema, "server")
		}
		if comparison < 0 {
			return Reject(contract.ProjectSchema, "client")
		}
		return Reject(contract.ProjectSchema, "server")
	}
	return nil
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

func validVersion(value string) bool {
	_, err := versionParts(value)
	return err == nil
}

func validProtocol(value string) bool {
	number, err := strconv.Atoi(value)
	return err == nil && number > 0 && strconv.Itoa(number) == value
}

// CompareVersions compares exact dotted numeric identifiers.
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

func protocolUpdateDirection(clientProtocol, hostProtocol string) string {
	client, clientErr := strconv.Atoi(clientProtocol)
	host, hostErr := strconv.Atoi(hostProtocol)
	if clientErr == nil && hostErr == nil && client > host {
		return "server"
	}
	return "client"
}
