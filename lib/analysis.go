package lib

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

type NodeValueAnalyzer interface {
	FindSymbol(name string, pos int) Symbol
	AnalyzeValue(n Node) Symbol
}

func locateNearestNode(cursor *sitter.TreeCursor, pos Position) *sitter.Node {
	cursor.GoToFirstChild()
	defer cursor.GoToParent()

	for {
		currentNode := cursor.CurrentNode()
		pointA := currentNode.StartPoint()
		pointB := currentNode.EndPoint()

		if pointA.Row+1 == uint32(pos.Line) {
			return currentNode
		} else if uint32(pos.Line) >= pointA.Row+1 && uint32(pos.Line) <= pointB.Row+1 {
			return locateNearestNode(cursor, pos)
		} else if !cursor.GoToNextSibling() {
			return nil
		}
	}
}

func TranslateError(template ErrorTemplate, contextData *ContextData) string {
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

func Analyze(errorTemplates ErrorTemplates, workingPath string, msg string) string {
	template := errorTemplates.Find(msg)
	if template == nil {
		panic("Template not found! \nMessage: \n" + msg)
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
	symbolGroupIdx := template.Language.StackTracePattern.SubexpIndex("symbol")
	pathGroupIdx := template.Language.StackTracePattern.SubexpIndex("path")
	posGroupIdx := template.Language.StackTracePattern.SubexpIndex("position")

	for _, rawStackTrace := range strings.Split(rawStackTraces, "\n") {
		submatches := template.Language.StackTracePattern.FindAllStringSubmatch(rawStackTrace, -1)
		if len(submatches) == 0 {
			continue
		}

		matches := submatches[0]
		rawSymbolName := matches[symbolGroupIdx]
		rawPath := matches[pathGroupIdx]
		rawPos := matches[posGroupIdx]

		// convert relative paths to absolute for parsing
		if !filepath.IsAbs(rawPath) {
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
		contents, err := os.ReadFile(node.DocumentPath)
		if err != nil {
			panic(err)
		}

		// check matched languages
		selectedLanguage := template.Language
		if !selectedLanguage.MatchPath(node.DocumentPath) {
			panic("No language found for " + node.DocumentPath)
		}

		// compile builtin types first of a language if not available
		selectedLanguage.Compile()

		// do semantic analysis
		parser.SetLanguage(selectedLanguage.SitterLanguage)
		tree, err := parser.ParseCtx(context.Background(), nil, contents)
		if err != nil {
			panic(err)
		}

		doc := contextData.AddDocument(node.DocumentPath, string(contents), selectedLanguage, tree)
		parser.Reset()

		analyzer := &SymbolAnalyzer{
			contextData: contextData,
			doc:         doc,
		}

		analyzer.AnalyzeTree(tree)
	}

	// error translation
	return TranslateError(template.ErrorTemplate, contextData)
}
