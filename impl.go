package lexer

import (
	"bytes"
	"unicode/utf8"
)

import (
	"github.com/iNamik/go_pkg/runes"
)

// Lexer::NextToken - Returns the next token from the reader.
func (l *lexer) NextToken() *Token {
	for {
		select {
		case token := <-l.tokens:
			return token
		default:
			l.state = l.state(l)
		}
	}
	panic("not reached")
}

// Lexer::NewLine
func (l *lexer) NewLine() {
	l.line++
	l.column = 0
}

// Lexer::Line
func (l *lexer) Line() int {
	return l.line
}

// Lexer::Column
func (l *lexer) Column() int {
	return l.column
}

// Lexer:PeekRune
func (l *lexer) PeekRune(n int) rune {
	ok := l.ensureRuneLen(l.pos + n + 1) // Correct for 0-based 'n'

	if !ok {
		return RuneEOF
	}

	i := l.runes.Peek(l.pos + n)

	return i.(rune)
}

// Lexer::NextRune
func (l *lexer) NextRune() rune {
	ok := l.ensureRuneLen(l.pos + 1)

	if !ok {
		return RuneEOF
	}

	i := l.runes.Peek(l.pos) // 0-based

	r := i.(rune)

	l.pos++

	l.tokenLen += utf8.RuneLen(r)

	l.column += utf8.RuneLen(r)

	return r
}

// Lexer::BackupRune
func (l *lexer) BackupRune() {
	l.BackupRunes(1)
}

// Lexer::BackupRunes
func (l *lexer) BackupRunes(n int) {
	for ; n > 0; n-- {
		if l.pos > 0 {
			l.pos--

			i := l.runes.Peek(l.pos) // 0-based
			r := i.(rune)

			l.tokenLen -= utf8.RuneLen(r)

			l.column -= utf8.RuneLen(r)
		} else {
			panic("Underflow Exception")
		}
	}
}

// Lexer::EmitToken
func (l *lexer) EmitToken(t TokenType) {
	l.emit(t, false)
}

// Lexer::EmitTokenWithBytes
func (l *lexer) EmitTokenWithBytes(t TokenType) {
	l.emit(t, true)
}

// Lexer::EmitToken
func (l *lexer) EmitEOF() {
	l.emit(T_EOF, false)
}

// Lexer::IgnoreToken
func (l *lexer) IgnoreToken() {
	l.consume(false)
}

// Lexer::Marker
func (l *lexer) Marker() *Marker {
	return &Marker{sequence: l.sequence, pos: l.pos, tokenLen: l.tokenLen, line: l.line, column: l.column}
}

// Lexer::CanReset
func (l *lexer) CanReset(m *Marker) bool {
	return m.sequence == l.sequence && m.pos <= l.runes.Len() && m.tokenLen <= l.peekPos
}

// Lexer::Reset
func (l *lexer) Reset(m *Marker) {
	if l.CanReset(m) == false {
		panic("Invalid marker")
	}
	l.pos = m.pos

	l.tokenLen = m.tokenLen

	l.line = m.line

	l.column = m.column
}

// Lexer::MatchZeroOrOneBytes
func (l *lexer) MatchZeroOrOneBytes(match []byte) bool {
	if r := l.PeekRune(0); r != RuneEOF && bytes.IndexRune(match, r) >= 0 {
		l.NextRune()
	}
	return true
}

// Lexer::MatchZeroOrOneRunes
func (l *lexer) MatchZeroOrOneRunes(match []rune) bool {
	if r := l.PeekRune(0); r != RuneEOF && runes.IndexRune(match, r) >= 0 {
		l.NextRune()
	}
	return true
}

// Lexer::MatchZeroOrOneRune
func (l *lexer) MatchZeroOrOneRune(match rune) bool {
	if r := l.PeekRune(0); r != RuneEOF && r == match {
		l.NextRune()
	}
	return true
}

// Lexer::MatchZeroOrOneFunc
func (l *lexer) MatchZeroOrOneFunc(match MatchFn) bool {
	if r := l.PeekRune(0); r != RuneEOF && match(r) {
		l.NextRune()
	}
	return true
}

// Lexer::MatchZeroOrMoreBytes
func (l *lexer) MatchZeroOrMoreBytes(match []byte) bool {
	for r := l.PeekRune(0); r != RuneEOF && bytes.IndexRune(match, r) >= 0; r = l.PeekRune(0) {
		l.NextRune()
	}
	return true
}

// Lexer::MatchZeroOrMoreRunes
func (l *lexer) MatchZeroOrMoreRunes(match []rune) bool {
	for r := l.PeekRune(0); r != RuneEOF && runes.IndexRune(match, r) >= 0; r = l.PeekRune(0) {
		l.NextRune()
	}
	return true
}

// Lexer::MatchZeroOrMoreFunc
func (l *lexer) MatchZeroOrMoreFunc(match MatchFn) bool {
	for r := l.PeekRune(0); r != RuneEOF && match(r); r = l.PeekRune(0) {
		l.NextRune()
	}
	return true
}

// Lexer::MatchOneBytes
func (l *lexer) MatchOneBytes(match []byte) bool {
	if r := l.PeekRune(0); r != RuneEOF && bytes.IndexRune(match, r) >= 0 {
		l.NextRune()
		return true
	}
	return false
}

// Lexer::MatchOneRunes
func (l *lexer) MatchOneRunes(match []rune) bool {
	if r := l.PeekRune(0); r != RuneEOF && runes.IndexRune(match, r) >= 0 {
		l.NextRune()
		return true
	}
	return false
}

// Lexer::MatchOneRune
func (l *lexer) MatchOneRune(match rune) bool {
	if r := l.PeekRune(0); r != RuneEOF && r == match {
		l.NextRune()
		return true
	}
	return false
}

// Lexer::MatchOneFunc
func (l *lexer) MatchOneFunc(match MatchFn) bool {
	if r := l.PeekRune(0); r != RuneEOF && match(r) {
		l.NextRune()
		return true
	}
	return false
}

// Lexer::MatchOneOrMoreBytes
func (l *lexer) MatchOneOrMoreBytes(match []byte) bool {
	var r rune
	if r = l.PeekRune(0); r != RuneEOF && bytes.IndexRune(match, r) >= 0 {
		l.NextRune()
		for r = l.PeekRune(0); r != RuneEOF && bytes.IndexRune(match, r) >= 0; r = l.PeekRune(0) {
			l.NextRune()
		}
		return true
	}
	return false
}

// Lexer::MatchOneOrMoreRunes
func (l *lexer) MatchOneOrMoreRunes(match []rune) bool {
	var r rune
	if r = l.PeekRune(0); r != RuneEOF && runes.IndexRune(match, r) >= 0 {
		l.NextRune()
		for r = l.PeekRune(0); r != RuneEOF && runes.IndexRune(match, r) >= 0; r = l.PeekRune(0) {
			l.NextRune()
		}
		return true
	}
	return false
}

// Lexer::MatchOneOrMoreFunc
func (l *lexer) MatchOneOrMoreFunc(match MatchFn) bool {
	var r rune
	if r = l.PeekRune(0); r != RuneEOF && match(r) {
		l.NextRune()
		for r = l.PeekRune(0); r != RuneEOF && match(r); r = l.PeekRune(0) {
			l.NextRune()
		}
		return true
	}
	return false
}

// Lexer::MatchMinMaxBytes
func (l *lexer) MatchMinMaxBytes(match []byte, min int, max int) bool {
	marker := l.Marker()
	count := 0
	for r := l.PeekRune(0); r != RuneEOF && bytes.IndexRune(match, r) >= 0; r = l.PeekRune(0) {
		l.NextRune()
		count++
		if max > 0 && count >= max { // Check here to avoid unused PeekRune()
			break
		}
	}
	if count < min {
		l.Reset(marker)
		return false
	}
	return true
}

// Lexer::MatchMinMaxRunes
func (l *lexer) MatchMinMaxRunes(match []rune, min int, max int) bool {
	marker := l.Marker()
	count := 0
	for r := l.PeekRune(0); r != RuneEOF && runes.IndexRune(match, r) >= 0; r = l.PeekRune(0) {
		l.NextRune()
		count++
		if max > 0 && count >= max { // Check here to avoid unused PeekRune()
			break
		}
	}
	if count < min {
		l.Reset(marker)
		return false
	}
	return true
}

// Lexer::MatchMinMaxFunc
func (l *lexer) MatchMinMaxFunc(match MatchFn, min int, max int) bool {
	marker := l.Marker()
	count := 0
	for r := l.PeekRune(0); r != RuneEOF && match(r); r = l.PeekRune(0) {
		l.NextRune()
		count++
		if max > 0 && count >= max { // Check here to avoid unused PeekRune()
			break
		}
	}
	if count < min {
		l.Reset(marker)
		return false
	}
	return true
}

// Lexer::NonMatchOneBytes
func (l *lexer) NonMatchOneBytes(match []byte) bool {
	if r := l.PeekRune(0); r != RuneEOF && bytes.IndexRune(match, r) == -1 {
		l.NextRune()
		return true
	}
	return false
}

// Lexer::NonMatchOneRunes
func (l *lexer) NonMatchOneRunes(match []rune) bool {
	if r := l.PeekRune(0); r != RuneEOF && runes.IndexRune(match, r) == -1 {
		l.NextRune()
		return true
	}
	return false
}

// Lexer::NonMatchOneFunc
func (l *lexer) NonMatchOneFunc(match MatchFn) bool {
	if r := l.PeekRune(0); r != RuneEOF && match(r) == false {
		l.NextRune()
		return true
	}
	return false
}

// Lexer::NonMatchOneOrMoreBytes
func (l *lexer) NonMatchOneOrMoreBytes(match []byte) bool {
	var r rune
	if r = l.PeekRune(0); r != RuneEOF && bytes.IndexRune(match, r) == -1 {
		l.NextRune()
		for r = l.PeekRune(0); r != RuneEOF && bytes.IndexRune(match, r) == -1; r = l.PeekRune(0) {
			l.NextRune()
		}
		return true
	}
	return false
}

// Lexer::NonMatchOneOrMoreRunes
func (l *lexer) NonMatchOneOrMoreRunes(match []rune) bool {
	var r rune
	if r = l.PeekRune(0); r != RuneEOF && runes.IndexRune(match, r) == -1 {
		l.NextRune()
		for r = l.PeekRune(0); r != RuneEOF && runes.IndexRune(match, r) == -1; r = l.PeekRune(0) {
			l.NextRune()
		}
		return true
	}
	return false
}

// Lexer::NonMatchOneOrMoreFunc
func (l *lexer) NonMatchOneOrMoreFunc(match MatchFn) bool {
	var r rune
	if r = l.PeekRune(0); r != RuneEOF && match(r) == false {
		l.NextRune()
		for r = l.PeekRune(0); r != RuneEOF && match(r) == false; r = l.PeekRune(0) {
			l.NextRune()
		}
		return true
	}
	return false
}

// Lexer::NonMatchZeroOrOneBytes
func (l *lexer) NonMatchZeroOrOneBytes(match []byte) bool {
	if r := l.PeekRune(0); r != RuneEOF && bytes.IndexRune(match, r) == -1 {
		l.NextRune()
	}
	return true
}

// Lexer::NonMatchZeroOrOneRunes
func (l *lexer) NonMatchZeroOrOneRunes(match []rune) bool {
	if r := l.PeekRune(0); r != RuneEOF && runes.IndexRune(match, r) == -1 {
		l.NextRune()
	}
	return true
}

// Lexer::NonMatchZeroOrOneFunc
func (l *lexer) NonMatchZeroOrOneFunc(match MatchFn) bool {
	if r := l.PeekRune(0); r != RuneEOF && match(r) == false {
		l.NextRune()
	}
	return true
}

// Lexer::NonMatchZeroOrMoreBytes
func (l *lexer) NonMatchZeroOrMoreBytes(match []byte) bool {
	for r := l.PeekRune(0); r != RuneEOF && bytes.IndexRune(match, r) == -1; r = l.PeekRune(0) {
		l.NextRune()
	}
	return true
}

// Lexer::NonMatchZeroOrMoreRunes
func (l *lexer) NonMatchZeroOrMoreRunes(match []rune) bool {
	for r := l.PeekRune(0); r != RuneEOF && runes.IndexRune(match, r) == -1; r = l.PeekRune(0) {
		l.NextRune()
	}
	return true
}

// Lexer::NonMatchZeroOrMoreFunc
func (l *lexer) NonMatchZeroOrMoreFunc(match MatchFn) bool {
	for r := l.PeekRune(0); r != RuneEOF && match(r) == false; r = l.PeekRune(0) {
		l.NextRune()
	}
	return true
}

// Lexer::MatchEOF
func (l *lexer) MatchEOF() bool {
	if r := l.PeekRune(0); r == RuneEOF {
		l.NextRune()
		return true
	}
	return false
}
