package main

import "fmt"

// Parser turns tokens into an AST.
type Parser struct {
	tokens []Token
	pos    int
}

// NewParser creates a parser for the given tokens.
func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens}
}

// ParseProgram parses a whole program.
func (p *Parser) ParseProgram() (*Program, error) {
	prog := &Program{}
	p.skipNewlines()
	for !p.atEnd() {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			prog.Statements = append(prog.Statements, stmt)
		}
		p.skipNewlines()
	}
	return prog, nil
}

// ==== Helpers ====

func (p *Parser) peek() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokEOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) peekAt(offset int) Token {
	if p.pos+offset >= len(p.tokens) {
		return Token{Type: TokEOF}
	}
	return p.tokens[p.pos+offset]
}

func (p *Parser) advance() Token {
	t := p.peek()
	p.pos++
	return t
}

func (p *Parser) match(t TokenType) bool {
	if p.peek().Type == t {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) expect(t TokenType) (Token, error) {
	tok := p.peek()
	if tok.Type != t {
		return tok, fmt.Errorf("line %d:%d: expected %s, got %s",
			tok.Line, tok.Col, typeName(t), tok)
	}
	p.advance()
	return tok, nil
}

func (p *Parser) atEnd() bool {
	return p.peek().Type == TokEOF
}

func (p *Parser) skipNewlines() {
	for p.peek().Type == TokNewline || p.peek().Type == TokSemi {
		p.advance()
	}
}

// ==== Statements ====

func (p *Parser) parseStatement() (Node, error) {
	tok := p.peek()

	switch tok.Type {
	case TokLet:
		return p.parseLet()
	case TokFn:
		return p.parseFunc()
	case TokIf:
		return p.parseIf()
	case TokWhile:
		return p.parseWhile()
	case TokReturn:
		return p.parseReturn()
	case TokIdent:
		// Could be assignment or expression
		if p.peekAt(1).Type == TokAssign {
			return p.parseAssign()
		}
		return p.parseExprStatement()
	default:
		return p.parseExprStatement()
	}
}

func (p *Parser) parseLet() (Node, error) {
	p.advance() // 'let'
	nameTok, err := p.expect(TokIdent)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(TokAssign); err != nil {
		return nil, err
	}
	value, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return LetStmt{Name: nameTok.Text, Value: value}, nil
}

func (p *Parser) parseAssign() (Node, error) {
	nameTok := p.advance() // ident
	p.advance()            // '='
	value, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return AssignStmt{Name: nameTok.Text, Value: value}, nil
}

func (p *Parser) parseReturn() (Node, error) {
	p.advance() // 'return'
	// Check if return has a value or if it's bare
	next := p.peek()
	if next.Type == TokNewline || next.Type == TokEOF ||
		next.Type == TokRBrace || next.Type == TokSemi {
		return ReturnStmt{Value: nil}, nil
	}
	value, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return ReturnStmt{Value: value}, nil
}

func (p *Parser) parseIf() (Node, error) {
	p.advance() // 'if'
	cond, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	p.skipNewlines()
	then, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	p.skipNewlines()
	var elseBlock []Node
	if p.match(TokElse) {
		p.skipNewlines()
		if p.peek().Type == TokIf {
			// else if -> wrap in another if statement
			nested, err := p.parseIf()
			if err != nil {
				return nil, err
			}
			elseBlock = []Node{nested}
		} else {
			elseBlock, err = p.parseBlock()
			if err != nil {
				return nil, err
			}
		}
	}
	return IfStmt{Cond: cond, Then: then, Else: elseBlock}, nil
}

func (p *Parser) parseWhile() (Node, error) {
	p.advance() // 'while'
	cond, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	p.skipNewlines()
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return WhileStmt{Cond: cond, Body: body}, nil
}

func (p *Parser) parseFunc() (Node, error) {
	p.advance() // 'fn'
	nameTok, err := p.expect(TokIdent)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(TokLParen); err != nil {
		return nil, err
	}
	var params []string
	for p.peek().Type != TokRParen {
		if len(params) > 0 {
			if _, err := p.expect(TokComma); err != nil {
				return nil, err
			}
		}
		paramTok, err := p.expect(TokIdent)
		if err != nil {
			return nil, err
		}
		params = append(params, paramTok.Text)
	}
	if _, err := p.expect(TokRParen); err != nil {
		return nil, err
	}
	p.skipNewlines()
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return FuncDecl{Name: nameTok.Text, Params: params, Body: body}, nil
}

// parseBlock parses "{" statements "}".
func (p *Parser) parseBlock() ([]Node, error) {
	if _, err := p.expect(TokLBrace); err != nil {
		return nil, err
	}
	p.skipNewlines()
	var stmts []Node
	for p.peek().Type != TokRBrace && !p.atEnd() {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
		p.skipNewlines()
	}
	if _, err := p.expect(TokRBrace); err != nil {
		return nil, err
	}
	return stmts, nil
}

func (p *Parser) parseExprStatement() (Node, error) {
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return ExprStmt{Expr: expr}, nil
}

// ==== Expressions (precedence climbing) ====

func (p *Parser) parseExpr() (Node, error) {
	return p.parseOr()
}

func (p *Parser) parseOr() (Node, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.peek().Type == TokOr {
		op := p.advance().Text
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = Binary{Op: op, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseAnd() (Node, error) {
	left, err := p.parseEquality()
	if err != nil {
		return nil, err
	}
	for p.peek().Type == TokAnd {
		op := p.advance().Text
		right, err := p.parseEquality()
		if err != nil {
			return nil, err
		}
		left = Binary{Op: op, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseEquality() (Node, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}
	for p.peek().Type == TokEq || p.peek().Type == TokNeq {
		op := p.advance().Text
		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		left = Binary{Op: op, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseComparison() (Node, error) {
	left, err := p.parseAddition()
	if err != nil {
		return nil, err
	}
	for p.peek().Type == TokLt || p.peek().Type == TokGt ||
		p.peek().Type == TokLte || p.peek().Type == TokGte {
		op := p.advance().Text
		right, err := p.parseAddition()
		if err != nil {
			return nil, err
		}
		left = Binary{Op: op, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseAddition() (Node, error) {
	left, err := p.parseMultiplication()
	if err != nil {
		return nil, err
	}
	for p.peek().Type == TokPlus || p.peek().Type == TokMinus {
		op := p.advance().Text
		right, err := p.parseMultiplication()
		if err != nil {
			return nil, err
		}
		left = Binary{Op: op, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseMultiplication() (Node, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.peek().Type == TokStar || p.peek().Type == TokSlash || p.peek().Type == TokPercent {
		op := p.advance().Text
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = Binary{Op: op, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseUnary() (Node, error) {
	tok := p.peek()
	if tok.Type == TokMinus || tok.Type == TokNot {
		p.advance()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return Unary{Op: tok.Text, Right: right}, nil
	}
	return p.parseCall()
}

func (p *Parser) parseCall() (Node, error) {
	expr, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	for p.peek().Type == TokLParen {
		p.advance()
		var args []Node
		for p.peek().Type != TokRParen {
			if len(args) > 0 {
				if _, err := p.expect(TokComma); err != nil {
					return nil, err
				}
			}
			arg, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
		if _, err := p.expect(TokRParen); err != nil {
			return nil, err
		}
		expr = Call{Callee: expr, Args: args}
	}
	return expr, nil
}

func (p *Parser) parsePrimary() (Node, error) {
	tok := p.advance()
	switch tok.Type {
	case TokNumber:
		return NumberLit{Value: tok.Text}, nil
	case TokString:
		return StringLit{Value: tok.Text}, nil
	case TokTrue:
		return BoolLit{Value: true}, nil
	case TokFalse:
		return BoolLit{Value: false}, nil
	case TokNil:
		return NilLit{}, nil
	case TokIdent:
		return Ident{Name: tok.Text}, nil
	case TokLParen:
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(TokRParen); err != nil {
			return nil, err
		}
		return expr, nil
	}
	return nil, fmt.Errorf("line %d:%d: unexpected token %s", tok.Line, tok.Col, tok)
}