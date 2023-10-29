package errgoengine

import (
	"fmt"
	"regexp"
	"strings"
)

type GenAnalyzeErrorFn func(cd *ContextData, m *MainError)

type ErrorTemplate struct {
	Name              string
	Pattern           string
	StackTracePattern string
	OnAnalyzeErrorFn  GenAnalyzeErrorFn
	OnGenExplainFn    GenExplainFn
	OnGenBugFixFn     GenBugFixFn
}

func CustomErrorPattern(pattern string) string {
	return "\"\"\"" + pattern
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

func (tmps *ErrorTemplates) Add(language *Language, template ErrorTemplate) (*CompiledErrorTemplate, error) {
	key := TemplateKey(language.Name, template.Name)
	if template.OnGenExplainFn == nil {
		return nil, fmt.Errorf("(%s) OnGenExplainFn is required", key)
	} else if template.OnGenBugFixFn == nil {
		return nil, fmt.Errorf("(%s) OnGenBugFixFn is required", key)
	}

	if key == "." {
		return nil, fmt.Errorf("language name and/or template name are empty")
	} else if tmp, templateExists := (*tmps)[key]; templateExists {
		return tmp, nil
	} else if !language.isCompiled {
		language.Compile()
	}

	var stackTracePattern *regexp.Regexp

	// TODO: add test
	patternForCompile := template.Pattern
	if strings.HasPrefix(patternForCompile, "\"\"\"") {
		// If it is a custom template error pattern which starts at triple quotes (""")
		patternForCompile = strings.TrimPrefix(patternForCompile, "\"\"\"")
	} else {
		// If not revert to language level error pattern if present
		if len(language.ErrorPattern) != 0 {
			patternForCompile = strings.ReplaceAll(language.ErrorPattern, "$message", template.Pattern)
		} else if !strings.Contains(patternForCompile, "$stacktrace") {
			patternForCompile = patternForCompile + "$stacktrace"
		}
	}

	if strings.Contains(patternForCompile, "$stacktrace") {
		patternForCompile =
			strings.ReplaceAll(patternForCompile, "$stacktrace", defaultStackTraceRegex)

		if len(template.StackTracePattern) != 0 {
			var err error
			stackTracePattern, err = regexp.Compile(template.StackTracePattern)
			if err != nil {
				return nil, err
			}
		}
	}

	compiledPattern, err := regexp.Compile("(?m)^" + patternForCompile + "$")
	if err != nil {
		return nil, err
	}

	(*tmps)[key] = &CompiledErrorTemplate{
		ErrorTemplate:     template,
		Language:          language,
		Pattern:           compiledPattern,
		StackTracePattern: stackTracePattern,
	}
	return (*tmps)[key], nil
}

func (tmps ErrorTemplates) MustAdd(language *Language, template ErrorTemplate) *CompiledErrorTemplate {
	tmp, err := tmps.Add(language, template)
	if err != nil {
		panic(fmt.Sprintf("ErrorTemplates.MustAdd: %s", err.Error()))
	}
	return tmp
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
