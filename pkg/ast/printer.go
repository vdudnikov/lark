package ast

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type printer struct {
	indent int
	writer io.Writer
}

func (p *printer) Visit(node Node) Visitor {
	indent := p.indent
	switch n := node.(type) {
	case *BasicLit:
		p.printf("BasicLit: Kind=%s, Value=%s, Pos=%v", n.Kind, n.Value, node.Pos())
	case *Name:
		p.printf("Name: Name=%s, Pos=%v", n.Name, node.Pos())
	case *QualName:
		p.printf("QualName: Name=%s, Module=%s, Pos=%v", n.Name, n.Module, node.Pos())
	case *BadExpr:
		p.printf("BadExpr: Pos=%v", n.Pos())
	case *UnaryExpr:
		p.printf("UnaryExpr: Op=%s, Pos=%v", n.Op, node.Pos())
		indent++
	case *BinaryExpr:
		p.printf("BinaryExpr: Op=%s, Pos=%v", n.Op, n.Pos())
		indent++
	default:
		panic(fmt.Sprintf("ast.Print: unexpected node type %T", n))
	}

	return &printer{indent, p.writer}
}

func (p *printer) printf(format string, args ...any) {
	fmt.Println(strings.Repeat("  ", p.indent) + fmt.Sprintf(format, args...))
}

func Fprint(writer io.Writer, node Node) {
	printer := &printer{writer: writer}
	Walk(printer, node)
}

func Print(node Node) {
	Fprint(os.Stdout, node)
}
