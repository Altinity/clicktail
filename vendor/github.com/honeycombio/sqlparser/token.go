// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sqlparser

import (
	"fmt"
)

const EOFCHAR = 0x100

// Tokenizer is the struct used to generate SQL
// tokens for the parser.
type Tokenizer struct {
	InRunes       []rune
	AllowComments bool
	ForceEOF      bool
	lastChar      rune
	Position      int
	errorToken    []rune
	LastError     string
	posVarIndex   int
	ParseTree     Statement
}

// NewStringTokenizer creates a new Tokenizer for the
// sql string.
func NewStringTokenizer(sql string) *Tokenizer {
	return &Tokenizer{InRunes: []rune(sql)}
}

var keywords = map[string]int{
	"all":           ALL,
	"alter":         ALTER,
	"analyze":       ANALYZE,
	"and":           AND,
	"as":            AS,
	"asc":           ASC,
	"between":       BETWEEN,
	"binary":        BINARY,
	"by":            BY,
	"case":          CASE,
	"create":        CREATE,
	"cross":         CROSS,
	"default":       DEFAULT,
	"delete":        DELETE,
	"desc":          DESC,
	"describe":      DESCRIBE,
	"distinct":      DISTINCT,
	"drop":          DROP,
	"duplicate":     DUPLICATE,
	"else":          ELSE,
	"end":           END,
	"except":        EXCEPT,
	"exists":        EXISTS,
	"explain":       EXPLAIN,
	"for":           FOR,
	"force":         FORCE,
	"from":          FROM,
	"group":         GROUP,
	"having":        HAVING,
	"if":            IF,
	"ignore":        IGNORE,
	"in":            IN,
	"index":         INDEX,
	"inner":         INNER,
	"insert":        INSERT,
	"intersect":     INTERSECT,
	"into":          INTO,
	"is":            IS,
	"join":          JOIN,
	"key":           KEY,
	"left":          LEFT,
	"like":          LIKE,
	"limit":         LIMIT,
	"lock":          LOCK,
	"minus":         MINUS,
	"natural":       NATURAL,
	"not":           NOT,
	"null":          NULL,
	"offset":        OFFSET,
	"on":            ON,
	"or":            OR,
	"order":         ORDER,
	"outer":         OUTER,
	"rename":        RENAME,
	"right":         RIGHT,
	"select":        SELECT,
	"set":           SET,
	"show":          SHOW,
	"straight_join": STRAIGHT_JOIN,
	"table":         TABLE,
	"then":          THEN,
	"to":            TO,
	"union":         UNION,
	"unique":        UNIQUE,
	"update":        UPDATE,
	"use":           USE,
	"using":         USING,
	"values":        VALUES,
	"view":          VIEW,
	"when":          WHEN,
	"where":         WHERE,

	//keywords for creat table

	"engine": ENGINE,

	//datatypes
	"bit":       BIT,
	"tinyint":   TINYINT,
	"smallint":  SMALLINT,
	"mediumint": MEDIUMINT,
	"int":       INT,
	"integer":   INTEGER,
	"bigint":    BIGINT,
	"real":      REAL,
	"double":    DOUBLE,
	"float":     FLOAT,
	"decimal":   DECIMAL,
	"numeric":   NUMERIC,

	"char":       CHAR,
	"varchar":    VARCHAR,
	"text":       TEXT,
	"mediumtext": MEDIUMTEXT,
	"charset":    CHARSET,

	"date":      DATE,
	"time":      TIME,
	"timestamp": TIMESTAMP,
	"datetime":  DATETIME,
	"year":      YEAR,

	//other keywords
	"unsigned":       UNSIGNED,
	"zerofill":       ZEROFILL,
	"primary":        PRIMARY,
	"auto_increment": AUTO_INCREMENT,
}

// Lex returns the next token form the Tokenizer.
// This function is used by go yacc.
func (tkn *Tokenizer) Lex(lval *yySymType) int {
	typ, val := tkn.Scan()
	for typ == COMMENT {
		if tkn.AllowComments {
			break
		}
		typ, val = tkn.Scan()
	}
	switch typ {
	case ID, STRING, NUMBER, VALUE_ARG, LIST_ARG, COMMENT:
		lval.runes = val
	}
	tkn.errorToken = val
	return typ
}

// Error is called by go yacc if there's a parsing error.
func (tkn *Tokenizer) Error(err string) {
	if tkn.errorToken != nil {
		tkn.LastError = fmt.Sprintf("%s at position %v near %s", err, tkn.Position, string(tkn.errorToken))
	} else {
		tkn.LastError = fmt.Sprintf("%s at position %v", err, tkn.Position)
	}
}

// Scan scans the tokenizer for the next token and returns
// the token type and an optional value.
func (tkn *Tokenizer) Scan() (int, []rune) {
	if tkn.ForceEOF {
		return 0, nil
	}

	if tkn.lastChar == 0 {
		tkn.next()
	}
	tkn.skipBlank()
	switch ch := tkn.lastChar; {
	case isLetter(ch):
		return tkn.scanIdentifier()
	case isDigit(ch):
		return tkn.scanNumber(false)
	case ch == ':':
		return tkn.scanBindVar()
	default:
		tkn.next()
		switch ch {
		case EOFCHAR:
			return 0, nil
		case '|':
			if tkn.lastChar == '|' {
				tkn.next()
				return OR, nil
			}
			return int(ch), nil
		case '&':
			if tkn.lastChar == '&' {
				tkn.next()
				return AND, nil
			}
			return int(ch), nil
		case '=', ',', ';', '(', ')', '+', '*', '%', '^', '~':
			return int(ch), nil
		case '?':
			tkn.posVarIndex++
			rv := fmt.Sprintf(":v%d", tkn.posVarIndex)
			return VALUE_ARG, []rune(rv)
		case '.':
			if isDigit(tkn.lastChar) {
				return tkn.scanNumber(true)
			}
			return int(ch), nil
		case '/':
			switch tkn.lastChar {
			case '/':
				tkn.next()
				return tkn.scanCommentType1("//")
			case '*':
				tkn.next()
				return tkn.scanCommentType2()
			default:
				return int(ch), nil
			}
		case '-':
			if tkn.lastChar == '-' {
				tkn.next()
				return tkn.scanCommentType1("--")
			}
			return int(ch), nil
		case '<':
			switch tkn.lastChar {
			case '>':
				tkn.next()
				return NE, nil
			case '=':
				tkn.next()
				switch tkn.lastChar {
				case '>':
					tkn.next()
					return NULL_SAFE_EQUAL, nil
				default:
					return LE, nil
				}
			default:
				return int(ch), nil
			}
		case '>':
			if tkn.lastChar == '=' {
				tkn.next()
				return GE, nil
			}
			return int(ch), nil
		case '!':
			if tkn.lastChar == '=' {
				tkn.next()
				return NE, nil
			}
			return LEX_ERROR, []rune("!")
		case '\'', '"':
			return tkn.scanString(ch, STRING)
		case '`':
			return tkn.scanLiteralIdentifier()
		default:
			return LEX_ERROR, []rune{ch}
		}
	}
}

func (tkn *Tokenizer) skipBlank() {
	ch := tkn.lastChar
	for ch == ' ' || ch == '\n' || ch == '\r' || ch == '\t' {
		tkn.next()
		ch = tkn.lastChar
	}
}

func (tkn *Tokenizer) scanIdentifier() (int, []rune) {
	startPos := tkn.Position - 1
	for tkn.next(); isLetter(tkn.lastChar) || isDigit(tkn.lastChar); tkn.next() {
	}
	scanned := tkn.InRunes[startPos : tkn.Position-1]
	if keywordID, found := keywords[string(scanned)]; found {
		return keywordID, scanned
	}
	return ID, scanned
}

func (tkn *Tokenizer) scanLiteralIdentifier() (int, []rune) {
	startPos := tkn.Position - 2
	if !isLetter(tkn.lastChar) {
		return LEX_ERROR, tkn.InRunes[startPos : startPos+1]
	}
	for tkn.next(); isLetter(tkn.lastChar) || isDigit(tkn.lastChar); tkn.next() {
	}
	if tkn.lastChar != '`' {
		return LEX_ERROR, tkn.InRunes[startPos : tkn.Position-1]
	}
	tkn.next()
	return ID, tkn.InRunes[startPos : tkn.Position-1]
}

func (tkn *Tokenizer) scanBindVar() (int, []rune) {
	startPos := tkn.Position
	token := VALUE_ARG
	tkn.next()
	if tkn.lastChar == ':' {
		token = LIST_ARG
		tkn.next()
	}
	if !isLetter(tkn.lastChar) {
		return LEX_ERROR, tkn.InRunes[startPos : tkn.Position-1]
	}
	for isLetter(tkn.lastChar) || isDigit(tkn.lastChar) || tkn.lastChar == '.' {
		tkn.next()
	}
	return token, tkn.InRunes[startPos : tkn.Position-1]
}

func (tkn *Tokenizer) scanMantissa(base int) {
	for digitVal(tkn.lastChar) < base {
		tkn.next()
	}
}

func (tkn *Tokenizer) scanNumber(seenDecimalPoint bool) (int, []rune) {
	startPos := tkn.Position - 1
	if seenDecimalPoint {
		tkn.scanMantissa(10)
		goto exponent
	}

	if tkn.lastChar == '0' {
		// int or float
		tkn.next()
		if tkn.lastChar == 'x' || tkn.lastChar == 'X' {
			// hexadecimal int
			tkn.next()
			tkn.scanMantissa(16)
		} else {
			// octal int or float
			seenDecimalDigit := false
			tkn.scanMantissa(8)
			if tkn.lastChar == '8' || tkn.lastChar == '9' {
				// illegal octal int or float
				seenDecimalDigit = true
				tkn.scanMantissa(10)
			}
			if tkn.lastChar == '.' || tkn.lastChar == 'e' || tkn.lastChar == 'E' {
				goto fraction
			}
			// octal int
			if seenDecimalDigit {
				return LEX_ERROR, tkn.InRunes[startPos : tkn.Position-1]
			}
		}
		goto exit
	}

	// decimal int or float
	tkn.scanMantissa(10)

fraction:
	if tkn.lastChar == '.' {
		tkn.next()
		tkn.scanMantissa(10)
	}

exponent:
	if tkn.lastChar == 'e' || tkn.lastChar == 'E' {
		tkn.next()
		if tkn.lastChar == '+' || tkn.lastChar == '-' {
			tkn.next()
		}
		tkn.scanMantissa(10)
	}

exit:
	return NUMBER, tkn.InRunes[startPos : tkn.Position-1]
}

func (tkn *Tokenizer) scanString(delim rune, typ int) (int, []rune) {
	startPos := tkn.Position
	for {
		ch := tkn.lastChar
		tkn.next()
		if ch == delim {
			if tkn.lastChar == delim {
				tkn.next()
			} else {
				break
			}
		} else if ch == '\\' {
			if tkn.lastChar == EOFCHAR {
				return LEX_ERROR, tkn.InRunes[startPos : tkn.Position-1]
			}
			tkn.next()
		}
		if ch == EOFCHAR {
			return LEX_ERROR, tkn.InRunes[startPos : tkn.Position-1]
		}
	}
	return typ, tkn.InRunes[startPos : tkn.Position-1]
}

func (tkn *Tokenizer) scanCommentType1(prefix string) (int, []rune) {
	startPos := tkn.Position
	for tkn.lastChar != EOFCHAR {
		if tkn.lastChar == '\n' {
			tkn.next()
			break
		}
		tkn.next()
	}
	return COMMENT, tkn.InRunes[startPos : tkn.Position-1]
}

func (tkn *Tokenizer) scanCommentType2() (int, []rune) {
	startPos := tkn.Position
	for {
		if tkn.lastChar == '*' {
			tkn.next()
			if tkn.lastChar == '/' {
				tkn.next()
				break
			}
			continue
		}
		if tkn.lastChar == EOFCHAR {
			return LEX_ERROR, tkn.InRunes[startPos : tkn.Position-1]
		}
		tkn.next()
	}
	return COMMENT, tkn.InRunes[startPos : tkn.Position-3]
}

func (tkn *Tokenizer) next() {
	if tkn.Position >= len(tkn.InRunes) {
		// Only EOF is possible.
		tkn.lastChar = EOFCHAR
		if tkn.Position == len(tkn.InRunes) {
			tkn.Position++
		}
		return
	}
	tkn.lastChar = tkn.InRunes[tkn.Position]
	tkn.Position++
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch == '@'
}

func digitVal(ch rune) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch) - '0'
	case 'a' <= ch && ch <= 'f':
		return int(ch) - 'a' + 10
	case 'A' <= ch && ch <= 'F':
		return int(ch) - 'A' + 10
	}
	return 16 // larger than any legal digit val
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}
