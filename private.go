package lexer

import (
	"bufio"
	"io"
	"unicode/utf8"
)
import (
	"github.com/iNamik/go_container/queue"
	"github.com/iNamik/go_pkg/bufio/bleeder"
)

const defaultBufSize = 1024 //4096

// lexer holds the state of the scanner.
type lexer struct {
	ioReader   io.Reader     // the original reader passed into New()
	reader     *bufio.Reader // reader buffer
	autoExpand bool          // should we auto-expand buffered reader?
	bufLen     int           // reader buffer len
	line       int           // current line in steram
	column     int           // current column within current line
	peekBytes  []byte        // cache of bufio.Reader.Peek()
	peekPos    int
	tokenLen   int
	runes      queue.Interface // rune buffer
	pos        int
	sequence   int         // Incremented after each emit/ignore - used to validate markers
	state      StateFn     // the next lexing function to enter
	tokens     chan *Token // channel of scanned tokens.
	eofToken   *Token
	eof        bool
}

// newLexer
func newLexer(startState StateFn, reader io.Reader, readerBufLen int, autoExpand bool, channelCap int) Lexer {
	r := bufio.NewReaderSize(reader, readerBufLen)
	l := &lexer{
		ioReader:   reader,
		reader:     r,
		bufLen:     readerBufLen,
		autoExpand: autoExpand,
		runes:      queue.New(4), // 4 is just a nice number that seems appropriate
		state:      startState,
		tokens:     make(chan *Token, channelCap),
		line:       1,
		column:     0,
		eofToken:   nil,
		eof:        false,
	}
	l.updatePeekBytes()
	return l
}

// ensureRuneLen
func (l *lexer) ensureRuneLen(n int) bool {
	for l.runes.Len() < n {
		// If auto-expand is enabled and
		// If our peek buffer is full (suggesting we are likely not at eof) and
		// If we don't have enough bytes left to safely decode a rune,
		if l.autoExpand == true && len(l.peekBytes) == l.bufLen && (len(l.peekBytes)-l.peekPos) < utf8.UTFMax {
			l.bufLen *= 2
			bl := bleeder.New(l.reader, l.ioReader)
			l.reader = bufio.NewReaderSize(bl, l.bufLen)
			l.updatePeekBytes()
		}
		rune, size := utf8.DecodeRune(l.peekBytes[l.peekPos:])
		if utf8.RuneError == rune {
			return false
		}
		l.runes.Add(rune)
		l.peekPos += size
	}

	return l.runes.Len() >= n
}

// emit
func (l *lexer) emit(t TokenType, emitBytes bool) {
	if T_EOF == t {
		if l.eof {
			panic("illegal state: EmitEOF() already called")
		}
		l.consume(false)
		l.eofToken = &Token{typ: T_EOF, bytes: nil, line: l.line, column: l.column + 1}
		l.eof = true
		l.tokens <- l.eofToken
	} else {
		line := l.line

		column := l.column - (l.tokenLen - 1)

		b := l.consume(emitBytes)

		l.tokens <- &Token{typ: t, bytes: b, line: line, column: column}
	}
}

// consume
func (l *lexer) consume(keepBytes bool) []byte {
	var b []byte
	if keepBytes {
		b = make([]byte, l.tokenLen)
		n, err := l.reader.Read(b)
		if err != nil || n != l.tokenLen {
			panic("Unexpected problem in bufio.Reader.Read(): " + err.Error())
		}
	} else {
		// May be better to just grab string and ignore, or always emit bytes
		for ; l.tokenLen > 0; l.tokenLen-- {
			_, err := l.reader.ReadByte()
			if err != nil {
				panic("Unexpected problem in bufio.Reader.ReadByte(): " + err.Error())
			}
		}
		b = nil
	}
	l.sequence++

	l.pos = 0

	l.tokenLen = 0

	l.peekPos = 0

	l.runes.Clear()

	l.updatePeekBytes()

	return b
}

// updatePeekBytes
func (l *lexer) updatePeekBytes() {
	var err error
	l.peekBytes, err = l.reader.Peek(l.bufLen)
	if err != nil && err != bufio.ErrBufferFull && err != io.EOF {
		panic(err)
	}
}
