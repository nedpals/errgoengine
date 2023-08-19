package java

import (
	"fmt"
	"strings"

	"github.com/nedpals/errgoengine/languages/java"
	"github.com/nedpals/errgoengine/lib"
)

func LoadErrorTemplates(errorTemplates *lib.ErrorTemplates) {
	errorTemplates.Add(java.Language, NullPointerException)
	errorTemplates.Add(java.Language, ArrayIndexOutOfBoundsException)
	errorTemplates.Add(java.Language, ArithmeticException)
	errorTemplates.Add(java.Language, PublicClassFilenameMismatchError)
}

func runtimeErrorPattern(errorName string, pattern string) string {
	p := fmt.Sprintf(
		`Exception in thread "(?P<thread>\w+)" %s`,
		strings.ReplaceAll(errorName, ".", `\.`),
	)
	if len(pattern) != 0 {
		p += ": " + pattern
	}
	return p
}

const comptimeStackTracePattern = `(?P<path>\S+):(?P<position>\d+)`

func comptimeErrorPattern(pattern string) string {
	return fmt.Sprintf(
		`(?P<stacktrace>(?:.|\s)*) error: %s.*`,
		pattern,
	)
}
