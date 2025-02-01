package ast

import "larklang.io/lark/pkg/scanner"

type Node interface {
	Pos() scanner.Pos
}

type (
	BadNode struct {
		From scanner.Pos
	}

	BasicLit struct {
		Kind     scanner.TokenKind
		ValuePos scanner.Pos
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

	Import struct {
		ImportPos scanner.Pos
		Path      *BasicLit
		Alias     *Name
	}

	ConstDef struct {
		ConstPos scanner.Pos
		Name     *Name
		Expr     Node
	}

	Type struct {
		Name *QualName
		Args []Node
	}

	TypeDef struct {
		TypePos scanner.Pos
		Name    *Name
		Type    *Type
	}

	Field struct {
		Name *Name
		Type *Type
	}

	StructDef struct {
		StructPos scanner.Pos
		Name      *Name
		Fields    []*Field
	}

	Module struct {
		Nodes []Node
	}
)

func (x *BasicLit) Pos() scanner.Pos   { return x.ValuePos }
func (x *Name) Pos() scanner.Pos       { return x.NamePos }
func (x *QualName) Pos() scanner.Pos   { return x.NodePos }
func (x *BadNode) Pos() scanner.Pos    { return x.From }
func (x *UnaryExpr) Pos() scanner.Pos  { return x.OpPos }
func (x *BinaryExpr) Pos() scanner.Pos { return x.Lhs.Pos() }
func (x *Import) Pos() scanner.Pos     { return x.ImportPos }
func (x *ConstDef) Pos() scanner.Pos   { return x.ConstPos }
func (x *Type) Pos() scanner.Pos       { return x.Name.Pos() }
func (x *TypeDef) Pos() scanner.Pos    { return x.TypePos }
func (x *Field) Pos() scanner.Pos      { return x.Name.Pos() }
func (x *StructDef) Pos() scanner.Pos  { return x.StructPos }
func (x *Module) Pos() scanner.Pos     { return scanner.Pos{Line: 0, Column: 0} }
