package main

import "fmt"

// TokenType identifies the kind of a token.
type TokenType int

const (
	TokEOF TokenType = iota

	// Literals
	TokNumber
	TokString
	TokIdent

	// Keywords
	TokLet
	TokFn
	TokIf
	TokElse
	TokWhile
	TokReturn
	TokAnd
	TokOr
	TokNot
	TokTrue
	TokFalse
	TokNil

	// Punctuation
	TokLParen    // (
	TokRParen    // )
	TokLBrace    // {
	TokRBrace    // }
	TokComma     // ,
	TokSemi      // ;
	TokAssign    // =
	TokPlus      // +
	TokMinus     // -
	TokStar      // *
	TokSlash     // /
	TokPercent   // %
	TokEq        // ==
	TokNeq       // !=
	TokLt        // <
	TokGt        // >
	TokLte       // <=
	TokGte       // >=
	TokArrow     // ->
	TokNewline   // significant for statement separation
)

// Token is one lexeme.
type Token struct {
	Type TokenType
	Text string
	Line int
	Col  int
}

func (t Token) String() string {
	if t.Text != "" {
		return fmt.Sprintf("%s(%q)", typeName(t.Type), t.Text)
	}
	return typeName(t.Type)
}

func typeName(t TokenType) string {
	switch t {
	case TokEOF:
		return "EOF"
	case TokNumber:
		return "NUMBER"
	case TokString:
		return "STRING"
	case TokIdent:
		return "IDENT"
	case TokLet:
		return "let"
	case TokFn:
		return "fn"
	case TokIf:
		return "if"
	case TokElse:
		return "else"
	case TokWhile:
		return "while"
	case TokReturn:
		return "return"
	case TokAnd:
		return "and"
	case TokOr:
		return "or"
	case TokNot:
		return "not"
	case TokTrue:
		return "true"
	case TokFalse:
		return "false"
	case TokNil:
		return "nil"
	case TokLParen:
		return "("
	case TokRParen:
		return ")"
	case TokLBrace:
		return "{"
	case TokRBrace:
		return "}"
	case TokComma:
		return ","
	case TokSemi:
		return ";"
	case TokAssign:
		return "="
	case TokPlus:
		return "+"
	case TokMinus:
		return "-"
	case TokStar:
		return "*"
	case TokSlash:
		return "/"
	case TokPercent:
		return "%"
	case TokEq:
		return "=="
	case TokNeq:
		return "!="
	case TokLt:
		return "<"
	case TokGt:
		return ">"
	case TokLte:
		return "<="
	case TokGte:
		return ">="
	case TokArrow:
		return "->"
	case TokNewline:
		return "NEWLINE"
	}
	return "UNKNOWN"
}

// keywords maps identifiers to their keyword token types.
var keywords = map[string]TokenType{
	"let":    TokLet,
	"fn":     TokFn,
	"if":     TokIf,
	"else":   TokElse,
	"while":  TokWhile,
	"return": TokReturn,
	"and":    TokAnd,
	"or":     TokOr,
	"not":    TokNot,
	"true":   TokTrue,
	"false":  TokFalse,
	"nil":    TokNil,
}