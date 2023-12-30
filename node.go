package errgoengine

import (
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

type SyntaxNode struct {
	*sitter.Node
	isTextCached bool
	Doc          *Document
	text         string
}

func (n SyntaxNode) Debug() {
	fmt.Println("[SyntaxNode.DEBUG]", n, n.Text())
}

func (n SyntaxNode) Text() string {
	if !n.isTextCached && !n.IsNull() {
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

func (n SyntaxNode) Query(q string, d ...any) *QueryNodeCursor {
	return queryNode2(n, fmt.Sprintf(q, d...))
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
	Query  *sitter.Query
	Cursor *sitter.QueryCursor
}

type QueryNodeCursor struct {
	ctx          QueryNodeCtx
	doc          *Document
	cursor       *sitter.TreeCursor
	matchCursor  *QueryMatchIterator
	rawQuery     string
	hasPredicate bool
}

func (c *QueryNodeCursor) Next() bool {
	if c.matchCursor == nil || c.matchCursor.ReachedEnd() {
		if !c.NextMatch() {
			c.cursor.Close()
			return false
		}
	}
	return c.matchCursor.Next()
}

func (c *QueryNodeCursor) NextMatch() bool {
	// use for loop to avoid stack overflow
	for {
		m, ok := c.ctx.Cursor.NextMatch()
		if !ok {
			c.matchCursor = nil
			return false
		}

		// to avoid overhead of calling FilterPredicates if there are no predicates
		if c.hasPredicate {
			match := c.ctx.Cursor.FilterPredicates(m, []byte(c.doc.Contents))
			m = match
		}

		// if there are no captures, skip to the next match
		if len(m.Captures) == 0 {
			continue
		}

		if c.matchCursor == nil {
			c.matchCursor = &QueryMatchIterator{-1, m}
		} else {
			// reuse the same match cursor
			c.matchCursor.match = m
			c.matchCursor.idx = -1
		}
		return true
	}
}

func (c *QueryNodeCursor) Match() *QueryMatchIterator {
	return c.matchCursor
}

func (c *QueryNodeCursor) CurrentNode() SyntaxNode {
	return WrapNode(c.doc, c.matchCursor.Current().Node)
}

func (c *QueryNodeCursor) Query() *sitter.Query {
	return c.ctx.Query
}

func (c *QueryNodeCursor) Len() int {
	if c.matchCursor == nil {
		if !c.NextMatch() {
			return 0
		}
	}
	return len(c.matchCursor.Captures())
}

func (c *QueryNodeCursor) CurrentTagName() string {
	capture := c.matchCursor.Current()
	return c.ctx.Query.CaptureNameForId(capture.Index)
}

type QueryMatchIterator struct {
	idx   int
	match *sitter.QueryMatch
}

func (it *QueryMatchIterator) Next() bool {
	if it.idx+1 >= len(it.match.Captures) {
		return false
	}

	it.idx++
	return true
}

func (it *QueryMatchIterator) Current() sitter.QueryCapture {
	return it.match.Captures[it.idx]
}

func (it *QueryMatchIterator) Captures() []sitter.QueryCapture {
	return it.match.Captures
}

func (it *QueryMatchIterator) ReachedEnd() bool {
	return it.idx+1 >= len(it.match.Captures)
}

func queryNode2(node SyntaxNode, queryR string) *QueryNodeCursor {
	q, err := sitter.NewQuery([]byte(queryR), node.Doc.Language.SitterLanguage)
	if err != nil {
		panic(err)
	}

	queryCursor := sitter.NewQueryCursor()
	queryCursor.Exec(q, node.Node)

	// check if queryR has predicates. usually predicates are used to filter
	// out matches that are not needed and the syntax for predicates starts
	// with "(#" (eg. (#eq? @name "int"), (#match? @name "int"))
	hasPredicate := false
	if strings.Contains(queryR, "(#") {
		hasPredicate = true
	}

	cursor := &QueryNodeCursor{
		ctx:          QueryNodeCtx{q, queryCursor},
		hasPredicate: hasPredicate,
		doc:          node.Doc,
		cursor:       sitter.NewTreeCursor(node.Node),
		rawQuery:     queryR,
	}

	cursor.NextMatch()
	return cursor
}
