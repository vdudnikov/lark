package ast

import "fmt"

type Visitor interface {
	Visit(node Node) (w Visitor)
}

func Walk(v Visitor, node Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	// walk children
	switch n := node.(type) {
	case *UnaryExpr:
		Walk(v, n.Expr)
	case *BinaryExpr:
		Walk(v, n.Lhs)
		Walk(v, n.Rhs)
	case *BadExpr, *BasicLit, *Name, *QualName:
		// nothing to do
	default:
		panic(fmt.Sprintf("ast.Walk: unexpected node type %T", n))
	}
}
