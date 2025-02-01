package ast

import "fmt"

type Visitor interface {
	Visit(node Node) (w Visitor)
	Exit(node Node)
}

func Walk(v Visitor, node Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	defer v.Exit(node)

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
		if n.Alias != nil {
			Walk(v, n.Alias)
		}
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
	case *Field:
		Walk(v, n.Name)
		Walk(v, n.Type)
	case *StructDef:
		Walk(v, n.Name)
		for _, child := range n.Fields {
			Walk(v, child)
		}
	case *Module:
		for _, child := range n.Nodes {
			Walk(v, child)
		}
	default:
		panic(fmt.Sprintf("ast.Walk: unexpected node type %T", n))
	}
}
