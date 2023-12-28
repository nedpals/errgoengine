package errgoengine

import (
	"io"

	sitter "github.com/smacker/go-tree-sitter"
)

type SyntaxNode struct {
	*sitter.Node
	isTextCached bool
	Doc          *Document
	text         string
}

func (n SyntaxNode) Text() string {
	if !n.isTextCached {
		n.isTextCached = true
		n.text = n.Content([]byte(n.Doc.Contents))
	}
	return n.text
}

func (n SyntaxNode) ChildByFieldName(field string) SyntaxNode {
	cNode := n.Node.ChildByFieldName(field)
	return WrapNode(n.Doc, cNode)
}

func (n SyntaxNode) Parent() SyntaxNode {
	return WrapNode(n.Doc, n.Node.Parent())
}

func (n SyntaxNode) NamedChild(idx int) SyntaxNode {
	cNode := n.Node.NamedChild(idx)
	return WrapNode(n.Doc, cNode)
}

func (n SyntaxNode) FirstNamedChild() SyntaxNode {
	return n.NamedChild(0)
}

func (n SyntaxNode) LastNamedChild() SyntaxNode {
	len := n.Node.NamedChildCount()
	return n.NamedChild(int(len) - 1)
}

func (n SyntaxNode) Child(idx int) SyntaxNode {
	cNode := n.Node.Child(idx)
	return WrapNode(n.Doc, cNode)
}

func (n SyntaxNode) PrevSibling() SyntaxNode {
	cNode := n.Node.PrevSibling()
	return WrapNode(n.Doc, cNode)
}

func (n SyntaxNode) PrevNamedSibling() SyntaxNode {
	cNode := n.Node.PrevNamedSibling()
	return WrapNode(n.Doc, cNode)
}

func (n SyntaxNode) NamedDescendantForPointRange(posRange Location) SyntaxNode {
	sRange := posRange.Range()
	cNode := n.Node.NamedDescendantForPointRange(sRange.StartPoint, sRange.EndPoint)
	return WrapNode(n.Doc, cNode)
}

func (n SyntaxNode) StartPosition() Position {
	p := n.Node.StartPoint()
	return Position{
		Line:   int(p.Row),
		Column: int(p.Column),
		Index:  int(n.Node.StartByte()),
	}
}

func (n SyntaxNode) EndPosition() Position {
	p := n.Node.EndPoint()
	return Position{
		Line:   int(p.Row),
		Column: int(p.Column),
		Index:  int(n.Node.EndByte()),
	}
}

func (n SyntaxNode) Location() Location {
	return Location{
		DocumentPath: n.Doc.Path,
		StartPos:     n.StartPosition(),
		EndPos:       n.EndPosition(),
	}
}

func (n SyntaxNode) RawNode() *sitter.Node {
	return n.Node
}

func WrapNode(doc *Document, n *sitter.Node) SyntaxNode {
	return SyntaxNode{
		isTextCached: false,
		text:         "",
		Doc:          doc,
		Node:         n,
	}
}

func nearestNodeFromPos(cursor *sitter.TreeCursor, pos Position) *sitter.Node {
	cursor.GoToFirstChild()
	defer cursor.GoToParent()

	for {
		currentNode := cursor.CurrentNode()
		pointA := currentNode.StartPoint()
		pointB := currentNode.EndPoint()

		if pointA.Row+1 == uint32(pos.Line) {
			return currentNode
		} else if uint32(pos.Line) >= pointA.Row+1 && uint32(pos.Line) <= pointB.Row+1 {
			return nearestNodeFromPos(cursor, pos)
		} else if !cursor.GoToNextSibling() {
			return nil
		}
	}
}

type QueryNodeCtx struct {
	Match  *sitter.QueryMatch
	Query  *sitter.Query
	Cursor *sitter.QueryCursor
}

func QueryNode(rootNode SyntaxNode, queryR io.Reader, callback func(QueryNodeCtx) bool) {
	query, err := io.ReadAll(queryR)
	if err != nil {
		panic(err)
	}

	q, err := sitter.NewQuery(query, rootNode.Doc.Language.SitterLanguage)
	if err != nil {
		panic(err)
	}

	queryCursor := sitter.NewQueryCursor()
	defer queryCursor.Close()

	queryCursor.Exec(q, rootNode.Node)

	for i := 0; ; i++ {
		m, ok := queryCursor.NextMatch()
		if !ok || !callback(QueryNodeCtx{m, q, queryCursor}) {
			break
		}
	}
}
