package main

import (
	"fmt"
	"strings"
)

// Lexer turns source code into tokens.
type Lexer struct {
	source string
	pos    int
	line   int
	col    int
}

// NewLexer creates a lexer for the given source.
func NewLexer(source string) *Lexer {
	// Strip UTF-8 BOM if present
	if strings.HasPrefix(source, "\ufeff") {
		source = source[3:]
	}
	return &Lexer{source: source, pos: 0, line: 1, col: 1}
}

// Tokenize returns all tokens (ending with EOF).
func (l *Lexer) Tokenize() ([]Token, error) {
	var tokens []Token
	for {
		tok, err := l.next()
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, tok)
		if tok.Type == TokEOF {
			break
		}
	}
	return tokens, nil
}

func (l *Lexer) peek() byte {
	if l.pos >= len(l.source) {
		return 0
	}
	return l.source[l.pos]
}

func (l *Lexer) peekAt(offset int) byte {
	if l.pos+offset >= len(l.source) {
		return 0
	}
	return l.source[l.pos+offset]
}

func (l *Lexer) advance() byte {
	c := l.source[l.pos]
	l.pos++
	if c == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return c
}

func (l *Lexer) makeToken(t TokenType, text string, line, col int) Token {
	return Token{Type: t, Text: text, Line: line, Col: col}
}

func (l *Lexer) next() (Token, error) {
	// Skip spaces and tabs (but not newlines — they're significant)
	for l.pos < len(l.source) {
		c := l.peek()
		if c == ' ' || c == '\t' || c == '\r' {
			l.advance()
		} else if c == '#' {
			// Comment until end of line
			for l.pos < len(l.source) && l.peek() != '\n' {
				l.advance()
			}
		} else {
			break
		}
	}

	if l.pos >= len(l.source) {
		return l.makeToken(TokEOF, "", l.line, l.col), nil
	}

	startLine := l.line
	startCol := l.col
	c := l.peek()

	// Newline
	if c == '\n' {
		l.advance()
		return l.makeToken(TokNewline, "\n", startLine, startCol), nil
	}

	// Numbers
	if c >= '0' && c <= '9' {
		var sb strings.Builder
		for l.pos < len(l.source) && (isDigit(l.peek()) || l.peek() == '.') {
			sb.WriteByte(l.advance())
		}
		return l.makeToken(TokNumber, sb.String(), startLine, startCol), nil
	}

	// Identifiers and keywords
	if isAlpha(c) || c == '_' {
		var sb strings.Builder
		for l.pos < len(l.source) && (isAlphaNumeric(l.peek()) || l.peek() == '_') {
			sb.WriteByte(l.advance())
		}
		text := sb.String()
		if kw, ok := keywords[text]; ok {
			return l.makeToken(kw, text, startLine, startCol), nil
		}
		return l.makeToken(TokIdent, text, startLine, startCol), nil
	}

	// Strings
	if c == '"' {
		l.advance()
		var sb strings.Builder
		for l.pos < len(l.source) && l.peek() != '"' {
			ch := l.advance()
			if ch == '\\' {
				if l.pos >= len(l.source) {
					return Token{}, fmt.Errorf("line %d:%d: unterminated string", startLine, startCol)
				}
				esc := l.advance()
				switch esc {
				case 'n':
					sb.WriteByte('\n')
				case 't':
					sb.WriteByte('\t')
				case '\\':
					sb.WriteByte('\\')
				case '"':
					sb.WriteByte('"')
				default:
					sb.WriteByte(esc)
				}
			} else {
				sb.WriteByte(ch)
			}
		}
		if l.pos >= len(l.source) {
			return Token{}, fmt.Errorf("line %d:%d: unterminated string", startLine, startCol)
		}
		l.advance()
		return l.makeToken(TokString, sb.String(), startLine, startCol), nil
	}

	// Two-character operators
	two := ""
	if l.pos+1 < len(l.source) {
		two = l.source[l.pos : l.pos+2]
	}
	switch two {
	case "==":
		l.advance()
		l.advance()
		return l.makeToken(TokEq, "==", startLine, startCol), nil
	case "!=":
		l.advance()
		l.advance()
		return l.makeToken(TokNeq, "!=", startLine, startCol), nil
	case "<=":
		l.advance()
		l.advance()
		return l.makeToken(TokLte, "<=", startLine, startCol), nil
	case ">=":
		l.advance()
		l.advance()
		return l.makeToken(TokGte, ">=", startLine, startCol), nil
	case "->":
		l.advance()
		l.advance()
		return l.makeToken(TokArrow, "->", startLine, startCol), nil
	}

	l.advance()
	switch c {
	case '(':
		return l.makeToken(TokLParen, "(", startLine, startCol), nil
	case ')':
		return l.makeToken(TokRParen, ")", startLine, startCol), nil
	case '{':
		return l.makeToken(TokLBrace, "{", startLine, startCol), nil
	case '}':
		return l.makeToken(TokRBrace, "}", startLine, startCol), nil
	case ',':
		return l.makeToken(TokComma, ",", startLine, startCol), nil
	case ';':
		return l.makeToken(TokSemi, ";", startLine, startCol), nil
	case '=':
		return l.makeToken(TokAssign, "=", startLine, startCol), nil
	case '+':
		return l.makeToken(TokPlus, "+", startLine, startCol), nil
	case '-':
		return l.makeToken(TokMinus, "-", startLine, startCol), nil
	case '*':
		return l.makeToken(TokStar, "*", startLine, startCol), nil
	case '/':
		return l.makeToken(TokSlash, "/", startLine, startCol), nil
	case '%':
		return l.makeToken(TokPercent, "%", startLine, startCol), nil
	case '<':
		return l.makeToken(TokLt, "<", startLine, startCol), nil
	case '>':
		return l.makeToken(TokGt, ">", startLine, startCol), nil
	}

	return Token{}, fmt.Errorf("line %d:%d: unexpected character %q", startLine, startCol, c)
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isAlphaNumeric(c byte) bool {
	return isAlpha(c) || isDigit(c)
}