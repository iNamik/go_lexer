package main

import "os"
import "fmt"
import "github.com/iNamik/go_lexer"

// Usage : wordcount <filename>
func usage() {
	fmt.Printf("usage: %s <filename>\n", os.Args[0])
}

// We define our lexer tokens starting from the pre-defined EOF token
const (
	T_EOF   lexer.TokenType = lexer.TokenTypeEOF
	T_SPACE                 = lexer.TokenTypeEOF + iota
	T_NEWLINE
	T_WORD
)

// List gleaned from isspace(3) manpage
var bytesNonWord = []byte{' ', '\t', '\f', '\v', '\n', '\r'}

var bytesSpace = []byte{' ', '\t', '\f', '\v'}

const charNewLine = '\n'

const charReturn = '\r'

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	var file *os.File
	var error error

	file, error = os.Open(os.Args[1])

	if error != nil {
		panic(error)
	}

	var chars int = 0

	var words int = 0

	var spaces int = 0

	var lines int = 0

	// To help us track last line
	var emptyLine bool = true

	// Create our lexer
	// New(startState, reader, readerBufLen, channelCap)
	lex := lexer.New(lexFunc, file, 100, 1)

	// Process lexer-emitted tokens
	for t := lex.NextToken(); lexer.TokenTypeEOF != t.Type(); t = lex.NextToken() {

		chars += len(t.Bytes())

		switch t.Type() {
		case T_WORD:
			words++
			emptyLine = false

		case T_NEWLINE:
			lines++
			spaces++
			emptyLine = true

		case T_SPACE:
			spaces += len(t.Bytes())
			emptyLine = false

		default:
			panic("unreachable")
		}
	}

	// If last line not empty, up line count
	if !emptyLine {
		lines++
	}

	fmt.Printf("%d words, %d spaces, %d lines, %d chars\n", words, spaces, lines, chars)
}

func lexFunc(l lexer.Lexer) lexer.StateFn {
	// EOF
	if l.MatchEOF() {
		l.EmitEOF()
		return nil // We're done here
	}

	// Non-Space run
	if l.NonMatchOneOrMoreBytes(bytesNonWord) {
		l.EmitTokenWithBytes(T_WORD)

		// Space run
	} else if l.MatchOneOrMoreBytes(bytesSpace) {
		l.EmitTokenWithBytes(T_SPACE)

		// Line Feed
	} else if charNewLine == l.PeekRune(0) {
		l.NextRune()
		l.EmitTokenWithBytes(T_NEWLINE)
		l.NewLine()

		// Carriage-Return with optional line-feed immediately following
	} else if charReturn == l.PeekRune(0) {
		l.NextRune()
		if charNewLine == l.PeekRune(0) {
			l.NextRune()
		}
		l.EmitTokenWithBytes(T_NEWLINE)
		l.NewLine()
	} else {
		panic("unreachable")
	}

	return lexFunc
}
