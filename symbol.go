package main

type SymbolKind int

const (
	SymbolKindUnknown  SymbolKind = 0
	SymbolKindBuiltin  SymbolKind = iota
	SymbolKindClass    SymbolKind = iota
	SymbolKindFunction SymbolKind = iota
	SymbolKindVariable SymbolKind = iota
	SymbolKindArray    SymbolKind = iota
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

type VariableSymbol struct {
	name       string
	location   Location
	returnType Symbol
}

func (sym VariableSymbol) Name() string {
	return sym.name
}

func (sym VariableSymbol) Kind() SymbolKind {
	return SymbolKindVariable
}

func (sym VariableSymbol) Location() Location {
	return sym.location
}

func (sym VariableSymbol) ReturnType() Symbol {
	return sym.returnType
}

type TopLevelSymbol struct {
	name     string
	kind     SymbolKind
	location Location
	children *SymbolTree
}

func (sym TopLevelSymbol) Name() string {
	return sym.name
}

func (sym TopLevelSymbol) Kind() SymbolKind {
	return sym.kind
}

func (sym TopLevelSymbol) Location() Location {
	return sym.location
}

func (sym TopLevelSymbol) Children() *SymbolTree {
	return sym.children
}

type BuiltinSymbol struct {
	name string
}

func (sym BuiltinSymbol) Name() string {
	return sym.name
}

func (sym BuiltinSymbol) Kind() SymbolKind {
	return SymbolKindBuiltin
}

func (sym BuiltinSymbol) Location() Location {
	return Location{}
}
