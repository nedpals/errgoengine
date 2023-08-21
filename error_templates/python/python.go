package python

import (
	"github.com/nedpals/errgoengine/languages/python"
	"github.com/nedpals/errgoengine/lib"
)

func LoadErrorTemplates(errorTemplates *lib.ErrorTemplates) {
	errorTemplates.Add(python.Language, ZeroDivisionError)
	errorTemplates.Add(python.Language, NameError)
}
