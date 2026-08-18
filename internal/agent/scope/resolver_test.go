package scope

import "testing"

func TestMatchesResourceFilter(t *testing.T) {
	task := Entity{Metadata: map[string]interface{}{"is_resource": false}}
	resource := Entity{Metadata: map[string]interface{}{"is_resource": true}}
	filters := map[string]interface{}{"is_resource": false}

	if !matches(task, filters) {
		t.Fatal("task did not match is_resource=false")
	}
	if matches(resource, filters) {
		t.Fatal("resource matched is_resource=false")
	}
}
