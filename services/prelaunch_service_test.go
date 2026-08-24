package services

import (
	"clustta/internal/repository"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildHookEnvironmentExpandsProjectRoot(t *testing.T) {
	projectRoot := filepath.Join("C:", "projects", "show")
	environment := buildHookEnvironment(projectRoot, []repository.PreLaunchEnvironmentVariable{{
		Name: "OCIO", Value: "<ProjectRoot>/configs/show.ocio",
	}})

	var ocio string
	for _, entry := range environment {
		if strings.HasPrefix(entry, "OCIO=") {
			ocio = entry
		}
	}
	require.Equal(t, "OCIO="+projectRoot+"/configs/show.ocio", ocio)
}

func TestResolveHookEnvironmentVariablesUsesSelectedIDs(t *testing.T) {
	variables := []repository.PreLaunchEnvironmentVariable{
		{ID: "ocio", Name: "OCIO", Value: "config.ocio"},
		{ID: "cache", Name: "CACHE", Value: "cache"},
	}

	selected := resolveHookEnvironmentVariables(variables, []string{"cache"})

	require.Equal(t, []repository.PreLaunchEnvironmentVariable{variables[1]}, selected)
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
