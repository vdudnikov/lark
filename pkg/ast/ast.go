package ast

import "larklang.io/lark/pkg/scanner"

type Node interface {
	Pos() scanner.Pos
}

type (
	BadNode struct {
		From scanner.Pos
		To   scanner.Pos
	}

	BasicLit struct {
		ValuePos scanner.Pos
		Kind     scanner.TokenKind
		Value    string
	}

	Name struct {
		NamePos scanner.Pos
		Name    string
	}

	QualName struct {
		NodePos scanner.Pos
		Name    *Name
		Module  *Name
	}

	UnaryExpr struct {
		OpPos scanner.Pos
		Op    scanner.TokenKind
		Expr  Node
	}

	BinaryExpr struct {
		Op  scanner.TokenKind
		Lhs Node
		Rhs Node
	}

	ImportSpec struct {
		Path  *BasicLit
		Alias *Name
	}

	ConstSpec struct {
		Name *Name
		Expr Node
	}

	Type struct {
		Name *QualName
		Args []Node
	}

	TypeAlias struct {
		TypePos scanner.Pos
		Name    *Name
		Type    *Type
	}

	Field struct {
		Name *Name
		Type *Type
	}

	Struct struct {
		StructPos scanner.Pos
		Name      *Name
		Fields    []*Field
	}

	File struct {
		Nodes []Node
	}
)

func (x *BasicLit) Pos() scanner.Pos   { return x.ValuePos }
func (x *Name) Pos() scanner.Pos       { return x.NamePos }
func (x *QualName) Pos() scanner.Pos   { return x.NodePos }
func (x *BadNode) Pos() scanner.Pos    { return x.From }
func (x *UnaryExpr) Pos() scanner.Pos  { return x.OpPos }
func (x *BinaryExpr) Pos() scanner.Pos { return x.Lhs.Pos() }
func (x *ImportSpec) Pos() scanner.Pos { return x.Path.Pos() }
func (x *ConstSpec) Pos() scanner.Pos  { return x.Name.Pos() }
func (x *Type) Pos() scanner.Pos       { return x.Name.Pos() }
func (x *TypeAlias) Pos() scanner.Pos  { return x.TypePos }
func (x *Field) Pos() scanner.Pos      { return x.Name.Pos() }
func (x *Struct) Pos() scanner.Pos     { return x.StructPos }
func (x *File) Pos() scanner.Pos     { return scanner.Pos{Line: 0, Column: 0} }
