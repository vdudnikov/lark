package parser

import (
	"fmt"

	"larklang.io/lark/pkg/ast"
	"larklang.io/lark/pkg/scanner"
)

type ErrorInfo struct {
	Pos     scanner.Pos
	Message string
}

type Parsed struct {
	// Module *ast.Module
	Module ast.Node
	Lines  []string
	Errors []ErrorInfo
}

type nudFn func() ast.Node
type ledFn func(lhs ast.Node, prec int) ast.Node
type parseExprRule struct {
	nud  nudFn
	led  ledFn
	prec int // only for led
}

const (
	precNone     = iota
	precLogicOr  // a||b
	precLogicAnd // a&&b
	precCmp      // a==b, a!=b, a<b, a<=b, a>b, a>=b
	precTerm     // a+b, a-b
	precFactor   // a*b, a/b, a%b
	precUnary    // !a, -a
	precPrimary  // bool, string, float, integer
)

type parser struct {
	scanner       *scanner.Scanner
	current       scanner.Token
	exprRuleTable map[scanner.TokenKind]parseExprRule
	errors        []ErrorInfo
}

func (p *parser) init(text []byte) {
	p.scanner = scanner.New(text, p.err)

	p.exprRuleTable = map[scanner.TokenKind]parseExprRule{
		scanner.NULL:       {p.parseBasicLit, nil, precNone},
		scanner.TRUE:       {p.parseBasicLit, nil, precNone},
		scanner.FALSE:      {p.parseBasicLit, nil, precNone},
		scanner.STRING:     {p.parseBasicLit, nil, precNone},
		scanner.INTEGER:    {p.parseBasicLit, nil, precNone},
		scanner.IDENTIFIER: {p.parseQualNameExpr, nil, precNone},
		scanner.FLOAT:      {p.parseBasicLit, nil, precNone},
		scanner.MINUS:      {p.parseUnaryExpr, p.parseBinaryExpr, precTerm},
		scanner.NOT:        {p.parseUnaryExpr, nil, precNone},
		scanner.PLUS:       {nil, p.parseBinaryExpr, precTerm},
		scanner.MULT:       {nil, p.parseBinaryExpr, precFactor},
		scanner.DIV:        {nil, p.parseBinaryExpr, precFactor},
		scanner.MOD:        {nil, p.parseBinaryExpr, precFactor},
		scanner.AND:        {nil, p.parseBinaryExpr, precLogicAnd},
		scanner.OR:         {nil, p.parseBinaryExpr, precLogicOr},
		scanner.EQ:         {nil, p.parseBinaryExpr, precCmp},
		scanner.GE:         {nil, p.parseBinaryExpr, precCmp},
		scanner.GT:         {nil, p.parseBinaryExpr, precCmp},
		scanner.LE:         {nil, p.parseBinaryExpr, precCmp},
		scanner.LT:         {nil, p.parseBinaryExpr, precCmp},
		scanner.NEQ:        {nil, p.parseBinaryExpr, precCmp},
	}

	p.next()
}

func (p *parser) scan(newline bool) scanner.Token {
	for {
		token := p.scanner.Scan()
		switch token.Kind {
		case scanner.COMMENT, scanner.ILLEGAL:
			continue
		case scanner.NEWLINE:
			if newline {
				return token
			}
		default:
			return token
		}
	}
}

var insert_semi = [...]bool{
	scanner.RIGHT_BRACE: true,
	scanner.RIGHT_BRACK: true,
	scanner.RIGHT_PAREN: true,
	scanner.INTEGER:     true,
	scanner.FLOAT:       true,
	scanner.IDENTIFIER:  true,
	scanner.STRING:      true,
	scanner.TRUE:        true,
	scanner.FALSE:       true,
	scanner.NULL:        true,
}

func (p *parser) next() {
	token := p.scan(true)
	if token.Kind == scanner.NEWLINE || token.Kind == scanner.ENDMARKER {
		if insert_semi[p.current.Kind] {
			token.Kind = scanner.SEMICOLON
		} else {
			token = p.scan(false)
		}
	}

	p.current = token
}

func (p *parser) err(pos scanner.Pos, msg string) {
	p.errors = append(p.errors, ErrorInfo{pos, msg})
}

func (p *parser) errf(pos scanner.Pos, format string, args ...any) {
	p.err(pos, fmt.Sprintf(format, args...))
}

func (p *parser) expectMsg(msg string) {
	p.err(p.current.Pos, fmt.Sprintf("expected %s, found '%s'", msg, p.current.Value))
}

func (p *parser) expect(kind scanner.TokenKind) scanner.Token {
	token := p.current
	if token.Kind != kind {
		p.expectMsg("'" + kind.String() + "'")
	}
	p.next() // make progress
	return token
}

func (p *parser) accept(kind scanner.TokenKind) bool {
	if p.current.Kind == kind {
		p.next()
		return true
	}
	return false
}

func (p *parser) parseItems(nud nudFn, stop scanner.TokenKind) []ast.Node {
	if p.accept(stop) {
		return nil
	}

	items := []ast.Node{nud()}
	for p.accept(scanner.COMMA) {
		// trailing comma
		if p.current.Kind == stop {
			break
		}
		items = append(items, nud())
	}

	p.expect(stop)
	return items
}

func (p *parser) parseExpr(prec int) ast.Node {
	token := p.current
	prefRule := p.exprRuleTable[token.Kind]
	if prefRule.nud == nil {
		p.expectMsg("expression")
		p.next()
		return &ast.BadNode{From: token.Pos}
	}

	root := prefRule.nud()
	token = p.current
	infRule := p.exprRuleTable[token.Kind]
	for infRule.prec > prec {
		// We do not check if infRule.led != nil because, for any random token
		// that is not an infix operator, the precedence will be 0, and the
		// loop will terminate.
		root = infRule.led(root, infRule.prec)
		token = p.current
		infRule = p.exprRuleTable[token.Kind]
	}

	return root
}

func (p *parser) parseBasicLit() ast.Node {
	lit := p.current
	p.next()

	return &ast.BasicLit{Kind: lit.Kind, ValuePos: lit.Pos, Value: lit.Value}
}

func (p *parser) parseName() *ast.Name {
	identifier := p.expect(scanner.IDENTIFIER)
	name := "_"
	if identifier.Kind == scanner.IDENTIFIER {
		name = identifier.Value
	}
	return &ast.Name{NamePos: identifier.Pos, Name: name}
}

func (p *parser) parseQualName() *ast.QualName {
	tmp := p.parseName()
	if p.accept(scanner.DOT) {
		name := p.parseName()
		return &ast.QualName{NodePos: tmp.NamePos, Name: name, Module: tmp}
	}

	return &ast.QualName{NodePos: tmp.NamePos, Name: tmp}
}

func (p *parser) parseQualNameExpr() ast.Node {
	return p.parseQualName()
}

func (p *parser) parseUnaryExpr() ast.Node {
	op := p.current
	p.next()
	return &ast.UnaryExpr{OpPos: op.Pos, Op: op.Kind, Expr: p.parseExpr(precUnary)}
}

func (p *parser) parseBinaryExpr(lhs ast.Node, prec int) ast.Node {
	op := p.current
	p.next()

	return &ast.BinaryExpr{Op: op.Kind, Lhs: lhs, Rhs: p.parseExpr(prec)}
}

func (p *parser) parseImport() *ast.Import {
	importKw := p.expect(scanner.IMPORT)
	path := p.expect(scanner.STRING)
	var alias *ast.Name
	if p.accept(scanner.AS) {
		alias = p.parseName()
	}

	return &ast.Import{
		ImportPos: importKw.Pos,
		Path:      &ast.BasicLit{Kind: scanner.STRING, ValuePos: path.Pos, Value: path.Value},
		Alias:     alias,
	}
}

func (p *parser) parseConstDef() *ast.ConstDef {
	constKw := p.expect(scanner.CONST)
	name := p.parseName()
	p.expect(scanner.ASSIGN)
	expr := p.parseExpr(precNone)

	return &ast.ConstDef{ConstPos: constKw.Pos, Name: name, Expr: expr}
}

func (p *parser) parseType() *ast.Type {
	parseItem := func() ast.Node {
		return p.parseType()
	}

	name := p.parseQualName()
	var args []ast.Node
	if p.accept(scanner.LEFT_BRACK) {
		args = p.parseItems(parseItem, scanner.RIGHT_BRACK)
	}

	return &ast.Type{Name: name, Args: args}
}

func (p *parser) parseTypeDef() *ast.TypeDef {
	typeKw := p.expect(scanner.TYPE)
	name := p.parseName()
	p.expect(scanner.ASSIGN)
	anytype := p.parseType()

	return &ast.TypeDef{TypePos: typeKw.Pos, Name: name, Type: anytype}
}

func (p *parser) parseField() *ast.Field {
	name := p.parseName()
	p.expect(scanner.COLON)
	anytype := p.parseType()

	return &ast.Field{Name: name, Type: anytype}
}

func (p *parser) parseStructDef() *ast.StructDef {
	structKw := p.expect(scanner.STRUCT)
	name := p.parseName()
	p.expect(scanner.LEFT_BRACE)

	var fields []*ast.Field
	for p.current.Kind != scanner.RIGHT_BRACE && p.current.Kind != scanner.ENDMARKER {
		fields = append(fields, p.parseField())
		p.expect(scanner.SEMICOLON)
	}

	p.expect(scanner.RIGHT_BRACE)

	return &ast.StructDef{StructPos: structKw.Pos, Name: name, Fields: fields}
}

func (p *parser) parse() *ast.Module {
	var nodes []ast.Node
	ndef := 0
	for p.current.Kind != scanner.ENDMARKER {
		var node ast.Node
		switch p.current.Kind {
		case scanner.IMPORT:
			node = p.parseImport()
			if ndef > 0 {
				p.errf(node.Pos(), "imports must appear before other declarations")
			}
		default:
			ndef++
			switch p.current.Kind {
			case scanner.CONST:
				node = p.parseConstDef()
			case scanner.TYPE:
				node = p.parseTypeDef()
			case scanner.STRUCT:
				node = p.parseStructDef()
			default:
				node = &ast.BadNode{From: p.current.Pos}
				p.expectMsg("definition")
				p.next()
			}
		}

		nodes = append(nodes, node)
		p.expect(scanner.SEMICOLON)
	}

	return &ast.Module{Nodes: nodes}
}

func Parse(text []byte) Parsed {
	p := &parser{}
	p.init(text)

	return Parsed{
		Module: p.parse(),
		Lines:  p.scanner.Lines(),
		Errors: p.errors,
	}
}
