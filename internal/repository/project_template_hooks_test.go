package repository

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestThreeDimensionalAnimationTemplateIncludesLaunchScripts(t *testing.T) {
	templates, err := loadTemplateDefinitions()
	require.NoError(t, err)

	for _, template := range templates.Templates {
		if template.Name != "3D Animation" {
			continue
		}
		require.Contains(t, template.CollectionTypes, TemplateType{Name: "Scripts", Icon: "code-bracket"})
		require.Len(t, template.PreLaunchHooks, 2)
		for _, hook := range template.PreLaunchHooks {
			content, err := templatePreLaunchScript(hook.Script)
			require.NoError(t, err)
			require.NotEmpty(t, content)
		}
		return
	}
	t.Fatal("3D Animation template not found")
}
