/**
 * JavaScript implementation of GitIgnore pattern matching
 */

/**
 * Parse a pattern line from a .gitignore file
 * @param {string} line - A line from a .gitignore file
 * @returns {Object|null} - Pattern object or null if line is a comment or empty
 */
function getPatternFromLine(line) {
  // Trim OS-specific carriage returns
  line = line.replace(/\r$/, "");

  // Strip comments [Rule 2]
  if (line.startsWith("#")) {
    return null;
  }

  // Trim string [Rule 3]
  // TODO: Handle [Rule 3], when the " " is escaped with a \
  line = line.trim();

  // Exit for no-ops
  if (line === "") {
    return null;
  }

  // Handle negation patterns [Rule 4]
  let negatePattern = false;
  if (line[0] === "!") {
    negatePattern = true;
    line = line.substring(1);
  }

  // Handle [Rule 2, 4], when # or ! is escaped with a \
  if (/^(\#|\!)/.test(line)) {
    line = line.substring(1);
  }

  // If we encounter a foo/*.blah in a folder, prepend the / char
  if (/([^\/+])\/.*\*\./.test(line) && line[0] !== "/") {
    line = "/" + line;
  }

  // Handle escaping the "." char
  line = line.replace(/\./g, "\\.");

  const magicStar = "#$~";

  // Handle "/**/" usage
  if (line.startsWith("/**/")) {
    line = line.substring(1);
  }
  line = line.replace(/\/\*\*\//g, "(/|/.+/)");
  line = line.replace(/\*\*\//g, "(|." + magicStar + "/)");
  line = line.replace(/\/\*\*/g, "(|/." + magicStar + ")");

  // Handle escaping the "*" char
  line = line.replace(/\\\*/g, "\\" + magicStar);
  line = line.replace(/\*/g, "([^/]*)");

  // Handle escaping the "?" char
  line = line.replace(/\?/g, "\\?");

  // Replace magic star back
  line = line.replace(new RegExp(magicStar, "g"), "*");

  // Build final regex
  let expr = "";
  if (line.endsWith("/")) {
    expr = line + "(|.*)$";
  } else {
    expr = line + "(|/.*)$";
  }

  if (expr.startsWith("/")) {
    expr = "^(|/)" + expr.substring(1);
  } else {
    expr = "^(|.*/)" + expr;
  }

  const pattern = new RegExp(expr);

  return { pattern, negatePattern };
}

/**
 * Class representing a GitIgnore pattern
 */
class IgnorePattern {
  /**
   * Create an IgnorePattern
   * @param {RegExp} pattern - Regular expression pattern
   * @param {boolean} negate - Whether this is a negated pattern
   * @param {number} lineNo - Line number in the gitignore file
   * @param {string} line - Original line from gitignore file
   */
  constructor(pattern, negate, lineNo, line) {
    this.pattern = pattern;
    this.negate = negate;
    this.lineNo = lineNo;
    this.line = line;
  }
}

/**
 * Class representing a GitIgnore file
 */
class GitIgnore {
  /**
   * Create a GitIgnore instance
   */
  constructor() {
    this.patterns = [];
  }

  /**
   * Check if a path matches any of the patterns
   * @param {string} filePath - Path to check
   * @returns {boolean} - Whether the path matches any patterns
   */
  matchesPath(filePath) {
    const [matchesPath] = this.matchesPathHow(filePath);
    return matchesPath;
  }

  /**
   * Check if a path matches any of the patterns and return how it matched
   * @param {string} filePath - Path to check
   * @returns {Array} - [boolean, IgnorePattern|null] Whether the path matches and which pattern matched
   */
  matchesPathHow(filePath) {
    // Replace OS-specific path separator
    filePath = filePath.replace(/\\/g, "/");

    let matchesPath = false;
    let matchedPattern = null;

    for (const pattern of this.patterns) {
      if (pattern.pattern.test(filePath)) {
        // If this is a regular target (not negated)
        if (!pattern.negate) {
          matchesPath = true;
          matchedPattern = pattern;
        } else if (matchesPath) {
          // Negated pattern, and matchesPath is already set
          matchesPath = false;
          // Note: In the original Go code, matchedPattern is not reset here
        }
      }
    }

    return [matchesPath, matchedPattern];
  }
}

/**
 * Compile ignore lines into a GitIgnore instance
 * @param {string[]} lines - Lines from a gitignore file
 * @returns {GitIgnore} - GitIgnore instance
 */
function compileIgnoreLines(...lines) {
  const gitIgnore = new GitIgnore();

  lines.forEach((line, index) => {
    const result = getPatternFromLine(line);
    if (result) {
      // LineNo is 1-based numbering to match `git check-ignore -v` output
      const { pattern, negatePattern } = result;
      const ignorePattern = new IgnorePattern(
        pattern,
        negatePattern,
        index + 1,
        line
      );
      gitIgnore.patterns.push(ignorePattern);
    }
  });

  return gitIgnore;
}

// Export the classes and functions
module.exports = {
  GitIgnore,
  IgnorePattern,
  compileIgnoreLines,
  getPatternFromLine,
};
