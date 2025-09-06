// Implement tests for the `ignore` library
package ignore

import (
	"fmt"
	"testing"

	"runtime"

	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////

// Validate "CompileIgnoreLines()"
func TestCompileIgnoreLines(t *testing.T) {
	lines := []string{"abc/def", "a/b/c", "b"}
	object := CompileIgnoreLines(lines...)

	// MatchesPath
	// Paths which are targeted by the above "lines"
	assert.Equal(t, true, object.MatchesPath("abc/def/child"), "abc/def/child should match")
	assert.Equal(t, true, object.MatchesPath("a/b/c/d"), "a/b/c/d should match")

	// Paths which are not targeted by the above "lines"
	assert.Equal(t, false, object.MatchesPath("abc"), "abc should not match")
	assert.Equal(t, false, object.MatchesPath("def"), "def should not match")
	assert.Equal(t, false, object.MatchesPath("bd"), "bd should not match")

	object = CompileIgnoreLines("abc/def", "a/b/c", "b")

	// Paths which are targeted by the above "lines"
	assert.Equal(t, true, object.MatchesPath("abc/def/child"), "abc/def/child should match")
	assert.Equal(t, true, object.MatchesPath("a/b/c/d"), "a/b/c/d should match")

	// Paths which are not targeted by the above "lines"
	assert.Equal(t, false, object.MatchesPath("abc"), "abc should not match")
	assert.Equal(t, false, object.MatchesPath("def"), "def should not match")
	assert.Equal(t, false, object.MatchesPath("bd"), "bd should not match")
}

func ExampleCompileIgnoreLines() {
	ignoreObject := CompileIgnoreLines([]string{"node_modules", "*.out", "foo/*.c"}...)

	// You can test the ignoreObject against various paths using the
	// "MatchesPath()" interface method. This pretty much is up to
	// the users interpretation. In the case of a ".gitignore" file,
	// a "match" would indicate that a given path would be ignored.
	fmt.Println(ignoreObject.MatchesPath("node_modules/test/foo.js"))
	fmt.Println(ignoreObject.MatchesPath("node_modules2/test.out"))
	fmt.Println(ignoreObject.MatchesPath("test/foo.js"))

	// Output:
	// true
	// true
	// false
}

func TestCompileIgnoreLines_CheckNestedDotFiles(t *testing.T) {
	lines := []string{
		"**/external/**/*.md",
		"**/external/**/*.json",
		"**/external/**/*.gzip",
		"**/external/**/.*ignore",

		"**/external/foobar/*.css",
		"**/external/barfoo/less",
		"**/external/barfoo/scss",
	}
	object := CompileIgnoreLines(lines...)
	assert.NotNil(t, object, "returned object should not be nil")

	assert.Equal(t, true, object.MatchesPath("external/foobar/angular.foo.css"), "external/foobar/angular.foo.css")
	assert.Equal(t, true, object.MatchesPath("external/barfoo/.gitignore"), "external/barfoo/.gitignore")
	assert.Equal(t, true, object.MatchesPath("external/barfoo/.bower.json"), "external/barfoo/.bower.json")
}

func TestCompileIgnoreLines_CarriageReturn(t *testing.T) {
	lines := []string{"abc/def\r", "a/b/c\r", "b\r"}
	object := CompileIgnoreLines(lines...)

	assert.Equal(t, true, object.MatchesPath("abc/def/child"), "abc/def/child should match")
	assert.Equal(t, true, object.MatchesPath("a/b/c/d"), "a/b/c/d should match")

	assert.Equal(t, false, object.MatchesPath("abc"), "abc should not match")
	assert.Equal(t, false, object.MatchesPath("def"), "def should not match")
	assert.Equal(t, false, object.MatchesPath("bd"), "bd should not match")
}

func TestCompileIgnoreLines_WindowsPath(t *testing.T) {
	if runtime.GOOS != "windows" {
		return
	}
	lines := []string{"abc/def", "a/b/c", "b"}
	object := CompileIgnoreLines(lines...)

	assert.Equal(t, true, object.MatchesPath("abc\\def\\child"), "abc\\def\\child should match")
	assert.Equal(t, true, object.MatchesPath("a\\b\\c\\d"), "a\\b\\c\\d should match")
}

func TestWildCardFiles(t *testing.T) {
	gitIgnore := []string{"*.swp", "/foo/*.wat", "bar/*.txt"}
	object := CompileIgnoreLines(gitIgnore...)

	// Paths which are targeted by the above "lines"
	assert.Equal(t, true, object.MatchesPath("yo.swp"), "should ignore all swp files")
	assert.Equal(t, true, object.MatchesPath("something/else/but/it/hasyo.swp"), "should ignore all swp files in other directories")

	assert.Equal(t, true, object.MatchesPath("foo/bar.wat"), "should ignore all wat files in foo - nonpreceding /")
	assert.Equal(t, true, object.MatchesPath("/foo/something.wat"), "should ignore all wat files in foo - preceding /")

	assert.Equal(t, true, object.MatchesPath("bar/something.txt"), "should ignore all txt files in bar - nonpreceding /")
	assert.Equal(t, true, object.MatchesPath("/bar/somethingelse.txt"), "should ignore all txt files in bar - preceding /")

	// Paths which are not targeted by the above "lines"
	assert.Equal(t, false, object.MatchesPath("something/not/infoo/wat.wat"), "wat files should only be ignored in foo")
	assert.Equal(t, false, object.MatchesPath("something/not/infoo/wat.txt"), "txt files should only be ignored in bar")
}

func TestPrecedingSlash(t *testing.T) {
	gitIgnore := []string{"/foo", "bar/"}
	object := CompileIgnoreLines(gitIgnore...)

	assert.Equal(t, true, object.MatchesPath("foo/bar.wat"), "should ignore all files in foo - nonpreceding /")
	assert.Equal(t, true, object.MatchesPath("/foo/something.txt"), "should ignore all files in foo - preceding /")

	assert.Equal(t, true, object.MatchesPath("bar/something.txt"), "should ignore all files in bar - nonpreceding /")
	assert.Equal(t, true, object.MatchesPath("/bar/somethingelse.go"), "should ignore all files in bar - preceding /")
	assert.Equal(t, true, object.MatchesPath("/boo/something/bar/boo.txt"), "should block all files if bar is a sub directory")

	assert.Equal(t, false, object.MatchesPath("something/foo/something.txt"), "should only ignore top level foo directories- not nested")
}

func TestMatchesLineNumbers(t *testing.T) {
	gitIgnore := []string{"/foo", "bar/", "*.swp"}
	object := CompileIgnoreLines(gitIgnore...)

	var matchesPath bool
	var reason *IgnorePattern

	// /foo
	matchesPath, reason = object.MatchesPathHow("foo/bar.wat")
	assert.Equal(t, true, matchesPath, "should ignore all files in foo - nonpreceding /")
	assert.NotNil(t, reason, "reason should not be nil")
	assert.Equal(t, 1, reason.LineNo, "should match with line 1")
	assert.Equal(t, gitIgnore[0], reason.Line, "should match with line /foo")

	matchesPath, reason = object.MatchesPathHow("/foo/something.txt")
	assert.Equal(t, true, matchesPath, "should ignore all files in foo - preceding /")
	assert.NotNil(t, reason, "reason should not be nil")
	assert.Equal(t, 1, reason.LineNo, "should match with line 1")
	assert.Equal(t, gitIgnore[0], reason.Line, "should match with line /foo")

	// bar/
	matchesPath, reason = object.MatchesPathHow("bar/something.txt")
	assert.Equal(t, true, matchesPath, "should ignore all files in bar - nonpreceding /")
	assert.NotNil(t, reason, "reason should not be nil")
	assert.Equal(t, 2, reason.LineNo, "should match with line 2")
	assert.Equal(t, gitIgnore[1], reason.Line, "should match with line bar/")

	matchesPath, reason = object.MatchesPathHow("/bar/somethingelse.go")
	assert.Equal(t, true, matchesPath, "should ignore all files in bar - preceding /")
	assert.NotNil(t, reason, "reason should not be nil")
	assert.Equal(t, 2, reason.LineNo, "should match with line 2")
	assert.Equal(t, gitIgnore[1], reason.Line, "should match with line bar/")

	matchesPath, reason = object.MatchesPathHow("/boo/something/bar/boo.txt")
	assert.Equal(t, true, matchesPath, "should block all files if bar is a sub directory")
	assert.NotNil(t, reason, "reason should not be nil")
	assert.Equal(t, 2, reason.LineNo, "should match with line 2")
	assert.Equal(t, gitIgnore[1], reason.Line, "should match with line bar/")

	// *.swp
	matchesPath, reason = object.MatchesPathHow("yo.swp")
	assert.Equal(t, true, matchesPath, "should ignore all swp files")
	assert.NotNil(t, reason, "reason should not be nil")
	assert.Equal(t, 3, reason.LineNo, "should match with line 3")
	assert.Equal(t, gitIgnore[2], reason.Line, "should match with line *.swp")

	matchesPath, reason = object.MatchesPathHow("something/else/but/it/hasyo.swp")
	assert.Equal(t, true, matchesPath, "should ignore all swp files in other directories")
	assert.NotNil(t, reason, "reason should not be nil")
	assert.Equal(t, 3, reason.LineNo, "should match with line 3")
	assert.Equal(t, gitIgnore[2], reason.Line, "should match with line *.swp")

	// other
	matchesPath, reason = object.MatchesPathHow("something/foo/something.txt")
	assert.Equal(t, false, matchesPath, "should only ignore top level foo directories- not nested")
	assert.Nil(t, reason, "reason should be nil as no match should happen")
}
