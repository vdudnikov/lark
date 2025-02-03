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

type ParsedFile struct {
	File    *ast.File
	Imports []*ast.ImportSpec
	Symtab  []Symbol
	Lines   []string
	Errors  []ErrorInfo
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

type SymbolType int

const (
	ConstSym SymbolType = iota
	FuncSym
	InterfaceSym
	StructSym
	AliasSym
)

type Symbol struct {
	Type SymbolType
	Name *ast.Name
	Decl ast.Node
}

type parser struct {
	scanner       *scanner.Scanner
	current       scanner.Token
	exprRuleTable map[scanner.TokenKind]parseExprRule
	errors        []ErrorInfo

	// Error recovery
	// (used to limit the number of calls to parser.advance
	// w/o making scanning progress - avoids potential endless
	// loops across multiple parser functions during error recovery)
	syncPos scanner.Pos // last synchronization position
	syncCnt int         // number of parser.advance calls without progress

	imports []*ast.ImportSpec
	symtab  []Symbol
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

var semiOnly = map[scanner.TokenKind]bool{
	scanner.SEMICOLON: true,
}

var declStart = map[scanner.TokenKind]bool{
	scanner.CONST:     true,
	scanner.FUNC:      true,
	scanner.IMPORT:    true,
	scanner.INTERFACE: true,
	scanner.STRUCT:    true,
	scanner.TYPE:      true,
}

// sync consumes tokens until the current token is in the 'to' set, or
// scanner.ENDMARKER. For error recovery.
func (p *parser) sync(to map[scanner.TokenKind]bool) {
	for ; p.current.Kind != scanner.ENDMARKER; p.next() {
		token := p.current
		if to[token.Kind] {
			if token.Pos == p.syncPos && p.syncCnt < 10 {
				p.syncCnt++
				return
			}
			if token.Pos.Greater(p.syncPos) {
				p.syncPos = p.current.Pos
				p.syncCnt = 0
				return
			}
		}
	}
}

// parseExpr parses an expression using the TDOP (Top-Down Operator Precedence)
// method. It processes prefix (nud) and infix (led) rules based on token
// precedence to construct an AST. If the current token has no valid prefix
// rule, an error node is returned. The function ensures correct operator
// precedence handling by iterating while the next token has a higher precedence.
func (p *parser) parseExpr(prec int) ast.Node {
	token := p.current
	prefRule := p.exprRuleTable[token.Kind]
	if prefRule.nud == nil {
		p.expectMsg("expression")
		p.next()
		return &ast.BadNode{From: token.Pos, To: p.current.Pos}
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
	name := "@"
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

func (p *parser) parseImportSpec() ast.Node {
	token := p.current
	var path string
	if token.Kind == scanner.STRING {
		path = token.Value
		p.next()
	} else {
		p.err(token.Pos, "import path must be a string")
		p.sync(semiOnly)
	}

	var alias *ast.Name
	if p.accept(scanner.AS) {
		alias = p.parseName()
	}

	spec := &ast.ImportSpec{
		Path:  &ast.BasicLit{ValuePos: token.Pos, Kind: scanner.STRING, Value: path},
		Alias: alias,
	}
	p.imports = append(p.imports, spec)

	return spec
}

func (p *parser) parseConstSpec() ast.Node {
	name := p.parseName()
	p.expect(scanner.ASSIGN)
	expr := p.parseExpr(precNone)

	spec := &ast.ConstSpec{Name: name, Expr: expr}
	p.symtab = append(p.symtab, Symbol{Type: ConstSym, Name: name, Decl: spec})

	return spec
}

func (p *parser) parseDecl() ast.Node {
	var parse nudFn
	token := p.current
	switch token.Kind {
	case scanner.IMPORT:
		parse = p.parseImportSpec
	case scanner.CONST:
		parse = p.parseConstSpec
	default:
		p.expectMsg("declaration")
		p.sync(declStart)
		return &ast.BadNode{From: token.Pos, To: p.current.Pos}
	}

	// consume keyword
	p.next()
	decl := parse()
	p.expect(scanner.SEMICOLON)

	return decl
}

func (p *parser) parse() *ast.File {
	var nodes []ast.Node
	for p.current.Kind != scanner.ENDMARKER {
		nodes = append(nodes, p.parseDecl())
	}

	return &ast.File{Nodes: nodes}
}

func Parse(text []byte) ParsedFile {
	p := &parser{}
	p.init(text)

	return ParsedFile{
		File:    p.parse(),
		Imports: p.imports,
		Symtab:  p.symtab,
		Lines:   p.scanner.Lines(),
		Errors:  p.errors,
	}
}
