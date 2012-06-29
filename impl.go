package lexer

import (
	"bytes"
	"unicode/utf8"
)

/**
 * Lexer::NextToken - Returns the next token from the reader.
 */
func (l *lexer) NextToken() *Token {
	for {
		select {
			case token := <- l.tokens:
				return token
			default:
				l.state = l.state(l)
		}
	}
	panic("not reached")
}

/**
 * Lexer::NewLine
 */
func (l *lexer) NewLine() {
	l.line++
	l.column = 0
}

/**
 * Lexer::Line
 */
func (l *lexer) Line() int {
	return l.line
}

/**
 * Lexer::Column
 */
func (l *lexer) Column() int {
	return l.column
}

/**
 * Lexer:PeekRune
 */
func (l *lexer) PeekRune(n int) rune {
	ok := l.ensureRuneLen( l.pos + n + 1 ) // Correct for 0-based 'n'

	if !ok {
		return RuneEOF
	}

	i := l.runes.Peek( l.pos + n )

	return i.(rune)
}

/**
 * Lexer::NextRune
 */
func (l *lexer) NextRune() rune {
	ok := l.ensureRuneLen( l.pos + 1 )

	if !ok {
		return RuneEOF
	}

	i := l.runes.Peek( l.pos ) // 0-based

	r := i.(rune)
	l.pos++
	l.tokenLen += utf8.RuneLen(r)
	l.column   += utf8.RuneLen(r)

	return r
}

/**
 * Lexer::BackupRune
 */
func (l *lexer) BackupRune() {
	l.BackupRunes(1)
}

/**
 * Lexer::BackupRunes
 */
func (l *lexer) BackupRunes(n int) {
	for  ; n > 0 ; n-- {
		if l.pos > 0 {
			l.pos--
			i := l.runes.Peek( l.pos ) // 0-based
			r := i.(rune)
			l.tokenLen -= utf8.RuneLen(r)
			l.column   -= utf8.RuneLen(r)
		} else {
			panic("Underflow Exception")
		}
	}
}

/**
 * Lexer::EmitToken
 */
func (l *lexer) EmitToken(t TokenType) {
	l.emit(t, false)
}

/**
 * Lexer::EmitTokenWithBytes
 */
func (l *lexer) EmitTokenWithBytes(t TokenType) {
	l.emit(t, true)
}

/**
 * Lexer::EmitToken
 */
func (l *lexer) EmitEOF() {
	l.emit(TokenTypeEOF, false)
}

/**
 * Lexer::IgnoreToken
 */
func (l *lexer) IgnoreToken() {
	l.consume(false)
}


/**
 * Lexer::Marker
 */
func (l *lexer) Marker() *Marker {
	return &Marker{sequence: l.sequence, pos: l.pos, tokenLen: l.tokenLen, line: l.line, column: l.column}
}

/**
 * Lexer::Reset
 */
func (l *lexer) Reset(m *Marker) {
	if (m.sequence != l.sequence || m.pos > l.runes.Len() || m.tokenLen > l.peekPos) {
		panic("Invalid marker")
	}
	l.pos      = m.pos
	l.tokenLen = m.tokenLen
	l.line     = m.line
	l.column   = m.column
}

/**
 * Lexer::MatchOne
 */
func (l *lexer) MatchOne(match []byte) bool {
	if r := l.PeekRune(0) ; r != RuneEOF && bytes.IndexRune( match, r ) >= 0 {
		l.NextRune()
		return true
	}
	return false
}

/**
 * Lexer::MatchOneOrMore
 */
func (l *lexer) MatchOneOrMore(match []byte) bool {
	var r rune
	if r = l.PeekRune(0) ; r != RuneEOF && bytes.IndexRune( match, r ) >= 0 {
		l.NextRune()
		for r = l.PeekRune(0) ; r != RuneEOF && bytes.IndexRune( match, r ) >= 0 ; r = l.PeekRune(0) {
			l.NextRune()
		}
		return true
	}
	return false
}

/**
 * Lexer::MatchNoneOrOne
 */
func (l *lexer) MatchNoneOrOne(match []byte) bool {
	if r := l.PeekRune(0) ; r != RuneEOF && bytes.IndexRune( match, r ) >= 0 {
		l.NextRune()
	}
	return true
}

/**
 * Lexer::MatchNoneOrMore
 */
func (l *lexer) MatchNoneOrMore(match []byte) bool {
	for r := l.PeekRune(0) ; r != RuneEOF && bytes.IndexRune( match, r ) >= 0 ; r = l.PeekRune(0) {
		l.NextRune()
	}
	return true
}

/**
 * Lexer::MatchEOF
 */
func (l *lexer) MatchEOF() bool {
	if r := l.PeekRune(0) ; r == RuneEOF {
		l.NextRune()
		return true
	}
	return false
}

/**
 * Lexer::NonMatchOne
 */
func (l *lexer) NonMatchOne(match []byte) bool {
	if r := l.PeekRune(0) ; r != RuneEOF && bytes.IndexRune( match, r ) == -1 {
		l.NextRune()
		return true
	}
	return false
}

/**
 * Lexer::NonMatchOneOrMore
 */
func (l *lexer) NonMatchOneOrMore(match []byte) bool {
	var r rune
	if r = l.PeekRune(0) ; r != RuneEOF && bytes.IndexRune( match, r ) == -1 {
		l.NextRune()
		for r = l.PeekRune(0) ; r != RuneEOF && bytes.IndexRune( match, r ) == -1 ; r = l.PeekRune(0) {
			l.NextRune()
		}
		return true
	}
	return false
}

/**
 * Lexer::NonMatchNoneOrOne
 */
func (l *lexer) NonMatchNoneOrOne(match []byte) bool {
	if r := l.PeekRune(0) ; r != RuneEOF && bytes.IndexRune( match, r ) == -1 {
		l.NextRune()
	}
	return true
}

/**
 * Lexer::NonMatchNoneOrMore
 */
func (l *lexer) NonMatchNoneOrMore(match []byte) bool {
	for r := l.PeekRune(0) ; r != RuneEOF && bytes.IndexRune( match, r ) == -1 ; r = l.PeekRune(0) {
		l.NextRune()
	}
	return true
}
