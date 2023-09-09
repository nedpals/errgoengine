package errgoengine

type ImportParams struct {
	Node       SyntaxNode
	CurrentDir string
}

type ResolvedImport struct {
	Path    string
	Name    string
	Symbols []string
}
