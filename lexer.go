package lexer

import (
	"io"
	"bufio"
	"github.com/iNamik/go_container/queue"
)

// TokenType identifies the type of lex tokens.
type TokenType int

// Token represents a token (with optional text string) returned from the scanner.
type Token struct {
	typ     TokenType
	bytes []byte
	line    int
	column  int
}
// Type returns the TokenType of the token
func (t *Token) Type ()   TokenType { return t.typ   }
// Bytes returns the byte array associated with the token, or nil if none
func (t *Token) Bytes() []byte      { return t.bytes }
// EOF returns true if the TokenType == TokenTypeEOF
func (t *Token) EOF()     bool      { return TokenTypeEOF == t.typ }
// Line returns the line number of the token
func (t *Token) Line()    int       { return t.line   }
// Column returns the column number of the token
func (t *Token) Column()  int       { return t.column }

// TokenType representing EOF
const TokenTypeEOF TokenType = -1

// Rune represending EOF
const RuneEOF = -1

// StateFn represents the state of the scanner as a function that returns the next state.
type StateFn func(Lexer) StateFn

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

	// MatchOne consumes the next rune if its in the list of bytes
	MatchOne       ([]byte) bool

	// MatchOneOrMore consumes a run of matching runes
	MatchOneOrMore ([]byte) bool

	// MatchNoneOrOne consumes the next rune if it matches, always returning true
	MatchNoneOrOne ([]byte) bool

	// MatchNoneOrMore consumes a run of matching runes, always returning true
	MatchNoneOrMore([]byte) bool

	// MatchEOF tries to match the next rune against RuneEOF
	MatchEOF()              bool

	// NonMatchOne consumes the next rune if its NOT in the list of bytes
	NonMatchOne       ([]byte) bool

	// NonMatchOneOrMore consumes a run of non-matching runes
	NonMatchOneOrMore ([]byte) bool

	// NonMatchNoneOrOne consumes the next rune if it does not match, always returning true
	NonMatchNoneOrOne ([]byte) bool

	// NonMatchNoneOrMore consumes a run of non-matching runes, always returning true
	NonMatchNoneOrMore([]byte) bool

	// PeekRune allows you to look ahead at runes without consuming them
	PeekRune (int) rune

	// NetRune consumes and returns the next rune in the input
	NextRune ()    rune

	// BackupRune un-conumes the last rune from the input
	BackupRune ()

	// BackupRunes un-consumes the last n runes from the input
	BackupRunes(int)

	// NewLine increments the line number counter, resets the column counter
	NewLine()

	// Line returns the current line number, 1-based
	Line()    int

	// Column returns the current column number, 1-based
	Column()  int

	// EmitToken emits a token of the specified type, consuming matched runes
	// without emitting them
	EmitToken         (TokenType)

	// EmitTokenWithBytes emits a token along with all the consumed runes
	EmitTokenWithBytes(TokenType)

	// IgnoreToken ignores the consumed bytes without emitting any tokens
	IgnoreToken()

	// EmitEOF emits a token of type TokenEOF
	EmitEOF()

	// NextToken retrieves the next emmitted token from the input
	NextToken  () *Token

	// Marker returns a marker that you can use to reset the lexer state later
	Marker() *Marker

	// Reset resets the lexer state to the specified marker
	Reset(*Marker)
}

// NewLexer returns a new Lexer object
func NewLexer(startState StateFn, reader io.Reader, readerBufLen int, channelCap int) Lexer {
	r := bufio.NewReaderSize(reader, readerBufLen)
	l := &lexer{
		reader   : r,
		bufLen   : readerBufLen,
		runes    : queue.NewQueue(4),
		state    : startState,
		tokens   : make(chan *Token, channelCap),
		line     : 1,
		column   : 0,
		eofToken : nil,
		eof      : false,
	}
	l.updatePeekBytes()
	return l
}

// RangeToBytes converts a range specifier into a byte array suitable for the Match* calls
func RangeToBytes(r string) []byte {
	b := make([]byte, 0)

	for i, l := 0, len(r) ; i < l ; {
		var left, right byte
		left = r[i];
		i++
		if i < l && r[i] == '-' {
			i++
			if (i < l) {
				right = r[i]
				if left <= right {
					for j := left; j != right; j++ {
						b = append(b, j)
					}
				} else {
					panic("error in range spec - range not low-to-high")
				}
			} else {
				panic("error in range spec - trailing '-'")
			}
		} else {
			b = append(b, left)
		}
	}

	return b;
}

