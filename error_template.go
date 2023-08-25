package errgoengine

import (
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

type ErrorTemplates []*CompiledErrorTemplate

const defaultStackTraceRegex = `(?P<stacktrace>(?:.|\s)*)`

func (tmps *ErrorTemplates) Add(language *Language, template ErrorTemplate) *CompiledErrorTemplate {
	if !language.isCompiled {
		language.Compile()
	}

	var stackTracePattern *regexp.Regexp

	patternForCompile := ""
	if len(language.ErrorPattern) != 0 {
		patternForCompile = strings.ReplaceAll(language.ErrorPattern, "$message", template.Pattern)
	} else if len(language.StackTracePattern) != 0 {
		patternForCompile = template.Pattern + "$stacktrace"
	}

	if strings.Contains(patternForCompile, "$stacktrace") {
		patternForCompile =
			strings.ReplaceAll(patternForCompile, "$stacktrace", defaultStackTraceRegex)

		if len(template.StackTracePattern) != 0 {
			stackTracePattern = regexp.MustCompile(template.StackTracePattern)
		}
	}

	compiledPattern := regexp.MustCompile("(?m)^" + patternForCompile + "$")

	*tmps = append(*tmps, &CompiledErrorTemplate{
		ErrorTemplate:     template,
		Language:          language,
		Pattern:           compiledPattern,
		StackTracePattern: stackTracePattern,
	})

	return (*tmps)[len(*tmps)-1]
}

func (tmps ErrorTemplates) Find(msg string) *CompiledErrorTemplate {
	for _, tmp := range tmps {
		if tmp.Pattern.MatchString(msg) {
			return tmp
		}
	}
	return nil
}
