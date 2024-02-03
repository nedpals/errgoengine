package errgoengine

import (
	"bytes"
	"fmt"
	"io/fs"

	sitter "github.com/smacker/go-tree-sitter"
)

type ErrgoEngine struct {
	SharedStore    *Store
	ErrorTemplates ErrorTemplates
	FS             *MultiReadFileFS
	OutputGen      *OutputGenerator
	IsTesting      bool
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

	// extract variables from the error message
	contextData.AddVariables(template.ExtractVariables(msg))

	// extract stack trace
	contextData.TraceStack = template.ExtractStackTrace(contextData)

	// open contents of the extracted stack file locations
	if err := ParseFromStackTrace(contextData, template.Language, e.FS); err != nil {
		// return error template for bugbuddy to handle
		// incomplete error messages
		return template, nil, err
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
	} else if template == FallbackErrorTemplate {
		contextData.MainError = &MainError{
			ErrorNode: nil,
			Document:  nil,
			Nearest:   SyntaxNode{},
		}
	} else {
		// return error template for bugbuddy to handle
		// incomplete error messages
		return template, nil, fmt.Errorf("main trace node document not found")
	}

	if contextData.MainError != nil && template.OnAnalyzeErrorFn != nil {
		template.OnAnalyzeErrorFn(contextData, contextData.MainError)
	}

	return template, contextData, nil
}

func (e *ErrgoEngine) Translate(template *CompiledErrorTemplate, contextData *ContextData) (mainExp string, fullExp string) {
	expGen := &ExplainGenerator{ErrorName: template.Name}
	fixGen := &BugFixGenerator{}
	if contextData.MainError != nil {
		fixGen.Document = contextData.MainError.Document
	}

	// execute error generator function
	template.OnGenExplainFn(contextData, expGen)

	// execute bug fix generator function
	if template.OnGenBugFixFn != nil {
		template.OnGenBugFixFn(contextData, fixGen)
	}

	if e.IsTesting {
		// add a code snippet that points to the error
		e.OutputGen.GenAfterExplain = func(gen *OutputGenerator) {
			err := contextData.MainError
			if err == nil {
				return
			}

			doc := err.Document
			if doc == nil || err.Nearest.IsNull() {
				return
			}

			startLineNr := err.Nearest.StartPosition().Line
			startLines := doc.LinesAt(max(startLineNr-1, 0), startLineNr)
			endLines := doc.LinesAt(min(startLineNr+1, doc.TotalLines()), min(startLineNr+2, doc.TotalLines()))
			arrowLength := int(err.Nearest.EndByte() - err.Nearest.StartByte())
			if arrowLength == 0 {
				arrowLength = 1
			}

			startArrowPos := err.Nearest.StartPosition().Column
			gen.Writeln("```")
			gen.WriteLines(startLines...)

			for i := 0; i < startArrowPos; i++ {
				if startLines[len(startLines)-1][i] == '\t' {
					gen.Builder.WriteString("    ")
				} else {
					gen.Builder.WriteByte(' ')
				}
			}

			for i := 0; i < arrowLength; i++ {
				gen.Builder.WriteByte('^')
			}

			gen.Break()
			gen.WriteLines(endLines...)
			gen.Writeln("```")
		}
	}

	output := e.OutputGen.Generate(expGen, fixGen)
	defer e.OutputGen.Reset()

	return expGen.Builder.String(), output
}

func ParseFromStackTrace(contextData *ContextData, defaultLanguage *Language, files fs.ReadFileFS) error {
	parser := sitter.NewParser()
	analyzer := &SymbolAnalyzer{ContextData: contextData}

	for _, node := range contextData.TraceStack {
		path := node.DocumentPath

		contents, err := files.ReadFile(path)
		if err != nil {
			// return err
			// Do not return error if file not found
			continue
		}

		// Skip stub files
		if len(contents) == 0 {
			continue
		}

		// check if document already exists
		existingDoc, docExists := contextData.Documents[path]

		// check matched languages
		selectedLanguage := defaultLanguage
		if docExists {
			selectedLanguage = existingDoc.Language
		} else {
			if !selectedLanguage.MatchPath(path) {
				return fmt.Errorf("no language found for %s", path)
			}

			// compile language first (if not yet)
			selectedLanguage.Compile()
		}

		// do semantic analysis
		contentReader := bytes.NewReader(contents)
		doc, err := ParseDocument(path, contentReader, parser, selectedLanguage, existingDoc)
		if err != nil {
			return err
		}

		// add doc if it does not already exist
		if doc != existingDoc {
			doc = contextData.AddDocument(doc)
		}

		analyzer.Analyze(doc)
	}

	return nil
}
