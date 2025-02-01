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
	case *BadNode, *BasicLit, *Name:
		// nothing to do
	case *QualName:
		Walk(v, n.Name)
		if n.Module != nil {
			Walk(v, n.Module)
		}
	case *UnaryExpr:
		Walk(v, n.Expr)
	case *BinaryExpr:
		Walk(v, n.Lhs)
		Walk(v, n.Rhs)
	case *Import:
		Walk(v, n.Path)
		Walk(v, n.Alias)
	case *ConstDef:
		Walk(v, n.Name)
		Walk(v, n.Expr)
	case *Type:
		Walk(v, n.Name)
		for _, child := range n.Args {
			Walk(v, child)
		}
	case *TypeDef:
		Walk(v, n.Name)
		Walk(v, n.Type)
	case *Module:
		for _, child := range n.Nodes {
			Walk(v, child)
		}
	default:
		panic(fmt.Sprintf("ast.Walk: unexpected node type %T", n))
	}
}
