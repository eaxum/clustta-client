package dcc

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var versionPattern = regexp.MustCompile(`^[0-9]+(?:\.[0-9]+){0,2}$`)

type ExecutableNotFoundError struct {
	Name    string
	Version string
}

func (e ExecutableNotFoundError) Error() string {
	if e.Version == "" {
		return fmt.Sprintf("%s was not found in standard install locations", e.Name)
	}
	return fmt.Sprintf("%s %s was not found in standard install locations", e.Name, e.Version)
}

func IsExecutableNotFound(err error) bool {
	var notFoundError ExecutableNotFoundError
	return errors.As(err, &notFoundError)
}

// FindExecutable locates a DCC executable, optionally matching a version.
func FindExecutable(name, version string) (string, error) {
	version = strings.TrimSpace(version)
	if version != "" {
		return findVersionedExecutable(name, version)
	}
	return findExecutable(name)
}

func findExecutable(name string) (string, error) {
	environmentKey := strings.ToUpper(name) + "_PATH"
	if environmentPath := os.Getenv(environmentKey); environmentPath != "" {
		if _, err := os.Stat(environmentPath); err == nil {
			return environmentPath, nil
		}
	}
	for _, executablePath := range defaultPaths(name) {
		if _, err := os.Stat(executablePath); err == nil {
			return executablePath, nil
		}
	}
	if executablePath, err := exec.LookPath(name); err == nil {
		return executablePath, nil
	}
	return "", ExecutableNotFoundError{Name: name}
}

func findVersionedExecutable(name, version string) (string, error) {
	if !versionPattern.MatchString(version) {
		return "", fmt.Errorf("invalid %s version %q", name, version)
	}
	for _, executablePath := range versionedPaths(name, version) {
		if _, err := os.Stat(executablePath); err == nil {
			return executablePath, nil
		}
	}
	for _, commandName := range versionedCommands(name, version) {
		if executablePath, err := exec.LookPath(commandName); err == nil {
			return executablePath, nil
		}
	}
	return "", ExecutableNotFoundError{Name: name, Version: version}
}

func defaultPaths(name string) []string {
	switch runtime.GOOS {
	case "windows":
		return windowsPaths(name)
	case "darwin":
		return darwinPaths(name)
	default:
		return linuxPaths(name)
	}
}

func windowsPaths(name string) []string {
	programFiles := os.Getenv("ProgramFiles")
	if programFiles == "" {
		programFiles = `C:\Program Files`
	}
	switch strings.ToLower(name) {
	case "blender":
		var paths []string
		base := filepath.Join(programFiles, "Blender Foundation")
		if entries, err := os.ReadDir(base); err == nil {
			for index := len(entries) - 1; index >= 0; index-- {
				paths = append(paths, filepath.Join(base, entries[index].Name(), "blender.exe"))
			}
		}
		return paths
	case "maya":
		var paths []string
		base := filepath.Join(programFiles, "Autodesk")
		if entries, err := os.ReadDir(base); err == nil {
			for index := len(entries) - 1; index >= 0; index-- {
				if strings.HasPrefix(strings.ToLower(entries[index].Name()), "maya") {
					paths = append(paths, filepath.Join(base, entries[index].Name(), "bin", "maya.exe"))
				}
			}
		}
		return paths
	case "houdini":
		var paths []string
		base := filepath.Join(programFiles, "Side Effects Software")
		if entries, err := os.ReadDir(base); err == nil {
			for index := len(entries) - 1; index >= 0; index-- {
				paths = append(paths, filepath.Join(base, entries[index].Name(), "bin", "houdini.exe"))
			}
		}
		return paths
	default:
		return nil
	}
}

func darwinPaths(name string) []string {
	switch strings.ToLower(name) {
	case "blender":
		return []string{"/Applications/Blender.app/Contents/MacOS/Blender"}
	case "maya":
		var paths []string
		base := "/Applications/Autodesk"
		if entries, err := os.ReadDir(base); err == nil {
			for index := len(entries) - 1; index >= 0; index-- {
				if strings.HasPrefix(strings.ToLower(entries[index].Name()), "maya") {
					paths = append(paths, filepath.Join(base, entries[index].Name(), "Maya.app", "Contents", "bin", "maya"))
				}
			}
		}
		return paths
	case "houdini":
		return []string{"/Applications/Houdini/Current/Houdini FX.app/Contents/MacOS/houdini"}
	default:
		return nil
	}
}

func linuxPaths(name string) []string {
	switch strings.ToLower(name) {
	case "blender":
		return []string{"/usr/bin/blender", "/snap/bin/blender", "/usr/local/bin/blender"}
	case "maya":
		var paths []string
		base := "/usr/autodesk"
		if entries, err := os.ReadDir(base); err == nil {
			for index := len(entries) - 1; index >= 0; index-- {
				if strings.HasPrefix(strings.ToLower(entries[index].Name()), "maya") {
					paths = append(paths, filepath.Join(base, entries[index].Name(), "bin", "maya"))
				}
			}
		}
		return paths
	case "houdini":
		return []string{"/opt/hfs/bin/houdini"}
	default:
		return nil
	}
}

func versionedPaths(name, version string) []string {
	aliases := versionAliases(version)
	majorVersion := strings.Split(version, ".")[0]
	switch runtime.GOOS {
	case "windows":
		programFiles := os.Getenv("ProgramFiles")
		if programFiles == "" {
			programFiles = `C:\Program Files`
		}
		switch strings.ToLower(name) {
		case "blender":
			paths := make([]string, 0, len(aliases))
			for _, alias := range aliases {
				paths = append(paths, filepath.Join(programFiles, "Blender Foundation", "Blender "+alias, "blender.exe"))
			}
			return paths
		case "maya":
			return []string{filepath.Join(programFiles, fmt.Sprintf(`Autodesk\Maya%s\bin\maya.exe`, majorVersion))}
		}
	case "darwin":
		switch strings.ToLower(name) {
		case "blender":
			paths := make([]string, 0, len(aliases))
			for _, alias := range aliases {
				paths = append(paths, filepath.Join("/Applications", "Blender "+alias+".app", "Contents", "MacOS", "Blender"))
			}
			return paths
		case "maya":
			return []string{fmt.Sprintf("/Applications/Autodesk/maya%s/Maya.app/Contents/bin/maya", majorVersion)}
		}
	default:
		switch strings.ToLower(name) {
		case "blender":
			paths := make([]string, 0, len(aliases)*2)
			for _, alias := range aliases {
				paths = append(paths, "/opt/blender-"+alias+"/blender", "/usr/local/blender-"+alias+"/blender")
			}
			return paths
		case "maya":
			return []string{fmt.Sprintf("/usr/autodesk/maya%s/bin/maya", majorVersion)}
		}
	}
	return nil
}

func versionedCommands(name, version string) []string {
	commands := make([]string, 0, len(versionAliases(version)))
	for _, alias := range versionAliases(version) {
		commands = append(commands, name+"-"+alias)
	}
	return commands
}

func versionAliases(version string) []string {
	aliases := []string{version}
	trimmed := strings.TrimSuffix(version, ".0")
	if trimmed != version {
		aliases = append(aliases, trimmed)
	}
	return aliases
}
