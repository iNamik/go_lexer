package lexer

import (
	"bufio"
	"io"
)
import (
	"github.com/iNamik/go_container/queue"
)

// TokenType identifies the type of lex tokens.
type TokenType int

// Token represents a token (with optional text string) returned from the scanner.
type Token struct {
	typ    TokenType
	bytes  []byte
	line   int
	column int
}

// Type returns the TokenType of the token
func (t *Token) Type() TokenType { return t.typ }

// Bytes returns the byte array associated with the token, or nil if none
func (t *Token) Bytes() []byte { return t.bytes }

// EOF returns true if the TokenType == TokenTypeEOF
func (t *Token) EOF() bool { return TokenTypeEOF == t.typ }

// Line returns the line number of the token
func (t *Token) Line() int { return t.line }

// Column returns the column number of the token
func (t *Token) Column() int { return t.column }

// TokenType representing EOF
const TokenTypeEOF TokenType = -1

// TokenType representing an unknown rune(s)
const TokenTypeUnknown TokenType = -2

// Rune represending EOF
const RuneEOF = -1

// StateFn represents the state of the scanner as a function that returns the next state.
type StateFn func(Lexer) StateFn

// MatchFn represents a callback function for matching runes that are not
// feasable for a range
type MatchFn func(rune) bool

// Marker stores the state of the lexer to allow rewinding
type Marker struct {
	sequence int
	pos      int
	tokenLen int
	line     int
	column   int
}

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

// New returns a new Lexer object
func New(startState StateFn, reader io.Reader, readerBufLen int, channelCap int) Lexer {
	r := bufio.NewReaderSize(reader, readerBufLen)
	l := &lexer{
		reader:   r,
		bufLen:   readerBufLen,
		runes:    queue.New(4), // 4 is just a nice number that seems appropriate
		state:    startState,
		tokens:   make(chan *Token, channelCap),
		line:     1,
		column:   0,
		eofToken: nil,
		eof:      false,
	}
	l.updatePeekBytes()
	return l
}
