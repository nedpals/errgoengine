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
	return &ErrgoEngine{
		SharedStore:    NewEmptyStore(),
		ErrorTemplates: ErrorTemplates{},
		FS: &MultiReadFileFS{
			FSs: []fs.ReadFileFS{
				&RawFS{},
			},
		},
		OutputGen: &OutputGenerator{},
	}
}

func (e *ErrgoEngine) AttachMainFS(instance fs.ReadFileFS) {
	// remove existing documents
	fs.WalkDir(e.FS.FSs[0], ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		} else if _, ok := e.SharedStore.Documents[path]; !ok {
			return nil
		}

		// delete document
		delete(e.SharedStore.Documents, path)
		return nil
	})

	// attach new fs
	e.FS.Attach(instance, 0)
}

func (e *ErrgoEngine) Analyze(workingPath, msg string) (*CompiledErrorTemplate, *ContextData, error) {
	template := e.ErrorTemplates.Match(msg)
	if template == nil {
		return nil, nil, fmt.Errorf("template not found. \nMessage: %s", msg)
	}

	// initial context data extraction
	contextData := NewContextData(e.SharedStore, workingPath)
	contextData.Analyzer = template.Language.AnalyzerFactory(contextData)
	contextData.AddVariable("message", msg)

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
	rawStackTraceItem := contextData.Variables["stacktrace"]
	symbolGroupIdx := template.StackTraceRegex().SubexpIndex("symbol")
	pathGroupIdx := template.StackTraceRegex().SubexpIndex("path")
	posGroupIdx := template.StackTraceRegex().SubexpIndex("position")
	stackTraceMatches := template.StackTraceRegex().FindAllStringSubmatch(rawStackTraceItem, -1)

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

		stLoc := template.Language.LocationConverter(LocationConverterContext{
			Path:        rawPath,
			Pos:         rawPos,
			ContextData: contextData,
		})
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
			// return nil, nil, err
			// Do not return error if file not found
			continue
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
	if doc, ok := contextData.Documents[mainTraceNode.DocumentPath]; ok {
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
	} else {
		return nil, nil, fmt.Errorf("main trace node document not found")
	}

	if contextData.MainError != nil && template.OnAnalyzeErrorFn != nil {
		template.OnAnalyzeErrorFn(contextData, contextData.MainError)
	}

	return template, contextData, nil
}

func (e *ErrgoEngine) Translate(template *CompiledErrorTemplate, contextData *ContextData) (mainExp string, fullExp string) {
	expGen := &ExplainGenerator{errorName: template.Name}
	fixGen := &BugFixGenerator{
		Document: contextData.MainError.Document,
	}

	// execute error generator function
	template.OnGenExplainFn(contextData, expGen)
	template.OnGenBugFixFn(contextData, fixGen)

	output := e.OutputGen.Generate(contextData, expGen, fixGen)
	defer e.OutputGen.Reset()

	return expGen.mainExp.String(), output
}
