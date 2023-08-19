package lib

import sitter "github.com/smacker/go-tree-sitter"

type Position struct {
	Line   int
	Column int
	Index  int
}

type Location struct {
	DocumentPath string
	Position
}

type Document struct {
	Path     string
	Contents string
	Language *Language
	Tree     *sitter.Tree
}
