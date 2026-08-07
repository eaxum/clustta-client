package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

type replacement struct {
	path    string
	pattern string
	replace string
}

func main() {
	versionFlag := flag.String("version", "", "release version, for example 0.4.36 or v0.4.36")
	dateFlag := flag.String("date", time.Now().Format("2006-01-02"), "release date in YYYY-MM-DD format")
	flag.Parse()

	version := strings.TrimSpace(*versionFlag)
	version = strings.TrimPrefix(version, "v")
	if version == "" {
		fatalf("version is required, for example: go run ./cmd/release-version -version 0.4.36")
	}
	if !regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(version) {
		fatalf("version must look like 0.4.36, got %q", version)
	}
	if !regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`).MatchString(*dateFlag) {
		fatalf("date must look like 2026-07-02, got %q", *dateFlag)
	}

	plistVersion := macOSBundleVersion(version)
	msixVersion := msixPackageVersion(version)
	dockerVersion := version
	flatpakReleaseTag := "v" + version
	flatpakTag := "v" + version + "-flatpak.1"

	updateConfigVersion(version)

	changes := []replacement{
		{
			path:    "internal/constants/constants.go",
			pattern: `(const fallbackVersion = ")[^"]+(")`,
			replace: "${1}" + version + "${2}",
		},
		{
			path:    "build/windows/wails.exe.manifest",
			pattern: `(name="com\.clustta\.clustta" version=")[^"]+(")`,
			replace: "${1}" + version + "${2}",
		},
		{
			path:    "build/windows/info.json",
			pattern: `("file_version": ")[^"]+(")`,
			replace: "${1}" + version + "${2}",
		},
		{
			path:    "build/windows/info.json",
			pattern: `("ProductVersion": ")[^"]+(")`,
			replace: "${1}" + version + "${2}",
		},
		{
			path:    "build/windows/nsis/wails_tools.nsh",
			pattern: `(!define INFO_PRODUCTVERSION ")[^"]+(")`,
			replace: "${1}" + version + "${2}",
		},
		{
			path:    "build/windows/Taskfile.yml",
			pattern: `(?m)^(\s*APP_VERSION:\s*")[^"]+(")(\r?)$`,
			replace: "${1}" + version + "${2}${3}",
		},
		{
			path:    "build/darwin/Taskfile.yml",
			pattern: `(?m)^(\s*APP_VERSION:\s*")[^"]+(")(\r?)$`,
			replace: "${1}" + version + "${2}${3}",
		},
		{
			path:    "build/linux/Taskfile.yml",
			pattern: `(?m)^(\s*APP_VERSION:\s*")[^"]+(")(\r?)$`,
			replace: "${1}" + version + "${2}${3}",
		},
		{
			path:    "build/darwin/Info.plist",
			pattern: `(?s)(<key>CFBundleShortVersionString</key>\s*<string>)[^<]+(</string>)`,
			replace: "${1}" + plistVersion + "${2}",
		},
		{
			path:    "build/darwin/Info.plist",
			pattern: `(?s)(<key>CFBundleVersion</key>\s*<string>)[^<]+(</string>)`,
			replace: "${1}" + plistVersion + "${2}",
		},
		{
			path:    "build/darwin/Info.dev.plist",
			pattern: `(?s)(<key>CFBundleShortVersionString</key>\s*<string>)[^<]+(</string>)`,
			replace: "${1}" + plistVersion + "${2}",
		},
		{
			path:    "build/darwin/Info.dev.plist",
			pattern: `(?s)(<key>CFBundleVersion</key>\s*<string>)[^<]+(</string>)`,
			replace: "${1}" + plistVersion + "${2}",
		},
		{
			path:    "build/linux/nfpm/nfpm.yaml",
			pattern: `(?m)^(version:\s*)"?[^"\r\n]+"?(\r?)$`,
			replace: "${1}\"" + version + "\"${2}",
		},
		{
			path:    "build/linux/clustta.metainfo.xml",
			pattern: `(<release version=")[^"]+(" date=")[^"]+(">)`,
			replace: "${1}" + version + "${2}" + *dateFlag + "${3}",
		},
		{
			path:    "build/linux/flatpak/com.clustta.clustta.metainfo.xml",
			pattern: `(<release version=")[^"]+(" date=")[^"]+(">)`,
			replace: "${1}" + version + "${2}" + *dateFlag + "${3}",
		},
		{
			path:    "build/linux/flatpak/com.clustta.clustta.yml",
			pattern: `(releases/download/)v[0-9]+\.[0-9]+\.[0-9]+(/clustta-)[0-9]+\.[0-9]+\.[0-9]+(-linux-amd64)`,
			replace: "${1}" + flatpakReleaseTag + "${2}" + version + "${3}",
		},
		{
			path:    "commands.txt",
			pattern: `(clustta/clustta:)[0-9]+\.[0-9]+\.[0-9]+`,
			replace: "${1}" + dockerVersion,
		},
		{
			path:    "Makefile",
			pattern: `make flatpak-tag VERSION=v[0-9]+\.[0-9]+\.[0-9]+-flatpak\.1`,
			replace: "make flatpak-tag VERSION=" + flatpakTag,
		},
	}

	if _, err := os.Stat("Clustta_template.xml"); err == nil {
		changes = append(changes, replacement{
			path:    "Clustta_template.xml",
			pattern: `(PackageName="Eaxum\.Clustta"[^>]* Version=")[^"]+(")`,
			replace: "${1}" + msixVersion + "${2}",
		})
	} else if !os.IsNotExist(err) {
		fatalf("stat Clustta_template.xml: %v", err)
	}

	for _, change := range changes {
		apply(change)
	}

	fmt.Printf("Stamped Clustta %s metadata for %s\n", version, *dateFlag)
	fmt.Printf("Stamped macOS bundle version as %s\n", plistVersion)
	fmt.Printf("Stamped MSIX package version as %s\n", msixVersion)
	fmt.Println("Note: Flatpak source URL/SHA still need updating after the Linux release binary exists.")
}

func apply(change replacement) {
	input, err := os.ReadFile(change.path)
	if err != nil {
		fatalf("read %s: %v", change.path, err)
	}

	re := regexp.MustCompile(change.pattern)
	if !re.Match(input) {
		fatalf("no match while updating %s with pattern %s", change.path, change.pattern)
	}
	output := re.ReplaceAllString(string(input), change.replace)

	if err := os.WriteFile(change.path, []byte(output), 0o644); err != nil {
		fatalf("write %s: %v", change.path, err)
	}
}

func updateConfigVersion(version string) {
	path := "build/config.yml"
	input, err := os.ReadFile(path)
	if err != nil {
		fatalf("read %s: %v", path, err)
	}

	lines := strings.SplitAfter(string(input), "\n")
	updated := false
	inInfo := false
	for i, line := range lines {
		body, lineEnding := splitLineEnding(line)
		trimmed := strings.TrimSpace(body)

		if trimmed == "info:" {
			inInfo = true
			continue
		}
		if inInfo && body != "" && !strings.HasPrefix(body, " ") && !strings.HasPrefix(body, "\t") {
			inInfo = false
		}
		if inInfo && strings.HasPrefix(strings.TrimLeft(body, " \t"), "version:") {
			indent := body[:len(body)-len(strings.TrimLeft(body, " \t"))]
			lines[i] = indent + `version: "` + version + `"` + lineEnding
			updated = true
			break
		}
	}
	if !updated {
		fatalf("no info.version line found in %s", path)
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, "")), 0o644); err != nil {
		fatalf("write %s: %v", path, err)
	}
}

func splitLineEnding(line string) (string, string) {
	lineEnding := ""
	if strings.HasSuffix(line, "\n") {
		lineEnding = "\n"
		line = strings.TrimSuffix(line, "\n")
	}
	if strings.HasSuffix(line, "\r") {
		lineEnding = "\r" + lineEnding
		line = strings.TrimSuffix(line, "\r")
	}
	return line, lineEnding
}

func macOSBundleVersion(version string) string {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return version
	}
	return parts[0] + "." + parts[1] + "." + parts[2] + "0"
}

func msixPackageVersion(version string) string {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return version
	}
	return parts[0] + "." + parts[1] + "." + parts[2] + ".0"
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "release-version: "+format+"\n", args...)
	os.Exit(1)
}
