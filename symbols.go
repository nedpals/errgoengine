package errgoengine

import "fmt"

type SymbolKind int

func (kind SymbolKind) String() string {
	switch kind {
	case SymbolKindBuiltin:
		return "builtin"
	case SymbolKindClass:
		return "class"
	case SymbolKindType:
		return "type"
	case SymbolKindFunction:
		return "function"
	case SymbolKindVariable:
		return "variable"
	case SymbolKindAssignment:
		return "assignment"
	case SymbolKindImport:
		return "import"
	default:
		return fmt.Sprintf("unknown(%d)", kind)
	}
}

func NewSymbolKindFromString(str string) SymbolKind {
	switch str {
	case "builtin":
		return SymbolKindBuiltin
	case "class":
		return SymbolKindClass
	case "function":
		return SymbolKindFunction
	case "variable":
		return SymbolKindVariable
	case "assignment":
		return SymbolKindAssignment
	case "array":
		return SymbolKindType
	case "import":
		return SymbolKindImport
	default:
		return SymbolKindUnknown
	}
}

const (
	SymbolKindUnknown    SymbolKind = 0
	SymbolKindUnresolved SymbolKind = iota
	SymbolKindBuiltin    SymbolKind = iota
	SymbolKindClass      SymbolKind = iota
	SymbolKindFunction   SymbolKind = iota
	SymbolKindVariable   SymbolKind = iota
	SymbolKindAssignment SymbolKind = iota
	SymbolKindType       SymbolKind = iota
	SymbolKindImport     SymbolKind = iota
)

type Symbol interface {
	Name() string
	Kind() SymbolKind
	Location() Location
}

type IReturnableSymbol interface {
	Symbol
	ReturnType() Symbol
}

func CastSymbolReturnable(sym Symbol) IReturnableSymbol {
	if retSym, ok := any(sym).(IReturnableSymbol); ok {
		return retSym
	}
	return nil
}

type IChildrenSymbol interface {
	Symbol
	Children() *SymbolTree
}

func GetFromSymbol(sym IChildrenSymbol, field string) Symbol {
	if sym.Children() != nil {
		for symName, sym := range sym.Children().Symbols {
			if symName == field {
				return sym
			}
		}
	}
	return nil
}

func CastChildrenSymbol(sym Symbol) IChildrenSymbol {
	if cSym, ok := any(sym).(IChildrenSymbol); ok {
		return cSym
	}
	return nil
}

type AssignmentSymbol struct {
	Variable          Symbol
	FallbackName      string
	Location_         Location
	ContentReturnType Symbol
}

func (sym AssignmentSymbol) Name() string {
	if sym.Variable != nil {
		return sym.Variable.Name()
	}
	return sym.FallbackName
}

func (sym AssignmentSymbol) Kind() SymbolKind {
	return SymbolKindAssignment
}

func (sym AssignmentSymbol) Location() Location {
	return sym.Location_
}

func (sym AssignmentSymbol) ReturnType() Symbol {
	return sym.ContentReturnType
}

type VariableSymbol struct {
	Name_       string
	Location_   Location
	ReturnType_ Symbol
}

func (sym VariableSymbol) Name() string {
	return sym.Name_
}

func (sym VariableSymbol) Kind() SymbolKind {
	return SymbolKindVariable
}

func (sym VariableSymbol) Location() Location {
	return sym.Location_
}

func (sym VariableSymbol) ReturnType() Symbol {
	return sym.ReturnType_
}

type TopLevelSymbol struct {
	Name_     string
	Kind_     SymbolKind
	Location_ Location
	Children_ *SymbolTree
}

func (sym TopLevelSymbol) Name() string {
	return sym.Name_
}

func (sym TopLevelSymbol) Kind() SymbolKind {
	return sym.Kind_
}

func (sym TopLevelSymbol) Location() Location {
	return sym.Location_
}

func (sym TopLevelSymbol) Children() *SymbolTree {
	return sym.Children_
}

type BuiltinSymbol struct {
	Name_ string
}

func (sym BuiltinSymbol) Name() string {
	return sym.Name_
}

func (sym BuiltinSymbol) Kind() SymbolKind {
	return SymbolKindBuiltin
}

func (sym BuiltinSymbol) Location() Location {
	return Location{}
}

func Builtin(name string) Symbol {
	return BuiltinSymbol{Name_: name}
}

type ImportSymbol struct {
	Alias           string
	Node            *DepNode
	ImportedSymbols []string
}

func (sym ImportSymbol) Name() string {
	return sym.Alias
}

func (sym ImportSymbol) Kind() SymbolKind {
	return SymbolKindImport
}

func (sym ImportSymbol) Location() Location {
	return Location{
		DocumentPath: sym.Node.Path,
		StartPos: Position{
			Line:   0,
			Column: 0,
			Index:  0,
		},
		EndPos: Position{
			Line:   0,
			Column: 0,
			Index:  0,
		},
	}
}

type unresolvedSymbol struct{}

func (sym unresolvedSymbol) Name() string {
	return "unresolved"
}

func (sym unresolvedSymbol) Kind() SymbolKind {
	return SymbolKindUnresolved
}

func (sym unresolvedSymbol) Location() Location {
	return Location{}
}

var UnresolvedSymbol Symbol = unresolvedSymbol{}

// TODO:
// func (sym ImportSymbol) Children() *SymbolTree {
// 	// TODO:
// 	return nil
// }
