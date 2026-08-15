package agent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveScriptDirectoryFromProjectRoot(t *testing.T) {
	projectRoot := t.TempDir()
	scriptDirectory := filepath.Join(projectRoot, "tools", "scripts")
	require.NoError(t, os.MkdirAll(scriptDirectory, 0o755))

	resolved, err := resolveScriptDirectory(projectRoot, "tools/scripts")

	require.NoError(t, err)
	require.Equal(t, scriptDirectory, resolved)
}
