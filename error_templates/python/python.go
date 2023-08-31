package python

import (
	lib "github.com/nedpals/errgoengine"
	"github.com/nedpals/errgoengine/languages/python"
)

func LoadErrorTemplates(errorTemplates *lib.ErrorTemplates) {
	// Runtime error
	errorTemplates.MustAdd(python.Language, ZeroDivisionError)
	errorTemplates.MustAdd(python.Language, NameError)
	errorTemplates.MustAdd(python.Language, ValueError)
	errorTemplates.MustAdd(python.Language, AttributeError)

	// Compile time error
	errorTemplates.MustAdd(python.Language, SyntaxError)
	errorTemplates.MustAdd(python.Language, IndentationError)
}

func compileTimeError(pattern string) string {
	return lib.CustomErrorPattern("$stacktrace" + pattern)
}
