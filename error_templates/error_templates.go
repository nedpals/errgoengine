package error_templates

import (
	lib "github.com/nedpals/errgoengine"
	"github.com/nedpals/errgoengine/error_templates/java"
	"github.com/nedpals/errgoengine/error_templates/python"
)

func LoadErrorTemplates(errorTemplates *lib.ErrorTemplates) {
	java.LoadErrorTemplates(errorTemplates)
	python.LoadErrorTemplates(errorTemplates)
}
