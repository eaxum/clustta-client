package utils

import (
	"bytes"
	"clustta/internal/auth_service"
	"clustta/internal/settings"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"time"
	"unicode"

	// "github.com/cespare/xxhash/v2"

	"github.com/jmoiron/sqlx"
	"github.com/nfnt/resize"
	"github.com/zeebo/xxh3"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type ProjectContext struct {
	ProjectPath string
	DbConn      *sqlx.DB
	Tx          *sqlx.Tx
	UseTx       bool
}

func OpenDb(dbPath string) (*sqlx.DB, error) {
	// dbConn, err := sqlx.Open("sqlite3", dbPath+"?_busy_timeout=30000")
	dbConn, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		return dbConn, err
	}

	// Enable WAL mode for better performance on large databases
	_, err = dbConn.Exec("PRAGMA journal_mode = WAL;")
	if err != nil {
		return dbConn, err
	}

	// Set synchronous mode to NORMAL for better performance (still safe with WAL)
	_, err = dbConn.Exec("PRAGMA synchronous = NORMAL;")
	if err != nil {
		return dbConn, err
	}

	// Increase busy timeout for large database operations
	_, err = dbConn.Exec("PRAGMA busy_timeout = 120000;")
	if err != nil {
		return dbConn, err
	}

	// Set WAL auto-checkpoint to smaller intervals for better commit performance
	_, err = dbConn.Exec("PRAGMA wal_autocheckpoint = 100;")
	if err != nil {
		return dbConn, err
	}

	return dbConn, err
}

func GenerateXXHashChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash_function := xxh3.New()

	// Read the file in chunks to conserve memory
	for {
		chunk := make([]byte, 16*1024)
		n, err := file.Read(chunk)
		if err != nil && err != io.EOF {
			return "", err
		}
		if n == 0 {
			break
		}
		hash_function.Write(chunk[:n])
	}

	hexHash := make([]byte, hex.EncodedLen(8))
	hex.Encode(hexHash, hash_function.Sum(nil))
	return string(hexHash), nil
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func GetCurrentTime() string {
	return time.Now().Format(time.RFC3339)
}

func GetEpochTime() int64 {
	return time.Now().Unix()
}

func EpochToRFC3339(epoch int64) string {
	// Convert epoch time to a Time object
	// t := time.Unix(epoch, 0).UTC()
	t := time.Unix(epoch, 0).In(time.Local)

	// Convert Time object to RFC3339 format
	rfc3339 := t.Format(time.RFC3339)

	return rfc3339
}

func RFC3339ToEpoch(rfc3339 string) (int64, error) {
	// Parse the RFC3339 formatted string to a Time object
	t, err := time.Parse(time.RFC3339, rfc3339)
	if err != nil {
		return 0, err
	}

	// Convert the Time object to epoch time
	epoch := t.Unix()

	return epoch, nil
}

func RevealInExplorer(filePath string) {
	if runtime.GOOS == "windows" {
		path := strings.Replace(filePath, "/", "\\", -1)
		exec.Command("explorer", "/select,", path).Start()
	} else if runtime.GOOS == "darwin" {
		exec.Command("open", "-R", filePath).Start()
	} else {
		exec.Command("xdg-open", "--select", filePath).Start()
	}
}

func LaunchFile(filePath string) error {
	if runtime.GOOS == "windows" {
		err := exec.Command("cmd", "/C", "start", filePath).Start()
		if err != nil {
			return err
		}
	} else if runtime.GOOS == "darwin" {
		err := exec.Command("open", filePath).Start()
		if err != nil {
			return err
		}
	} else {
		err := exec.Command("xdg-open", filePath).Start()
		if err != nil {
			return err
		}
	}
	return nil
}

func FileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		// panic(err)
		return false
	}
	return !info.IsDir()
}

func DirExists(dir string) bool {
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func ExpandPath(path string) (string, error) {
	usr, _ := user.Current()
	dir := usr.HomeDir
	if path == "~" {
		// In case of "~", which won't be caught by the "else if"
		path = dir
	} else if strings.HasPrefix(path, "~/") {
		// Use strings.HasPrefix so we don't match paths like
		// "/something/~/something/"
		path = filepath.Join(dir, path[2:])
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path, err
	}
	return absPath, nil
}

func ToSnakeCase(text string) string {
	var result strings.Builder
	lastWasUpper := false

	for _, char := range text {
		if unicode.IsUpper(char) {
			if !lastWasUpper {
				if result.Len() > 0 {
					result.WriteByte('_')
				}
			}
			result.WriteRune(unicode.ToLower(char))
			lastWasUpper = true
		} else if unicode.IsLetter(char) || unicode.IsDigit(char) {
			result.WriteRune(char)
			lastWasUpper = false
		} else {
			if result.Len() > 0 && !lastWasUpper {
				result.WriteByte('_')
			}
			lastWasUpper = false
		}
	}

	return result.String()
}

func StructToMap(data interface{}, snake_case bool) map[string]interface{} {
	result := make(map[string]interface{})
	val := reflect.ValueOf(data)
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i).Interface()
		if snake_case {
			result[ToSnakeCase(field.Name)] = value
		} else {
			result[field.Name] = value
		}
	}

	return result
}

func Contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}
func NonCaseSensitiveContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// IsValidURL checks if the input string is a valid URL
func IsValidURL(link string) bool {
	parsedURL, err := url.ParseRequestURI(link)
	if err != nil {
		return false
	}

	return parsedURL.Scheme == "http" || parsedURL.Scheme == "https"
}

func IsValidPointer(pointer string) bool {
	if !IsValidURL(pointer) && !FileExists(pointer) {
		return false
	}
	return true
}

func GetProjectFile(tx *sqlx.Tx) (string, error) {
	// var name string
	// err := tx.Get(&name, "SELECT value FROM config WHERE name = 'name'")
	// if err != nil {
	// 	return "", err
	// }
	// return name, nil
	var filePath string
	query := "SELECT file FROM pragma_database_list WHERE name = 'main'"
	err := tx.Get(&filePath, query)
	if err != nil {
		return "", err
	}
	return filePath, nil
}
func GetProjectName(tx *sqlx.Tx) (string, error) {
	// var name string
	// err := tx.Get(&name, "SELECT value FROM config WHERE name = 'name'")
	// if err != nil {
	// 	return "", err
	// }
	// return name, nil
	var filePath string
	query := "SELECT file FROM pragma_database_list WHERE name = 'main'"
	err := tx.Get(&filePath, query)
	if err != nil {
		return "", err
	}
	fileName := strings.Split(filepath.Base(filePath), ".")[0]
	return fileName, nil
}

func GetProjectFolder(tx *sqlx.Tx) (string, error) {
	// var name string
	// err := tx.Get(&name, "SELECT value FROM config WHERE name = 'name'")
	// if err != nil {
	// 	return "", err
	// }
	// return name, nil
	var filePath string
	query := "SELECT file FROM pragma_database_list WHERE name = 'main'"
	err := tx.Get(&filePath, query)
	if err != nil {
		return "", err
	}
	return filepath.Dir(filePath), nil
}

//	func GetLocalProjectWorkingDir(tx *sqlx.Tx) (string, error) {
//		workingDir, err := settings.GetLocalWorkingDirectory()
//		if err != nil {
//			return "", err
//		}
//		studioName, err := GetStudioName(tx)
//		if err != nil {
//			return "", err
//		}
//		projectName, err := GetProjectName(tx)
//		if err != nil {
//			return "", err
//		}
//		filePath := filepath.Join(workingDir, studioName, projectName)
//		return filePath, nil
//	}
func GetOldProjectWorkingDir(tx *sqlx.Tx, user auth_service.User) (string, error) {
	workingDir, err := settings.GetUserDataFolder(user)
	if err != nil {
		return "", err
	}
	studioName, err := GetStudioName(tx)
	if err != nil {
		return "", err
	}
	projectName, err := GetProjectName(tx)
	if err != nil {
		return "", err
	}
	filePath := filepath.Join(workingDir, studioName, projectName)
	return filePath, nil
}

func Slugify(text, separator string, lowercase bool) string {
	const defaultSeparator = "-"
	if separator == "" {
		separator = defaultSeparator
	}

	quotePattern := regexp.MustCompile(`[']+`)
	allowedCharsPattern := regexp.MustCompile(`[^-a-z0-9]+`)
	duplicateDashPattern := regexp.MustCompile(`-{2,}`)
	// numbersPattern := regexp.MustCompile(`(?<=\d),(?=\d)`)

	// Replace quotes with dashes - pre-process
	text = quotePattern.ReplaceAllString(text, separator)

	// Make the text lowercase (optional)
	if lowercase {
		text = strings.ToLower(text)
	}

	// Remove generated quotes -- post-process
	text = quotePattern.ReplaceAllString(text, "")

	// Cleanup numbers
	// text = numbersPattern.ReplaceAllString(text, "")

	// Replace disallowed characters
	text = allowedCharsPattern.ReplaceAllString(text, defaultSeparator)

	// Remove redundant dashes
	text = duplicateDashPattern.ReplaceAllString(text, defaultSeparator)
	text = strings.Trim(text, defaultSeparator)

	// Replace default separator with custom separator if needed
	if separator != defaultSeparator {
		text = strings.ReplaceAll(text, defaultSeparator, separator)
	}

	return text
}

func MatchesRegex(path string, pattern string) (bool, error) {
	// Normalize path separators for the current OS
	path = filepath.ToSlash(path)
	// Compile the regex
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}
	// Match the path against the regex
	return re.MatchString(path), nil
}

func CreateSchema(db *sqlx.DB, schema string) error {
	statements := SplitStatements(schema)
	for _, statement := range statements {
		_, err := db.Exec(statement)
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateSchemaTx(tx *sqlx.Tx, schema string) error {
	statements := SplitStatements(schema)
	for _, statement := range statements {
		_, err := tx.Exec(statement)
		if err != nil {
			return err
		}
	}
	return nil
}

func FormatDateTime(t time.Time, timezone *time.Location, locale language.Tag) string {
	// Convert to specified timezone
	localTime := t.In(timezone)

	// Create a localized printer for formatting
	p := message.NewPrinter(locale)

	// Format the time using a localized format
	// You can customize the format string based on locale preferences
	return p.Sprintf("%s", localTime.Format(time.RFC3339))
}

// Helper function to resize and return the image bytes
func ResizeImage(fileBytes []byte, width, height uint) ([]byte, error) {
	// Detect image type
	img, format, err := image.Decode(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	// Resize image to specified width and height
	resizedImg := resize.Resize(width, height, img, resize.Lanczos3)

	// Buffer to hold the resized image
	var buf bytes.Buffer

	// Encode resized image back into appropriate format
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, resizedImg, nil)
	case "png":
		err = png.Encode(&buf, resizedImg)
	default:
		return nil, fmt.Errorf("unsupported image format: %s", format)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode resized image: %v", err)
	}

	return buf.Bytes(), nil
}

func BytesToHumanReadable(bytes int) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d Bytes", bytes)
	}
}

func NormalizePath(path string) string {
	// First convert backslashes to forward slashes
	path = strings.ReplaceAll(path, "\\", "/")

	// Remove duplicate slashes
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}

	// Use filepath.Clean to handle . and .. components
	// This also normalizes path separators to the OS-specific separator
	// path = filepath.Clean(path)

	return path
}

func RemoveDuplicates(slice []string) []string {
	seen := make(map[string]struct{})
	result := []string{}

	for _, item := range slice {
		if _, ok := seen[item]; !ok {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func GetParent(p string) string {
	p, _ = strings.CutSuffix(p, "/")
	p, _ = strings.CutPrefix(p, "/")
	dir := path.Dir(p)
	if dir == "." {
		return ""
	}
	return "/" + dir + "/"
}

func GetEntityPaths(entityPath string) []string {
	entityPath, _ = strings.CutSuffix(entityPath, "/")
	entityPath, _ = strings.CutPrefix(entityPath, "/")
	parts := strings.Split(entityPath, "/")
	var entityPaths []string
	var current string

	if entityPath == "" {
		return []string{}
	}

	for _, part := range parts {
		if current == "" {
			current = "/" + part
		} else {
			current = path.Join(current, part)
		}
		current = current + "/"
		entityPaths = append(entityPaths, current)
	}

	return entityPaths
}
