package java

import (
	"fmt"
	"strings"

	lib "github.com/nedpals/errgoengine"
)

type privateAccessErrorCtx struct {
	ClassDeclarationNode lib.SyntaxNode
}

var PrivateAccessError = lib.ErrorTemplate{
	Name:              "PrivateAccessError",
	Pattern:           comptimeErrorPattern(`(?P<field>\S+) has private access in (?P<class>\S+)`),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		pCtx := privateAccessErrorCtx{}
		className := cd.Variables["class"]
		rootNode := lib.WrapNode(m.Nearest.Doc, m.Nearest.Doc.Tree.RootNode())

		// locate the right node first
		query := fmt.Sprintf(`((field_access (identifier) . (identifier) @field-name) @field (#eq? @field-name "%s"))`, cd.Variables["field"])
		lib.QueryNode(rootNode, strings.NewReader(query), func(ctx lib.QueryNodeCtx) bool {
			match := ctx.Cursor.FilterPredicates(ctx.Match, []byte(m.Nearest.Doc.Contents))
			for _, c := range match.Captures {
				node := lib.WrapNode(m.Nearest.Doc, c.Node)
				m.Nearest = node
				return false
			}
			return true
		})

		// get class declaration node
		classQuery := fmt.Sprintf(`(class_declaration name: (identifier) @class-name (#eq? @class-name "%s")) @class`, className)
		lib.QueryNode(rootNode, strings.NewReader(classQuery), func(ctx lib.QueryNodeCtx) bool {
			match := ctx.Cursor.FilterPredicates(ctx.Match, []byte(m.Nearest.Doc.Contents))
			for _, c := range match.Captures {
				node := lib.WrapNode(m.Nearest.Doc, c.Node)
				pCtx.ClassDeclarationNode = node
				return false
			}
			return true
		})

		m.Context = pCtx
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("This error occurs when you try to access a private variable from another class, which is not allowed.")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		ctx := cd.MainError.Context.(privateAccessErrorCtx)

		// fmt.Println(cd.Analyzer.AnalyzeNode(cd.MainError.Nearest))
		classDeclScope := cd.InitOrGetSymbolTree(cd.MainDocumentPath()).GetNearestScopedTree(ctx.ClassDeclarationNode.EndPosition().Index)

		gen.Add("Use a public accessor method", func(s *lib.BugFixSuggestion) {
			methodCreatorSb := &strings.Builder{}
			bodyNode := ctx.ClassDeclarationNode.ChildByFieldName("body")

			// get return type of the private field
			privateVarRetType := cd.Analyzer.FallbackSymbol()
			if gotSym := classDeclScope.Find(cd.Variables["field"]); gotSym != nil {
				if gotSym, ok := gotSym.(lib.IReturnableSymbol); ok {
					privateVarRetType = gotSym.ReturnType()
				}
			}

			// create the method
			space := ""
			lastNamedChild := bodyNode.LastNamedChild()
			if !lastNamedChild.IsNull() {
				targetPos := lastNamedChild.StartPosition()
				// get the last named child and use the space as its base
				space = cd.MainError.Document.LineAt(targetPos.Line)[:targetPos.Column]
				methodCreatorSb.WriteString("\n\n" + space)
			}

			accessorMethodName := "get" + strings.ToUpper(string(cd.Variables["field"][0])) + cd.Variables["field"][1:]
			methodCreatorSb.WriteString(fmt.Sprintf("public %s %s() {\n", privateVarRetType.Name(), accessorMethodName))
			methodCreatorSb.WriteString(strings.Repeat(space, 2) + fmt.Sprintf("return this.%s;\n", cd.Variables["field"]))
			methodCreatorSb.WriteString(space + "}\n")

			targetPos := lastNamedChild.EndPosition()
			s.AddStep("To access a private variable from another class, create a public accessor method in `%s`.", cd.Variables["class"]).
				AddFix(lib.FixSuggestion{
					NewText:       methodCreatorSb.String(),
					StartPosition: lib.Position{Line: targetPos.Line, Column: targetPos.Column},
					EndPosition:   lib.Position{Line: targetPos.Line, Column: targetPos.Column},
				})

			fieldNode := cd.MainError.Nearest.ChildByFieldName("field")

			s.AddStep("Then, use this method to get the value.").
				AddFix(lib.FixSuggestion{
					NewText:       accessorMethodName + "()",
					StartPosition: fieldNode.StartPosition(),
					EndPosition:   fieldNode.EndPosition(),
					Description:   "This way, you respect encapsulation by using a method to access the private variable.",
				})
		})

		gen.Add("Make the variable public (not recommended)", func(s *lib.BugFixSuggestion) {
			targetLoc := lib.Location{}
			if gotSym := classDeclScope.Find(cd.Variables["field"]); gotSym != nil {
				// get the node within that position
				rawDescendantNode := cd.MainError.Document.Tree.RootNode().NamedDescendantForPointRange(
					gotSym.Location().Range().StartPoint,
					gotSym.Location().Range().EndPoint,
				)

				if rawDescendantNode.Type() == "variable_declarator" {
					rawDescendantNode = rawDescendantNode.Parent()
				}

				rawFirstChild := rawDescendantNode.NamedChild(0)
				if firstChild := lib.WrapNode(cd.MainError.Document, rawFirstChild); firstChild.Type() == "modifiers" {
					targetLoc = firstChild.Location()
				} else {
					targetLoc.StartPos = firstChild.StartPosition()
					targetLoc.EndPos = targetLoc.StartPos
				}
			}

			newText := "public"
			if targetLoc.StartPos.Eq(targetLoc.EndPos) {
				newText += " "
			}

			s.AddStep("If you must access the variable directly, you can make it public, but this is generally not recommended for maintaining encapsulation.").
				AddFix(lib.FixSuggestion{
					NewText:       newText,
					StartPosition: targetLoc.StartPos,
					EndPosition:   targetLoc.EndPos,
				})
		})
	},
}
