package services

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

const (
	codeFileMaximumBytes = 5 * 1024 * 1024
	codeSearchLimit      = 200
)

type CodeFileType struct {
	ID         string   `json:"id"`
	LanguageID string   `json:"language_id"`
	Extensions []string `json:"extensions"`
	Filenames  []string `json:"filenames"`
	CanOpen    bool     `json:"can_open"`
	CanEdit    bool     `json:"can_edit"`
	CanDiff    bool     `json:"can_diff"`
	CanSearch  bool     `json:"can_search"`
}

type CodeFileClassification struct {
	Recognized bool   `json:"recognized"`
	Text       bool   `json:"text"`
	LanguageID string `json:"language_id"`
	Encoding   string `json:"encoding"`
	LineEnding string `json:"line_ending"`
	Editable   bool   `json:"editable"`
	Diffable   bool   `json:"diffable"`
	Searchable bool   `json:"searchable"`
	ReadOnly   bool   `json:"read_only"`
	Reason     string `json:"reason"`
	Size       int64  `json:"size"`
}

type CodeFileDocument struct {
	Path            string                 `json:"path"`
	Name            string                 `json:"name"`
	Content         string                 `json:"content"`
	ContentIdentity string                 `json:"content_identity"`
	Classification  CodeFileClassification `json:"classification"`
}

type CodeFileSearchResult struct {
	Path       string `json:"path"`
	Name       string `json:"name"`
	LanguageID string `json:"language_id"`
	Line       int    `json:"line,omitempty"`
	Preview    string `json:"preview,omitempty"`
}

var codeFileTypes = []CodeFileType{
	{ID: "plaintext", LanguageID: "plaintext", Extensions: []string{".txt", ".log", ".rst"}, CanOpen: true, CanEdit: true, CanDiff: true, CanSearch: true},
	{ID: "markdown", LanguageID: "markdown", Extensions: []string{".md"}, CanOpen: true, CanEdit: true, CanDiff: true, CanSearch: true},
	{ID: "json", LanguageID: "json", Extensions: []string{".json", ".jsonc"}, CanOpen: true, CanEdit: true, CanDiff: true, CanSearch: true},
	{ID: "yaml", LanguageID: "yaml", Extensions: []string{".yaml", ".yml"}, CanOpen: true, CanEdit: true, CanDiff: true, CanSearch: true},
	{ID: "configuration", LanguageID: "ini", Extensions: []string{".toml", ".ini", ".cfg", ".conf", ".env", ".editorconfig"}, Filenames: []string{".env", ".editorconfig", ".gitignore", ".gitattributes"}, CanOpen: true, CanEdit: true, CanDiff: true, CanSearch: true},
	{ID: "xml", LanguageID: "xml", Extensions: []string{".xml"}, CanOpen: true, CanEdit: true, CanDiff: true, CanSearch: true},
	{ID: "csv", LanguageID: "plaintext", Extensions: []string{".csv"}, CanOpen: true, CanEdit: true, CanDiff: true, CanSearch: true},
	{ID: "web", LanguageID: "html", Extensions: []string{".html", ".htm", ".vue"}, CanOpen: true, CanEdit: true, CanDiff: true, CanSearch: true},
	{ID: "styles", LanguageID: "css", Extensions: []string{".css", ".scss", ".less"}, CanOpen: true, CanEdit: true, CanDiff: true, CanSearch: true},
	{ID: "javascript", LanguageID: "javascript", Extensions: []string{".js", ".jsx", ".mjs", ".cjs"}, CanOpen: true, CanEdit: true, CanDiff: true, CanSearch: true},
	{ID: "typescript", LanguageID: "typescript", Extensions: []string{".ts", ".tsx"}, CanOpen: true, CanEdit: true, CanDiff: true, CanSearch: true},
	{ID: "python", LanguageID: "python", Extensions: []string{".py"}, CanOpen: true, CanEdit: true, CanDiff: true, CanSearch: true},
	{ID: "go", LanguageID: "go", Extensions: []string{".go"}, CanOpen: true, CanEdit: true, CanDiff: true, CanSearch: true},
	{ID: "rust", LanguageID: "rust", Extensions: []string{".rs"}, CanOpen: true, CanEdit: true, CanDiff: true, CanSearch: true},
	{ID: "cpp", LanguageID: "cpp", Extensions: []string{".c", ".h", ".cpp", ".hpp", ".cc"}, CanOpen: true, CanEdit: true, CanDiff: true, CanSearch: true},
	{ID: "csharp", LanguageID: "csharp", Extensions: []string{".cs"}, CanOpen: true, CanEdit: true, CanDiff: true, CanSearch: true},
	{ID: "java", LanguageID: "java", Extensions: []string{".java"}, CanOpen: true, CanEdit: true, CanDiff: true, CanSearch: true},
	{ID: "lua", LanguageID: "lua", Extensions: []string{".lua"}, CanOpen: true, CanEdit: true, CanDiff: true, CanSearch: true},
	{ID: "shader", LanguageID: "plaintext", Extensions: []string{".gd", ".shader", ".glsl", ".vert", ".frag", ".hlsl", ".cginc", ".usf", ".ush"}, CanOpen: true, CanEdit: true, CanDiff: true, CanSearch: true},
	{ID: "shell", LanguageID: "shell", Extensions: []string{".sh", ".ps1", ".bat", ".cmd"}, CanOpen: true, CanEdit: true, CanDiff: true, CanSearch: true},
	{ID: "build", LanguageID: "plaintext", Extensions: []string{".cmake"}, Filenames: []string{"Dockerfile", "Containerfile", "Makefile", "CMakeLists.txt"}, CanOpen: true, CanEdit: true, CanDiff: true, CanSearch: true},
}

func (f *FSService) ListCodeFileTypes() []CodeFileType {
	return append([]CodeFileType(nil), codeFileTypes...)
}

func (f *FSService) ClassifyCodeFile(path string) (CodeFileClassification, error) {
	canonicalPath, err := canonicalCodeFilePath(path)
	if err != nil {
		return CodeFileClassification{Reason: "outside_project"}, err
	}

	info, err := os.Stat(canonicalPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return CodeFileClassification{Reason: "missing"}, err
		}
		return CodeFileClassification{}, err
	}
	if info.IsDir() {
		return CodeFileClassification{Reason: "unsupported_type"}, nil
	}

	classification := CodeFileClassification{Size: info.Size()}
	fileType, recognized := findCodeFileType(filepath.Base(canonicalPath))
	classification.Recognized = recognized
	if recognized {
		classification.LanguageID = fileType.LanguageID
	}
	if info.Size() > codeFileMaximumBytes {
		classification.Reason = "file_too_large"
		return classification, nil
	}

	data, err := os.ReadFile(canonicalPath)
	if err != nil {
		if errors.Is(err, fs.ErrPermission) {
			classification.Reason = "permission_denied"
			return classification, nil
		}
		return classification, err
	}
	content, encoding, ok := decodeCodeFile(data)
	if !ok {
		classification.Reason = "binary_content"
		return classification, nil
	}

	classification.Text = true
	classification.Encoding = encoding
	classification.LineEnding = detectLineEnding(content)
	classification.Editable = fileType.CanEdit || !recognized
	classification.Diffable = fileType.CanDiff || !recognized
	classification.Searchable = fileType.CanSearch || !recognized
	classification.ReadOnly = info.Mode().Perm()&0200 == 0
	if classification.ReadOnly {
		classification.Editable = false
		classification.Reason = "permission_denied"
	}
	if !recognized {
		classification.LanguageID = "plaintext"
	}
	return classification, nil
}

func (f *FSService) ReadCodeFile(path string) (CodeFileDocument, error) {
	classification, err := f.ClassifyCodeFile(path)
	if err != nil {
		return CodeFileDocument{}, err
	}
	if !classification.Text {
		return CodeFileDocument{}, fmt.Errorf("file is not editable text: %s", classification.Reason)
	}

	canonicalPath, err := canonicalCodeFilePath(path)
	if err != nil {
		return CodeFileDocument{}, err
	}
	data, err := os.ReadFile(canonicalPath)
	if err != nil {
		return CodeFileDocument{}, err
	}
	content, _, _ := decodeCodeFile(data)
	return CodeFileDocument{
		Path:            canonicalPath,
		Name:            filepath.Base(canonicalPath),
		Content:         content,
		ContentIdentity: codeContentIdentity(data),
		Classification:  classification,
	}, nil
}

func (f *FSService) WriteCodeFile(path, content, expectedIdentity string) (CodeFileDocument, error) {
	document, err := f.ReadCodeFile(path)
	if err != nil {
		return CodeFileDocument{}, err
	}
	if !document.Classification.Editable {
		return CodeFileDocument{}, fmt.Errorf("file is not editable: %s", document.Classification.Reason)
	}
	if document.ContentIdentity != expectedIdentity {
		return CodeFileDocument{}, errors.New("file changed since it was opened")
	}

	encoded := encodeCodeFile(content, document.Classification.Encoding)
	temporary, err := os.CreateTemp(filepath.Dir(document.Path), ".clustta-code-*.tmp")
	if err != nil {
		return CodeFileDocument{}, err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err := temporary.Write(encoded); err != nil {
		temporary.Close()
		return CodeFileDocument{}, err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return CodeFileDocument{}, err
	}
	if err := temporary.Close(); err != nil {
		return CodeFileDocument{}, err
	}
	info, err := os.Stat(document.Path)
	if err != nil {
		return CodeFileDocument{}, err
	}
	if err := os.Chmod(temporaryPath, info.Mode()); err != nil {
		return CodeFileDocument{}, err
	}
	if err := replaceCodeFile(temporaryPath, document.Path); err != nil {
		return CodeFileDocument{}, err
	}
	return f.ReadCodeFile(document.Path)
}

func (f *FSService) SearchCodeFiles(root, query string, maximumResults int) ([]CodeFileSearchResult, error) {
	canonicalRoot, err := canonicalCodeFilePath(root)
	if err != nil {
		return nil, err
	}
	if maximumResults <= 0 || maximumResults > codeSearchLimit {
		maximumResults = codeSearchLimit
	}
	query = strings.ToLower(strings.TrimSpace(query))
	results := make([]CodeFileSearchResult, 0)
	err = filepath.WalkDir(canonicalRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || len(results) >= maximumResults {
			return walkErr
		}
		if entry.IsDir() {
			if path != canonicalRoot && strings.HasPrefix(entry.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		fileType, recognized := findCodeFileType(entry.Name())
		if !recognized {
			return nil
		}
		result := CodeFileSearchResult{Path: path, Name: entry.Name(), LanguageID: fileType.LanguageID}
		if query != "" && !strings.Contains(strings.ToLower(entry.Name()), query) {
			info, infoErr := entry.Info()
			if infoErr != nil || info.Size() > codeFileMaximumBytes {
				return nil
			}
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return nil
			}
			content, _, valid := decodeCodeFile(data)
			if !valid {
				return nil
			}
			line, preview, found := findCodeSearchMatch(content, query)
			if !found {
				return nil
			}
			result.Line = line
			result.Preview = preview
		}
		results = append(results, result)
		return nil
	})
	return results, err
}

func findCodeFileType(name string) (CodeFileType, bool) {
	for _, fileType := range codeFileTypes {
		for _, filename := range fileType.Filenames {
			if strings.EqualFold(name, filename) {
				return fileType, true
			}
		}
	}

	lowerName := strings.ToLower(name)
	types := append([]CodeFileType(nil), codeFileTypes...)
	sort.SliceStable(types, func(i, j int) bool {
		return longestExtension(types[i]) > longestExtension(types[j])
	})
	for _, fileType := range types {
		for _, extension := range fileType.Extensions {
			if strings.HasSuffix(lowerName, extension) {
				return fileType, true
			}
		}
	}
	return CodeFileType{}, false
}

func longestExtension(fileType CodeFileType) int {
	longest := 0
	for _, extension := range fileType.Extensions {
		if len(extension) > longest {
			longest = len(extension)
		}
	}
	return longest
}

func canonicalCodeFilePath(path string) (string, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	canonicalPath, err := filepath.EvalSymlinks(absolutePath)
	if err != nil {
		return "", err
	}
	if err := validatePath(canonicalPath); err != nil {
		return "", err
	}
	return canonicalPath, nil
}

func decodeCodeFile(data []byte) (string, string, bool) {
	encoding := "utf-8"
	if bytes.HasPrefix(data, []byte{0xef, 0xbb, 0xbf}) {
		data = data[3:]
		encoding = "utf-8-bom"
	}
	if bytes.HasPrefix(data, []byte{0xff, 0xfe}) {
		return decodeUTF16(data[2:], binary.LittleEndian, "utf-16-le")
	}
	if bytes.HasPrefix(data, []byte{0xfe, 0xff}) {
		return decodeUTF16(data[2:], binary.BigEndian, "utf-16-be")
	}
	if bytes.IndexByte(data, 0) >= 0 {
		return "", "", false
	}
	if !utf8.Valid(data) {
		return "", "", false
	}
	return string(data), encoding, true
}

func decodeUTF16(data []byte, byteOrder binary.ByteOrder, encoding string) (string, string, bool) {
	if len(data)%2 != 0 {
		return "", "", false
	}
	codeUnits := make([]uint16, len(data)/2)
	for index := range codeUnits {
		codeUnits[index] = byteOrder.Uint16(data[index*2:])
	}
	return string(utf16.Decode(codeUnits)), encoding, true
}

func encodeCodeFile(content, encoding string) []byte {
	switch encoding {
	case "utf-8-bom":
		return append([]byte{0xef, 0xbb, 0xbf}, []byte(content)...)
	case "utf-16-le":
		return encodeUTF16(content, binary.LittleEndian, []byte{0xff, 0xfe})
	case "utf-16-be":
		return encodeUTF16(content, binary.BigEndian, []byte{0xfe, 0xff})
	default:
		return []byte(content)
	}
}

func encodeUTF16(content string, byteOrder binary.ByteOrder, prefix []byte) []byte {
	codeUnits := utf16.Encode([]rune(content))
	data := make([]byte, len(prefix)+(len(codeUnits)*2))
	copy(data, prefix)
	for index, codeUnit := range codeUnits {
		byteOrder.PutUint16(data[len(prefix)+(index*2):], codeUnit)
	}
	return data
}

func detectLineEnding(content string) string {
	if strings.Contains(content, "\r\n") {
		return "CRLF"
	}
	return "LF"
}

func codeContentIdentity(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func findCodeSearchMatch(content, query string) (int, string, bool) {
	for index, line := range strings.Split(content, "\n") {
		if strings.Contains(strings.ToLower(line), query) {
			return index + 1, strings.TrimSpace(line), true
		}
	}
	return 0, "", false
}
