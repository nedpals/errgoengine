package errgoengine

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

type ErrgoEngine struct {
	WorkingPath    string
	ErrorTemplates ErrorTemplates
	FS             fs.ReadFileFS
}

func New(workingPath string) *ErrgoEngine {
	return &ErrgoEngine{
		WorkingPath:    workingPath,
		ErrorTemplates: ErrorTemplates{},
		FS:             &RootFS{},
	}
}

func (e *ErrgoEngine) Analyze(msg string) (*CompiledErrorTemplate, *ContextData, error) {
	template := e.ErrorTemplates.Find(msg)
	if template == nil {
		return nil, nil, fmt.Errorf("template not found. \nMessage: %s", msg)
	}

	// initial context data extraction
	contextData := &ContextData{WorkingPath: e.WorkingPath}
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
	symbolGroupIdx := template.Language.stackTraceRegex.SubexpIndex("symbol")
	pathGroupIdx := template.Language.stackTraceRegex.SubexpIndex("path")
	posGroupIdx := template.Language.stackTraceRegex.SubexpIndex("position")

	stackTraceMatches := template.Language.stackTraceRegex.FindAllStringSubmatch(rawStackTraces, -1)

	for _, submatches := range stackTraceMatches {
		if len(submatches) == 0 {
			continue
		}

		rawSymbolName := submatches[symbolGroupIdx]
		rawPath := submatches[pathGroupIdx]
		rawPos := submatches[posGroupIdx]

		// convert relative paths to absolute for parsing
		if len(e.WorkingPath) != 0 && !filepath.IsAbs(rawPath) {
			rawPath = filepath.Clean(filepath.Join(e.WorkingPath, rawPath))
		}

		stLoc := template.Language.LocationConverter(rawPath, rawPos)
		if contextData.StackTraceGraph == nil {
			contextData.StackTraceGraph = StackTraceGraph{}
		}

		contextData.StackTraceGraph.Add(rawSymbolName, stLoc)
	}

	// open contents of the extracted stack file locations
	parser := sitter.NewParser()

	for _, node := range contextData.StackTraceGraph {
		contents, err := e.FS.ReadFile(node.DocumentPath)
		if err != nil {
			return nil, nil, err
		}

		// check matched languages
		selectedLanguage := template.Language
		if !selectedLanguage.MatchPath(node.DocumentPath) {
			return nil, nil, fmt.Errorf("no language found for %s", node.DocumentPath)
		}

		// compile language first (if not yet)
		selectedLanguage.Compile()

		// do semantic analysis
		parser.SetLanguage(selectedLanguage.SitterLanguage)
		tree, err := parser.ParseCtx(context.Background(), nil, contents)
		if err != nil {
			return nil, nil, err
		}

		doc := contextData.AddDocument(node.DocumentPath, string(contents), selectedLanguage, tree)
		parser.Reset()

		analyzer := &SymbolAnalyzer{
			contextData: contextData,
			doc:         doc,
		}

		analyzer.AnalyzeTree(tree)
	}

	return template, contextData, nil
}

func (e *ErrgoEngine) Translate(template *CompiledErrorTemplate, contextData *ContextData) string {
	// locate main error
	for _, node := range contextData.StackTraceGraph {
		if !strings.HasPrefix(node.DocumentPath, contextData.WorkingPath) {
			continue
		}

		// get nearest node
		doc := contextData.Documents[node.DocumentPath]
		nearest := doc.Tree.RootNode().NamedDescendantForPointRange(
			sitter.Point{Row: uint32(node.Line)},
			sitter.Point{Row: uint32(node.Line)},
		)

		if nearest.StartPoint().Row != uint32(node.Line) {
			cursor := sitter.NewTreeCursor(nearest)
			nearest = locateNearestNode(cursor, node.Position)
		}

		contextData.MainError = MainError{
			ErrorNode: &node,
			Document:  doc,
			Nearest:   WrapNode(doc, nearest),
		}

		break
	}

	// execute error generator function
	explanation := template.OnGenExplainFn(contextData)

	return explanation
}
