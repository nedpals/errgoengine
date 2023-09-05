package errgoengine

type ImportParams struct {
	Node       Node
	CurrentDir string
}

type ResolvedImport struct {
	Path    string
	Name    string
	Symbols []string
}
