package lib

// TODO: turn into a real graph
type StackTraceGraph []STGNode

func (g *StackTraceGraph) Add(symbolName string, loc Location) {
	*g = append(*g, STGNode{
		Location:   loc,
		SymbolName: symbolName,
		IsMain:     len(*g) == 0,
	})
}

type STGNode struct {
	Location
	SymbolName string
	IsMain     bool
}
