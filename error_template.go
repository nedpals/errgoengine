package errgoengine

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

type ErrorTemplate struct {
	Name              string
	Pattern           string
	StackTracePattern string
	OnAnalyzeErrorFn  func(cd *ContextData, m *MainError)
	OnGenExplainFn    func(cd *ContextData, gen *ExplainGenerator)
	OnGenBugFixFn     func(cd *ContextData, gen *BugFixGenerator)
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
	if tmp.Pattern == nil {
		return map[string]string{}
	}
	variables := map[string]string{}
	groupNames := tmp.Pattern.SubexpNames()
	for _, submatches := range tmp.Pattern.FindAllStringSubmatch(msg, -1) {
		for idx, matchedContent := range submatches {
			if len(groupNames[idx]) == 0 {
				continue
			}

			if v, ok := variables[groupNames[idx]]; ok && len(v) != 0 {
				continue
			}

			variables[groupNames[idx]] = matchedContent
		}
	}
	return variables
}

func (tmp *CompiledErrorTemplate) ExtractStackTrace(cd *ContextData) TraceStack {
	traceStack := TraceStack{}
	stackTraceRegex := tmp.StackTraceRegex()
	if stackTraceRegex == nil {
		return traceStack
	}

	workingPath := cd.WorkingPath
	rawStackTraceItem := cd.Variables["stacktrace"]
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
	if tmp == FallbackErrorTemplate {
		return true
	}
	return tmp.Pattern.MatchString(str)
}

type ErrorTemplates map[string]*CompiledErrorTemplate

const defaultStackTraceRegexOpening = `(?P<stacktrace>(?:`
const defaultStackTraceRegexClosing = `)*)`
const defaultStackTraceRegex = defaultStackTraceRegexOpening + `.|\s` + defaultStackTraceRegexClosing

var stackTraceCaptureGroupRegex = regexp.MustCompile(`\(\?P<(?:symbol|path|position)>([a-z0-9A-Z_\\//+.]+)\)`)

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
		rawStackTracePattern := template.StackTracePattern
		if len(rawStackTracePattern) == 0 && len(language.StackTracePattern) != 0 {
			rawStackTracePattern = language.StackTracePattern
		}

		if len(template.StackTracePattern) != 0 {
			var err error
			stackTracePattern, err = regexp.Compile(template.StackTracePattern)
			if err != nil {
				return nil, err
			}
		}

		if len(rawStackTracePattern) != 0 {
			// strip first the group names before replacing $stacktrace
			strippedStackTracePattern := stackTraceCaptureGroupRegex.ReplaceAllString(rawStackTracePattern, "$1")

			// replace $stacktrace with the actual stack trace pattern
			patternForCompile =
				strings.ReplaceAll(patternForCompile, "$stacktrace",
					defaultStackTraceRegexOpening+strippedStackTracePattern+defaultStackTraceRegexClosing)
		} else {
			patternForCompile =
				strings.ReplaceAll(patternForCompile, "$stacktrace", defaultStackTraceRegex)
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
	if len(language) == 0 {
		return name
	}
	return fmt.Sprintf("%s.%s", language, name)
}

var FallbackErrorTemplate = &CompiledErrorTemplate{
	ErrorTemplate: ErrorTemplate{
		Name:    "UnknownError",
		Pattern: `.*`,
		OnGenExplainFn: func(cd *ContextData, gen *ExplainGenerator) {
			gen.Add("There are no available error templates for this error.\n")
			gen.Add("```\n")
			gen.Add(cd.Variables["message"])
			gen.Add("\n```")
		},
	},
	Language: &Language{
		AnalyzerFactory: func(cd *ContextData) LanguageAnalyzer {
			return nil
		},
	},
	Pattern:           nil,
	StackTracePattern: nil,
}
