package errgoengine

import (
	"fmt"
	"path/filepath"
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

func (tmp *CompiledErrorTemplate) ExtractVariables(msg string) map[string]string {
	variables := map[string]string{}
	groupNames := tmp.Pattern.SubexpNames()
	for _, submatches := range tmp.Pattern.FindAllStringSubmatch(msg, -1) {
		for idx, matchedContent := range submatches {
			if len(groupNames[idx]) == 0 {
				continue
			}

			variables[groupNames[idx]] = matchedContent
		}
	}
	return variables
}

func (tmp *CompiledErrorTemplate) ExtractStackTrace(cd *ContextData) TraceStack {
	traceStack := TraceStack{}
	workingPath := cd.WorkingPath

	rawStackTraceItem := cd.Variables["stacktrace"]
	stackTraceRegex := tmp.StackTraceRegex()
	symbolGroupIdx := stackTraceRegex.SubexpIndex("symbol")
	pathGroupIdx := stackTraceRegex.SubexpIndex("path")
	posGroupIdx := stackTraceRegex.SubexpIndex("position")
	stackTraceMatches := stackTraceRegex.FindAllStringSubmatch(rawStackTraceItem, -1)

	for _, submatches := range stackTraceMatches {
		if len(submatches) == 0 {
			continue
		}

		rawSymbolName := ""
		if symbolGroupIdx != -1 {
			rawSymbolName = submatches[symbolGroupIdx]
		}
		rawPath := submatches[pathGroupIdx]
		rawPos := submatches[posGroupIdx]

		// convert relative paths to absolute for parsing
		if len(workingPath) != 0 && !filepath.IsAbs(rawPath) {
			rawPath = filepath.Clean(filepath.Join(workingPath, rawPath))
		}

		stLoc := tmp.Language.LocationConverter(LocationConverterContext{
			Path:        rawPath,
			Pos:         rawPos,
			ContextData: cd,
		})

		traceStack.Add(rawSymbolName, stLoc)
	}

	return traceStack
}

func (tmp *CompiledErrorTemplate) Match(str string) bool {
	return tmp.Pattern.MatchString(str)
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
		if tmp.Match(msg) {
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
