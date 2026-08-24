package dcc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersionAliasesUsesShortInstallVersion(t *testing.T) {
	require.Equal(t, []string{"5.2.0", "5.2"}, versionAliases("5.2.0"))
	require.Equal(t, []string{"2025.1"}, versionAliases("2025.1"))
}

func TestFindExecutableRejectsInvalidVersion(t *testing.T) {
	_, err := FindExecutable("blender", "../5.2")
	require.ErrorContains(t, err, "invalid blender version")
}

func TestExecutableNotFoundErrorCanBeIdentified(t *testing.T) {
	err := ExecutableNotFoundError{Name: "blender", Version: "5.2"}
	require.True(t, IsExecutableNotFound(err))
	require.Equal(t, "blender 5.2 was not found in standard install locations", err.Error())
}
