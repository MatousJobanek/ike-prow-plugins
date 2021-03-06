package testkeeper

import (
	"regexp"
	"strings"
)

const (
	regexpDefinitionPrefix = "regex{{"
	regexpDefinitionSuffix = "}}"
	anyPathWildcard        = "**"
	anyNameWildcard        = "*"
	anyNameRegexp          = "[^/]*"
	anythingRegexp         = ".*"
	twoStarsReplacement    = "<two-stars-replacement>"
	endOfLineRegexp        = "$"
	directorySeparator     = "/"
)

// FilePattern contains regexp that matches a file
type FilePattern struct {
	Regexp string
}

// Matches checks if the given string (representing path to a file) contains a substring that matches Regexp stored in this matcher
func (matcher *FilePattern) Matches(filename string) bool {
	return regexp.MustCompile(matcher.Regexp).MatchString(filename)
}

// FilePatterns is an alias type representing slice of FilePattern
type FilePatterns []FilePattern

// Matches iterates over all patterns and returns first successful match or false if none patterns matched
func (f *FilePatterns) Matches(filename string) bool {

	for _, matcher := range *f {
		if matcher.Matches(filename) {
			return true
		}
	}

	return false
}

// ParseFilePatterns takes the given patterns and parses to an array of FilePattern instances
func ParseFilePatterns(filePatterns []string) FilePatterns {
	patterns := make([]FilePattern, 0, len(filePatterns))
	for _, pattern := range filePatterns {
		patterns = append(patterns, FilePattern{
			Regexp: parseFilePattern(strings.TrimSpace(pattern)),
		})
	}
	return patterns
}

func parseFilePattern(pattern string) string {

	// if it is regex{{...}} then just return the content
	if strings.HasPrefix(pattern, regexpDefinitionPrefix) && strings.HasSuffix(pattern, regexpDefinitionSuffix) {
		return pattern[len(regexpDefinitionPrefix) : len(pattern)-len(regexpDefinitionSuffix)]
	}

	// if not, then transform the pattern to regexp
	slashIndex := strings.LastIndexAny(pattern, directorySeparator)

	path := transformPathPatternToRegexp(pattern[:slashIndex+1])
	fileName := transformFilenamePatternToRegexp(pattern[slashIndex+1:], path)

	expr := path + fileName

	if strings.HasSuffix(expr, directorySeparator) {
		expr += anythingRegexp
	} else {
		expr += endOfLineRegexp
	}

	return expr
}

func transformPathPatternToRegexp(path string) string {
	for strings.HasPrefix(path, anyPathWildcard+"/") {
		path = path[strings.Index(path, "/")+1:]
	}
	path = escapeDots(path)
	path = strings.Replace(path, anyPathWildcard, twoStarsReplacement, -1)
	path = replaceAnyNameWildcards(path)
	return strings.Replace(path, twoStarsReplacement, anythingRegexp, -1)
}

func transformFilenamePatternToRegexp(fileName string, path string) string {
	fileName = escapeDots(fileName)

	if strings.HasPrefix(fileName, anyNameWildcard) {
		newPrefix := anyNameRegexp
		if path == "" || len(fileName) == 1 {
			newPrefix = anythingRegexp
		}
		return newPrefix + replaceAnyNameWildcards(fileName[1:])
	}
	return replaceAnyNameWildcards(fileName)

}

func escapeDots(s string) string {
	return strings.Replace(s, ".", "\\.", -1)
}

func replaceAnyNameWildcards(s string) string {
	return strings.Replace(s, anyNameWildcard, anyNameRegexp, -1)
}
