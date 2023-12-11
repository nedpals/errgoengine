package errgoengine

import (
	"bytes"
	"fmt"
	"io/fs"
	"path/filepath"

	sitter "github.com/smacker/go-tree-sitter"
)

type ErrgoEngine struct {
	SharedStore    *Store
	ErrorTemplates ErrorTemplates
	FS             *MultiReadFileFS
	OutputGen      *OutputGenerator
}

func New() *ErrgoEngine {
	filesystems := make([]fs.ReadFileFS, 2)
	filesystems[0] = &RawFS{}

	return &ErrgoEngine{
		SharedStore:    NewEmptyStore(),
		ErrorTemplates: ErrorTemplates{},
		FS: &MultiReadFileFS{
			FSs: filesystems,
		},
		OutputGen: &OutputGenerator{},
	}
}

func (e *ErrgoEngine) Analyze(workingPath, msg string) (*CompiledErrorTemplate, *ContextData, error) {
	template := e.ErrorTemplates.Match(msg)
	if template == nil {
		return nil, nil, fmt.Errorf("template not found. \nMessage: %s", msg)
	}

	// initial context data extraction
	contextData := NewContextData(e.SharedStore, workingPath)
	contextData.Analyzer = template.Language.AnalyzerFactory(contextData)

	if template.Language.stubFs != nil {
		e.FS.FSs[1] = template.Language.stubFs
	}

	groupNames := template.Pattern.SubexpNames()
	for _, submatches := range template.Pattern.FindAllStringSubmatch(msg, -1) {
		for idx, matchedContent := range submatches {
			if len(groupNames[idx]) == 0 {
				continue
			}

			contextData.AddVariable(groupNames[idx], matchedContent)
		}
	}

	// extract stack trace
	rawStackTraces := contextData.Variables["stacktrace"]
	symbolGroupIdx := template.StackTraceRegex().SubexpIndex("symbol")
	pathGroupIdx := template.StackTraceRegex().SubexpIndex("path")
	posGroupIdx := template.StackTraceRegex().SubexpIndex("position")
	stackTraceMatches := template.StackTraceRegex().FindAllStringSubmatch(rawStackTraces, -1)

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

		stLoc := template.Language.LocationConverter(rawPath, rawPos)
		if contextData.TraceStack == nil {
			contextData.TraceStack = TraceStack{}
		}

		contextData.TraceStack.Add(rawSymbolName, stLoc)
	}

	// open contents of the extracted stack file locations
	parser := sitter.NewParser()
	analyzer := &SymbolAnalyzer{ContextData: contextData}

	for _, node := range contextData.TraceStack {
		contents, err := e.FS.ReadFile(node.DocumentPath)
		if err != nil {
			return nil, nil, err
		}

		// Skip stub files
		if len(contents) == 0 {
			continue
		}

		var selectedLanguage *Language
		existingDoc, docExists := contextData.Documents[node.DocumentPath]

		// check matched languages
		if docExists {
			selectedLanguage = existingDoc.Language
		} else {
			selectedLanguage = template.Language
			if !selectedLanguage.MatchPath(node.DocumentPath) {
				return nil, nil, fmt.Errorf("no language found for %s", node.DocumentPath)
			}

			// compile language first (if not yet)
			selectedLanguage.Compile()
		}

		// do semantic analysis
		doc, err := ParseDocument(node.DocumentPath, bytes.NewReader(contents), parser, selectedLanguage, existingDoc)
		if err != nil {
			return nil, nil, err
		}

		// add doc if it does not already exist
		if doc != existingDoc {
			doc = contextData.AddDocument(doc)
		}

		analyzer.Analyze(doc)
	}

	// locate main error
	mainTraceNode := contextData.TraceStack.NearestTo(contextData.WorkingPath)

	// get nearest node
	doc := contextData.Documents[mainTraceNode.DocumentPath]
	nearest := doc.Tree.RootNode().NamedDescendantForPointRange(
		sitter.Point{Row: uint32(mainTraceNode.StartPos.Line)},
		sitter.Point{Row: uint32(mainTraceNode.EndPos.Line)},
	)

	if nearest.StartPoint().Row != uint32(mainTraceNode.StartPos.Line) {
		cursor := sitter.NewTreeCursor(nearest)
		nearest = nearestNodeFromPos(cursor, mainTraceNode.StartPos)
	}

	// further analyze main error
	contextData.MainError = &MainError{
		ErrorNode: &mainTraceNode,
		Document:  doc,
		Nearest:   WrapNode(doc, nearest),
	}

	if contextData.MainError != nil && template.OnAnalyzeErrorFn != nil {
		template.OnAnalyzeErrorFn(contextData, contextData.MainError)
	}

	return template, contextData, nil
}

func (e *ErrgoEngine) Translate(template *CompiledErrorTemplate, contextData *ContextData) string {
	expGen := &ExplainGenerator{errorName: template.Name}
	fixGen := &BugFixGenerator{}

	// execute error generator function
	template.OnGenExplainFn(contextData, expGen)
	template.OnGenBugFixFn(contextData, fixGen)

	output := e.OutputGen.Generate(contextData, expGen, fixGen)
	defer e.OutputGen.Reset()
	return output
}
