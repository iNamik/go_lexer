go_lexer
========

**Lexer API in Go**


ABOUT
-----

The 'go_lexer' package is an API to help you create hand-written lexers and parsers.

The package was inspired by Rob Pikes' video [Lexical Scanning In Go](http://youtu.be/HxaD_trXwRE) and golang's 'template' package.

LEXER INTERFACE
---------------

Below is the interface for the main Lexer type:

	// lexer.Lexer helps you tokenize bytes
	type Lexer interface {

		// PeekRune allows you to look ahead at runes without consuming them
		PeekRune(int) rune

		// NetRune consumes and returns the next rune in the input
		NextRune() rune

		// BackupRune un-conumes the last rune from the input
		BackupRune()

		// BackupRunes un-consumes the last n runes from the input
		BackupRunes(int)

		// NewLine increments the line number counter, resets the column counter
		NewLine()

		// Line returns the current line number, 1-based
		Line() int

		// Column returns the current column number, 1-based
		Column() int

		// EmitToken emits a token of the specified type, consuming matched runes
		// without emitting them
		EmitToken(TokenType)

		// EmitTokenWithBytes emits a token along with all the consumed runes
		EmitTokenWithBytes(TokenType)

		// IgnoreToken ignores the consumed bytes without emitting any tokens
		IgnoreToken()

		// EmitEOF emits a token of type TokenEOF
		EmitEOF()

		// NextToken retrieves the next emmitted token from the input
		NextToken() *Token

		// Marker returns a marker that you can use to reset the lexer state later
		Marker() *Marker

		// CanReset confirms if the marker is still valid
		CanReset(*Marker) bool

		// Reset resets the lexer state to the specified marker
		Reset(*Marker)

		// MatchZeroOrOneBytes consumes the next rune if it matches, always returning true
		MatchZeroOrOneBytes([]byte) bool

		// MatchZeroOrOneRuness consumes the next rune if it matches, always returning true
		MatchZeroOrOneRunes([]rune) bool

		// MatchZeroOrOneRune consumes the next rune if it matches, always returning true
		MatchZeroOrOneRune(rune) bool

		// MatchZeroOrMoreBytes consumes a run of matching runes, always returning true
		MatchZeroOrMoreBytes([]byte) bool

		// MatchZeroOrMoreRunes consumes a run of matching runes, always returning true
		MatchZeroOrMoreRunes([]rune) bool

		// MatchOneBytes consumes the next rune if its in the list of bytes
		MatchOneBytes([]byte) bool

		// MatchOneRune consumes the next rune if its in the list of bytes
		MatchOneRunes([]rune) bool

		// MatchOneRune consumes the next rune if it matches
		MatchOneRune(rune) bool

		// MatchOneOrMoreBytes consumes a run of matching runes
		MatchOneOrMoreBytes([]byte) bool

		// MatchOneOrMoreRunes consumes a run of matching runes
		MatchOneOrMoreRunes([]rune) bool

		// NonMatchZeroOrOneBytes consumes the next rune if it does not match, always returning true
		NonMatchZeroOrOneBytes([]byte) bool

		// NonMatchZeroOrOneRunes consumes the next rune if it does not match, always returning true
		NonMatchZeroOrOneRunes([]rune) bool

		// NonMatchZeroOrMoreBytes consumes a run of non-matching runes, always returning true
		NonMatchZeroOrMoreBytes([]byte) bool

		// NonMatchZeroOrMoreRunes consumes a run of non-matching runes, always returning true
		NonMatchZeroOrMoreRunes([]rune) bool

		// NonMatchOneBytes consumes the next rune if its NOT in the list of bytes
		NonMatchOneBytes([]byte) bool

		// NonMatchOneRunes consumes the next rune if its NOT in the list of bytes
		NonMatchOneRunes([]rune) bool

		// NonMatchOneOrMoreBytes consumes a run of non-matching runes
		NonMatchOneOrMoreBytes([]byte) bool

		// NonMatchOneOrMoreRunes consumes a run of non-matching runes
		NonMatchOneOrMoreRunes([]rune) bool

		// MatchEOF tries to match the next rune against RuneEOF
		MatchEOF() bool
	}


EXAMPLE
-------

Below is a sample word count program that uses the lexer API:

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


INSTALL
-------

The package is built using the Go tool.  Assuming you have correctly set the
$GOPATH variable, you can run the folloing command:

	go get github.com/iNamik/go_lexer


DEPENDENCIES
------------

go_lexer depends on the iNamik go_container queue package:

* https://github.com/iNamik/go_container


AUTHORS
-------

 * David Farrell
