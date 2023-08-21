package error_templates

import (
	"github.com/nedpals/errgoengine/error_templates/java"
	"github.com/nedpals/errgoengine/error_templates/python"
	"github.com/nedpals/errgoengine/lib"
)

func LoadErrorTemplates(errorTemplates *lib.ErrorTemplates) {
	java.LoadErrorTemplates(errorTemplates)
	python.LoadErrorTemplates(errorTemplates)
}
