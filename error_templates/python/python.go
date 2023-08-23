package python

import (
	lib "github.com/nedpals/errgoengine"
	"github.com/nedpals/errgoengine/languages/python"
)

func LoadErrorTemplates(errorTemplates *lib.ErrorTemplates) {
	errorTemplates.Add(python.Language, ZeroDivisionError)
	errorTemplates.Add(python.Language, NameError)
}
