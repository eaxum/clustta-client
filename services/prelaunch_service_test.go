package services

import (
	"clustta/internal/repository"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildHookEnvironmentExpandsProjectRoot(t *testing.T) {
	projectRoot := t.TempDir()
	configPath := filepath.Join(projectRoot, "configs", "show.ocio")
	require.NoError(t, os.MkdirAll(filepath.Dir(configPath), 0o755))
	require.NoError(t, os.WriteFile(configPath, []byte("ocio_profile_version: 2"), 0o600))
	environment, err := buildHookEnvironment(projectRoot, []repository.PreLaunchEnvironmentVariable{{
		Name: "OCIO", Value: "<ProjectRoot>/configs/show.ocio",
	}})

	require.NoError(t, err)
	var ocio string
	for _, entry := range environment {
		if strings.HasPrefix(entry, "OCIO=") {
			ocio = entry
		}
	}
	require.Equal(t, "OCIO="+configPath, ocio)
}

func TestBuildHookEnvironmentRejectsMissingProjectPath(t *testing.T) {
	projectRoot := t.TempDir()

	_, err := buildHookEnvironment(projectRoot, []repository.PreLaunchEnvironmentVariable{{
		Name: "OCIO", Value: "<ProjectRoot>/configs/missing.ocio",
	}})

	require.EqualError(t, err, "environment variable OCIO points to a missing path: "+
		filepath.Join(projectRoot, "configs", "missing.ocio"))
}

func TestBuildHookEnvironmentRejectsMissingRelativeOCIOPath(t *testing.T) {
	projectRoot := t.TempDir()

	_, err := buildHookEnvironment(projectRoot, []repository.PreLaunchEnvironmentVariable{{
		Name: "OCIO", Value: "configs/missing.ocio",
	}})

	require.EqualError(t, err, "environment variable OCIO points to a missing path: "+
		filepath.Join(projectRoot, "configs", "missing.ocio"))
}

func TestResolveHookEnvironmentVariablesUsesSelectedIDs(t *testing.T) {
	variables := []repository.PreLaunchEnvironmentVariable{
		{ID: "ocio", Name: "OCIO", Value: "config.ocio"},
		{ID: "cache", Name: "CACHE", Value: "cache"},
	}

	selected := resolveHookEnvironmentVariables(variables, []string{"cache"})

	require.Equal(t, []repository.PreLaunchEnvironmentVariable{variables[1]}, selected)
}

func TestResolveHookEnvironmentPathExpandsProjectRoot(t *testing.T) {
	projectRoot := t.TempDir()

	value, isProjectPath := resolveHookEnvironmentPath(projectRoot, repository.PreLaunchEnvironmentVariable{
		Name: "CACHE", Value: "<ProjectRoot>/cache",
	})

	require.True(t, isProjectPath)
	require.Equal(t, filepath.Join(projectRoot, "cache"), value)
}

func TestPathContainsNestedAsset(t *testing.T) {
	projectRoot := t.TempDir()
	configDirectory := filepath.Join(projectRoot, "Configs")

	require.True(t, pathContains(projectRoot, filepath.Join(configDirectory, "show.ocio")))
	require.True(t, pathContains(configDirectory, filepath.Join(configDirectory, "show.ocio")))
	require.False(t, pathContains(configDirectory, filepath.Join(projectRoot, "Scripts", "setup.py")))
}

func TestBlenderBootstrapOpensSceneBeforeScripts(t *testing.T) {
	source, err := buildDCCBootstrap("shot.blend", "project", repository.PreLaunchHook{
		FailurePolicy: repository.PreLaunchFailureBlock,
	}, []string{"setup.py"})

	require.NoError(t, err)
	openIndex := strings.Index(source, "bpy.ops.wm.open_mainfile")
	scriptIndex := strings.Index(source, "for script in scripts")
	require.Greater(t, openIndex, 0)
	require.Greater(t, scriptIndex, openIndex)
}

func TestMayaBootstrapRunsScriptsBeforeOpening(t *testing.T) {
	source, err := buildDCCBootstrap("shot.ma", "project", repository.PreLaunchHook{
		FailurePolicy: repository.PreLaunchFailureBlock,
	}, []string{"setup.py"})

	require.NoError(t, err)
	scriptIndex := strings.Index(source, "for script in scripts")
	openIndex := strings.Index(source, "cmds.file(target")
	require.Greater(t, scriptIndex, 0)
	require.Greater(t, openIndex, scriptIndex)
}
