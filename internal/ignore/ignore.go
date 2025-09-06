package ignore

import (
	"os"
	"regexp"
	"strings"
)

////////////////////////////////////////////////////////////

// IgnoreParser is an interface with `MatchesPaths`.
type IgnoreParser interface {
	MatchesPath(f string) bool
	MatchesPathHow(f string) (bool, *IgnorePattern)
}

////////////////////////////////////////////////////////////

// This function pretty much attempts to mimic the parsing rules
// listed above at the start of this file
func getPatternFromLine(line string) (*regexp.Regexp, bool) {
	// Trim OS-specific carriage returns.
	line = strings.TrimRight(line, "\r")

	// Strip comments [Rule 2]
	if strings.HasPrefix(line, `#`) {
		return nil, false
	}

	// Trim string [Rule 3]
	// TODO: Handle [Rule 3], when the " " is escaped with a \
	line = strings.Trim(line, " ")

	// Exit for no-ops and return nil which will prevent us from
	// appending a pattern against this line
	if line == "" {
		return nil, false
	}

	// TODO: Handle [Rule 4] which negates the match for patterns leading with "!"
	negatePattern := false
	if line[0] == '!' {
		negatePattern = true
		line = line[1:]
	}

	// Handle [Rule 2, 4], when # or ! is escaped with a \
	// Handle [Rule 4] once we tag negatePattern, strip the leading ! char
	if regexp.MustCompile(`^(\#|\!)`).MatchString(line) {
		line = line[1:]
	}

	// If we encounter a foo/*.blah in a folder, prepend the / char
	if regexp.MustCompile(`([^\/+])/.*\*\.`).MatchString(line) && line[0] != '/' {
		line = "/" + line
	}

	// Handle escaping the "." char
	line = regexp.MustCompile(`\.`).ReplaceAllString(line, `\.`)

	magicStar := "#$~"

	// Handle "/**/" usage
	if strings.HasPrefix(line, "/**/") {
		line = line[1:]
	}
	line = regexp.MustCompile(`/\*\*/`).ReplaceAllString(line, `(/|/.+/)`)
	line = regexp.MustCompile(`\*\*/`).ReplaceAllString(line, `(|.`+magicStar+`/)`)
	line = regexp.MustCompile(`/\*\*`).ReplaceAllString(line, `(|/.`+magicStar+`)`)

	// Handle escaping the "*" char
	line = regexp.MustCompile(`\\\*`).ReplaceAllString(line, `\`+magicStar)
	line = regexp.MustCompile(`\*`).ReplaceAllString(line, `([^/]*)`)

	// Handle escaping the "?" char
	line = strings.Replace(line, "?", `\?`, -1)

	line = strings.Replace(line, magicStar, "*", -1)

	// Temporary regex
	var expr = ""
	if strings.HasSuffix(line, "/") {
		expr = line + "(|.*)$"
	} else {
		expr = line + "(|/.*)$"
	}
	if strings.HasPrefix(expr, "/") {
		expr = "^(|/)" + expr[1:]
	} else {
		expr = "^(|.*/)" + expr
	}
	pattern, _ := regexp.Compile(expr)

	return pattern, negatePattern
}

////////////////////////////////////////////////////////////

// IgnorePattern encapsulates a pattern and if it is a negated pattern.
type IgnorePattern struct {
	Pattern *regexp.Regexp
	Negate  bool
	LineNo  int
	Line    string
}

// GitIgnore wraps a list of ignore pattern.
type GitIgnore struct {
	patterns []*IgnorePattern
}

// CompileIgnoreLines accepts a variadic set of strings, and returns a GitIgnore
// instance which converts and appends the lines in the input to regexp.Regexp
// patterns held within the GitIgnore objects "patterns" field.
func CompileIgnoreLines(lines ...string) *GitIgnore {
	gi := &GitIgnore{}
	for i, line := range lines {
		pattern, negatePattern := getPatternFromLine(line)
		if pattern != nil {
			// LineNo is 1-based numbering to match `git check-ignore -v` output
			ip := &IgnorePattern{pattern, negatePattern, i + 1, line}
			gi.patterns = append(gi.patterns, ip)
		}
	}
	return gi
}

////////////////////////////////////////////////////////////

// MatchesPath returns true if the given GitIgnore structure would target
// a given path string `f`.
func (gi *GitIgnore) MatchesPath(f string) bool {
	matchesPath, _ := gi.MatchesPathHow(f)
	return matchesPath
}

// MatchesPathHow returns true, `pattern` if the given GitIgnore structure would target
// a given path string `f`.
// The IgnorePattern has the Line, LineNo fields.
func (gi *GitIgnore) MatchesPathHow(f string) (bool, *IgnorePattern) {
	// Replace OS-specific path separator.
	f = strings.Replace(f, string(os.PathSeparator), "/", -1)

	matchesPath := false
	var mip *IgnorePattern
	for _, ip := range gi.patterns {
		if ip.Pattern.MatchString(f) {
			// If this is a regular target (not negated with a gitignore
			// exclude "!" etc)
			if !ip.Negate {
				matchesPath = true
				mip = ip
			} else if matchesPath {
				// Negated pattern, and matchesPath is already set
				matchesPath = false
			}
		}
	}
	return matchesPath, mip
}
