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
	ErrorTemplates ErrorTemplates
	FS             fs.ReadFileFS
}

func New() *ErrgoEngine {
	return &ErrgoEngine{
		ErrorTemplates: ErrorTemplates{},
		FS:             &RootFS{},
	}
}

func (e *ErrgoEngine) Analyze(workingPath, msg string) (*CompiledErrorTemplate, *ContextData, error) {
	template := e.ErrorTemplates.Match(msg)
	if template == nil {
		return nil, nil, fmt.Errorf("template not found. \nMessage: %s", msg)
	}

	// initial context data extraction
	contextData := &ContextData{WorkingPath: workingPath}
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

		rawSymbolName := submatches[symbolGroupIdx]
		rawPath := submatches[pathGroupIdx]
		rawPos := submatches[posGroupIdx]

		// convert relative paths to absolute for parsing
		if len(workingPath) != 0 && !filepath.IsAbs(rawPath) {
			rawPath = filepath.Clean(filepath.Join(workingPath, rawPath))
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
