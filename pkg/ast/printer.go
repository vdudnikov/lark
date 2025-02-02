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
	case *BadNode:
		p.printf("BadNode: Pos=%v", n.Pos())
	case *BasicLit:
		p.printf("BasicLit: Kind=%s, Value=%s, Pos=%v", n.Kind, n.Value, node.Pos())
	case *Name:
		p.printf("Name: Name=%s, Pos=%v", n.Name, node.Pos())
	case *QualName:
		module := ""
		if n.Module != nil {
			module = n.Module.Name
		}
		p.printf("QualName: Module=%s, Name=%s, Pos=%v", module, n.Name.Name, node.Pos())
		return nil
	case *UnaryExpr:
		p.printf("UnaryExpr: Op=%s, Pos=%v", n.Op, node.Pos())
		indent++
	case *BinaryExpr:
		p.printf("BinaryExpr: Op=%s, Pos=%v", n.Op, n.Pos())
		indent++
	case *Import:
		alias := ""
		if n.Alias != nil {
			alias = n.Alias.Name
		}
		p.printf("Import: Path=%s, Alias=%s, Pos=%v", n.Path.Value, alias, n.Pos())
        return nil
	case *ConstDef:
		p.printf("Const: Pos=%v", n.Pos())
		indent++
	case *Type:
		p.printf("Type: Pos=%v", n.Pos())
		indent++
	case *TypeDef:
		p.printf("TypeDef: Pos=%v", n.Pos())
		indent++
	case *Field:
		p.printf("Field: Pos=%v", n.Pos())
		indent++
	case *StructDef:
		p.printf("StructDef: Pos=%v", n.Pos())
		indent++
	case *Module:
		// nothing to do
	default:
		panic(fmt.Sprintf("ast.Print: unexpected node type %T", n))
	}

	return &printer{indent, p.writer}
}

func (p *printer) Exit(node Node) {
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
