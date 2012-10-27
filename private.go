package lexer

import (
	"bufio"
	"io"
	"unicode/utf8"
)
import (
	"github.com/iNamik/go_container/queue"
)

// lexer holds the state of the scanner.
type lexer struct {
	reader      *bufio.Reader // reader buffer
	bufLen      int           // reader buffer len
	startBufLen int           // reader buffer len at start
	line        int           // current line in steram
	column      int           // current column within current line
	peekBytes   []byte        // cache of bufio.Reader.Peek()
	peekPos     int
	tokenLen    int
	runes       queue.Interface // rune buffer
	pos         int
	sequence    int         // Incremented after each emit/ignore - used to validate markers
	state       StateFn     // the next lexing function to enter
	tokens      chan *Token // channel of scanned tokens.
	eofToken    *Token
	eof         bool
}

// ensureRuneLen
func (l *lexer) ensureRuneLen(n int) bool {
	for l.runes.Len() < n {
		// check if the buffer is empty
		if l.peekPos >= l.bufLen {
			l.increaseBufferSize()
			continue
		}

		rune, size := utf8.DecodeRune(l.peekBytes[l.peekPos:])

		// check if rune decode failed, try to replace with larger buffer if so
		if rune == utf8.RuneError && l.peekPos+size >= l.bufLen {
			l.increaseBufferSize()
			continue
		}

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
	if TokenTypeEOF == t {
		if l.eof {
			panic("illegal state: EmitEOF() already called")
		}
		l.consume(false)
		l.eofToken = &Token{typ: TokenTypeEOF, bytes: nil, line: l.line, column: l.column + 1}
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

// replace current reader with a new reader with buffer size increased by startBufLen
func (l *lexer) increaseBufferSize() {
	l.bufLen += l.startBufLen
	l.reader = bufio.NewReaderSize(l.reader, l.bufLen)
	l.updatePeekBytes()
}

// updatePeekBytes
func (l *lexer) updatePeekBytes() {
	var err error
	l.peekBytes, err = l.reader.Peek(l.bufLen)
	if err != nil && err != bufio.ErrBufferFull && err != io.EOF {
		panic(err)
	}
}

// runesContainRune
func runesContainRune(runes []rune, r rune) bool {
	for i, l := 0, len(runes); i < l; i++ {
		if r == runes[i] {
			return true
		}
	}
	return false
}
