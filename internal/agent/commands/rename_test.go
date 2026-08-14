package commands

import (
	"clustta/internal/agent/planning"
	"clustta/internal/agent/scope"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenameOptionsApplyTextTransformations(t *testing.T) {
	options := renameOptions{
		prependText:  "pre-",
		appendText:   "-maps",
		findText:     "House",
		replaceText:  "Home",
		removePrefix: "Old-",
		removeSuffix: "-Source",
	}

	name, matched, err := options.apply(scope.Entity{Name: "Old-House-Source"}, 0)

	require.NoError(t, err)
	require.True(t, matched)
	require.Equal(t, "pre-Home-maps", name)
}

func TestRenameOptionsApplyFormatAndTemplate(t *testing.T) {
	options := renameOptions{
		format:   "snake_case",
		template: "maps-{name}",
	}

	name, matched, err := options.apply(scope.Entity{Name: "Aremu House"}, 0)

	require.NoError(t, err)
	require.True(t, matched)
	require.Equal(t, "maps-aremu_house", name)
}

func TestRenameOptionsApplySequentialNumberSuffix(t *testing.T) {
	options := renameOptions{
		numbering: &renameNumbering{
			start: 10, step: 10, padding: 4, position: "suffix", separator: "-",
		},
	}

	name, matched, err := options.apply(scope.Entity{Name: "Shot"}, 2)

	require.NoError(t, err)
	require.True(t, matched)
	require.Equal(t, "Shot-0030", name)
}

func TestRenameOptionsApplySequentialNumberTemplate(t *testing.T) {
	options := renameOptions{
		template:  "{number}_{name}",
		numbering: &renameNumbering{start: 1, step: 1, padding: 3},
	}

	name, matched, err := options.apply(scope.Entity{Name: "Layout"}, 4)

	require.NoError(t, err)
	require.True(t, matched)
	require.Equal(t, "005_Layout", name)
}

func TestRenameOptionsApplyExactNameMapping(t *testing.T) {
	options := renameOptions{nameMappings: []renameNameMapping{{oldName: "Aremu House", newName: "aremu-house-maps"}}}

	name, matched, err := options.apply(scope.Entity{Name: "Aremu House"}, 0)
	require.NoError(t, err)
	require.True(t, matched)
	require.Equal(t, "aremu-house-maps", name)

	name, matched, err = options.apply(scope.Entity{Name: "Axe"}, 1)
	require.NoError(t, err)
	require.False(t, matched)
	require.Equal(t, "Axe", name)
}

func TestRenameOptionsApplyEntityMappingBeforeNameMapping(t *testing.T) {
	options := renameOptions{nameMappings: []renameNameMapping{
		{oldName: "Shot", newName: "generic-shot"},
		{entityID: "shot-2", oldName: "Shot", newName: "Shot-002"},
	}}

	name, matched, err := options.apply(scope.Entity{ID: "shot-2", Name: "Shot"}, 0)

	require.NoError(t, err)
	require.True(t, matched)
	require.Equal(t, "Shot-002", name)
}

func TestParseRenameOptionsAcceptsEmptyReplacementAndSeparator(t *testing.T) {
	options, err := parseRenameOptions(map[string]interface{}{
		"find_text":    "-old",
		"replace_text": "",
		"numbering": map[string]interface{}{
			"start": float64(0), "step": float64(1), "padding": float64(2),
			"position": "prefix", "separator": "",
		},
	})

	require.NoError(t, err)
	name, matched, err := options.apply(scope.Entity{Name: "item-old"}, 1)
	require.NoError(t, err)
	require.True(t, matched)
	require.Equal(t, "01item", name)
}

func TestParseRenameOptionsRejectsConflictingRules(t *testing.T) {
	_, err := parseRenameOptions(map[string]interface{}{
		"new_name":    "one-name",
		"append_text": "-maps",
	})

	require.EqualError(t, err, "new_name cannot be combined with other rename rules")
}

func TestParseRenameOptionsRejectsNumberPlaceholderWithoutNumbering(t *testing.T) {
	_, err := parseRenameOptions(map[string]interface{}{"template": "{name}-{number}"})

	require.EqualError(t, err, "numbering is required when template contains {number}")
}

func TestParseRenameOptionsRejectsDuplicateMappings(t *testing.T) {
	_, err := parseRenameOptions(map[string]interface{}{
		"name_mappings": []interface{}{
			map[string]interface{}{"old_name": "A", "new_name": "B"},
			map[string]interface{}{"old_name": "A", "new_name": "C"},
		},
	})

	require.EqualError(t, err, `duplicate old_name "A" in name_mappings`)
}

func TestPreserveSelectedRenameTargets(t *testing.T) {
	args := map[string]interface{}{
		"append_text": "-maps",
		"numbering":   map[string]interface{}{"start": float64(1)},
	}
	changes := []planning.Change{
		{
			Entity: scope.Entity{ID: "first", Name: "House"},
			After:  map[string]interface{}{"name": "House-maps-01"},
		},
		{
			Entity: scope.Entity{ID: "second", Name: "House"},
			After:  map[string]interface{}{"name": "House-maps-02"},
		},
	}

	preserveSelectedRenameTargets(args, changes)

	require.NotContains(t, args, "append_text")
	require.NotContains(t, args, "numbering")
	options, err := parseRenameOptions(args)
	require.NoError(t, err)
	firstName, matched, err := options.apply(changes[0].Entity, 0)
	require.NoError(t, err)
	require.True(t, matched)
	require.Equal(t, "House-maps-01", firstName)
	secondName, matched, err := options.apply(changes[1].Entity, 1)
	require.NoError(t, err)
	require.True(t, matched)
	require.Equal(t, "House-maps-02", secondName)
}
