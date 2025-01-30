package ast

import "larklang.io/lark/pkg/scanner"

type Node interface {
	Pos() scanner.Pos
}

type (
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
		Name    string
		Module  string
	}

	BadExpr struct {
		From scanner.Pos
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
)

func (x *BasicLit) Pos() scanner.Pos   { return x.ValuePos }
func (x *Name) Pos() scanner.Pos       { return x.NamePos }
func (x *QualName) Pos() scanner.Pos   { return x.NodePos }
func (x *BadExpr) Pos() scanner.Pos    { return x.From }
func (x *UnaryExpr) Pos() scanner.Pos  { return x.OpPos }
func (x *BinaryExpr) Pos() scanner.Pos { return x.Lhs.Pos() }
