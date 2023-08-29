package errgoengine

import (
	"fmt"
	"regexp"
	"strings"
)

type ErrorTemplate struct {
	Name              string
	Pattern           string
	StackTracePattern string
	OnGenExplainFn    GenExplainFn
	OnGenBugFixFn     GenBugFixFn
}

type CompiledErrorTemplate struct {
	ErrorTemplate
	Language          *Language
	Pattern           *regexp.Regexp
	StackTracePattern *regexp.Regexp
}

func (tmp *CompiledErrorTemplate) StackTraceRegex() *regexp.Regexp {
	if tmp.StackTracePattern != nil {
		return tmp.StackTracePattern
	} else if len(tmp.Language.StackTracePattern) != 0 && tmp.Language.stackTraceRegex == nil {
		panic("expected stacktrace pattern got compiled, got nil regex instead")
	}
	return tmp.Language.stackTraceRegex
}

type ErrorTemplates map[string]*CompiledErrorTemplate

const defaultStackTraceRegex = `(?P<stacktrace>(?:.|\s)*)`

func (tmps *ErrorTemplates) Add(language *Language, template ErrorTemplate) *CompiledErrorTemplate {
	key := TemplateKey(language.Name, template.Name)
	if key == "." {
		panic("Invalid template registration.")
	} else if tmp, templateExists := (*tmps)[key]; templateExists {
		return tmp
	} else if !language.isCompiled {
		language.Compile()
	}

	var stackTracePattern *regexp.Regexp

	// TODO: add test
	patternForCompile := template.Pattern
	if len(language.ErrorPattern) != 0 {
		patternForCompile = strings.ReplaceAll(language.ErrorPattern, "$message", template.Pattern)
	} else if !strings.Contains(patternForCompile, "$stacktrace") {
		patternForCompile = patternForCompile + "$stacktrace"
	}

	if strings.Contains(patternForCompile, "$stacktrace") {
		patternForCompile =
			strings.ReplaceAll(patternForCompile, "$stacktrace", defaultStackTraceRegex)

		if len(template.StackTracePattern) != 0 {
			stackTracePattern = regexp.MustCompile(template.StackTracePattern)
		}
	}

	compiledPattern := regexp.MustCompile("(?m)^" + patternForCompile + "$")
	(*tmps)[key] = &CompiledErrorTemplate{
		ErrorTemplate:     template,
		Language:          language,
		Pattern:           compiledPattern,
		StackTracePattern: stackTracePattern,
	}
	return (*tmps)[key]
}

func (tmps ErrorTemplates) Match(msg string) *CompiledErrorTemplate {
	for _, tmp := range tmps {
		if tmp.Pattern.MatchString(msg) {
			return tmp
		}
	}
	return nil
}

func (tmps ErrorTemplates) Find(language, name string) *CompiledErrorTemplate {
	key := TemplateKey(language, name)
	tmp, exists := tmps[key]
	if !exists {
		return nil
	}
	return tmp
}

func TemplateKey(language, name string) string {
	return fmt.Sprintf("%s.%s", language, name)
}
