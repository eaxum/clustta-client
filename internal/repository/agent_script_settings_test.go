package repository

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeAgentScriptSettings(t *testing.T) {
	settings, err := NormalizeAgentScriptSettings(AgentScriptSettings{
		Directory:  "tools/scripts",
		Extensions: []string{"PY", "*.bat", ".py", " .MEL "},
	})

	require.NoError(t, err)
	require.Equal(t, "tools/scripts", settings.Directory)
	require.Equal(t, []string{".bat", ".mel", ".py"}, settings.Extensions)
}

func TestNormalizeAgentScriptSettingsRejectsOutsideDirectory(t *testing.T) {
	_, err := NormalizeAgentScriptSettings(AgentScriptSettings{Directory: "../scripts"})
	require.ErrorContains(t, err, "relative to the project root")
}

func TestNormalizeAgentScriptSettingsRejectsAbsoluteDirectory(t *testing.T) {
	_, err := NormalizeAgentScriptSettings(AgentScriptSettings{Directory: `C:\scripts`})
	require.ErrorContains(t, err, "relative to the project root")
}
