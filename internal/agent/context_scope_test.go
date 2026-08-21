package agent

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplyAuthoritativeScopeResolvesEntityReferencePathToID(t *testing.T) {
	context := turnContext{EntityReferences: []map[string]interface{}{
		{
			"token": "@Assets/characters/aremu/aremu.blend", "type": "asset",
			"id": "asset-1", "name": "aremu", "extension": ".blend",
			"path": "Assets/characters/aremu/aremu.blend",
		},
	}}
	args := map[string]interface{}{
		"scope": map[string]interface{}{
			"source": "entity", "path": "Assets/characters/aremu/aremu.blend",
		},
	}

	applyAuthoritativeScope(args, context)

	scopeArgs := args["scope"].(map[string]interface{})
	require.Equal(t, "asset-1", scopeArgs["entity_id"])
	require.NotContains(t, scopeArgs, "path")
}

func TestApplyAuthoritativeScopeLeavesHereScopeMappedToBrowserLocation(t *testing.T) {
	context := turnContext{
		HereScope: map[string]interface{}{"entity_id": "browser-collection"},
		EntityReferences: []map[string]interface{}{
			{"token": "@Scripts", "type": "collection", "id": "scripts", "name": "Scripts"},
		},
	}
	args := map[string]interface{}{
		"scope": map[string]interface{}{"source": "here", "entity_id": "scripts"},
	}

	applyAuthoritativeScope(args, context)

	scopeArgs := args["scope"].(map[string]interface{})
	require.Equal(t, "browser-collection", scopeArgs["entity_id"])
}

func TestApplyAuthoritativeScopeExpandsReferencedCollectionContents(t *testing.T) {
	context := turnContext{EntityReferences: []map[string]interface{}{
		{"token": "@Scripts", "type": "collection", "id": "scripts", "name": "Scripts"},
	}}
	args := map[string]interface{}{
		"scope": map[string]interface{}{
			"source": "entity", "path": "Scripts", "types": []interface{}{"asset"},
		},
	}

	applyAuthoritativeScope(args, context)

	scopeArgs := args["scope"].(map[string]interface{})
	require.Equal(t, "scripts", scopeArgs["entity_id"])
	require.Equal(t, true, scopeArgs["recursive"])
}

func TestApplyAuthoritativeScopeResolvesMultipleReferenceTokens(t *testing.T) {
	context := turnContext{EntityReferences: []map[string]interface{}{
		{"token": "@Characters/Zeus.blend", "id": "zeus"},
		{"token": "@Characters/Aremu.blend", "id": "aremu"},
	}}
	args := map[string]interface{}{
		"scope": map[string]interface{}{
			"source":     "entities",
			"entity_ids": []interface{}{"@Characters/Zeus.blend", "@Characters/Aremu.blend"},
		},
	}

	applyAuthoritativeScope(args, context)

	scopeArgs := args["scope"].(map[string]interface{})
	require.Equal(t, []interface{}{"zeus", "aremu"}, scopeArgs["entity_ids"])
}
