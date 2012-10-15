/*
Package rangeutil provides services for conversion and iteration of range
specifications for use with iNamik/go_lexer

Currently, a 'range specifiction' is defined as a string of unicode characters,
with the ability to specify a range of chararacters by using a '-' between two
characters.  Think of it as a simplified version of standard regex bracket
expression.

Here are some examples:

	Digit        "0123456789" // With no range specifiers
	Digit        "0-9"        // With range specifiers
	Hex Digit    "0-9a-zA-F"  // Allowing for upper and lower case
	Decimal      "-.0-9"      // '-' At beginning means litteral '-'

NOTE: The current implentation of the range specification processer may
allow a literal '-' anywhere that a range is not implied, but the only officially
supported location for a literal '-' is at the beginning of the string.

NOTE: The full unicode character set should be supported
*/
package rangeutil

import (
	"bytes"
	"unicode/utf8"
)

// RangeIteratorCallback defines the prototype for the IterateRangeSpec
// callback function.  The function takes two rune parameters, 'low' and 'hi'
// (which can have the same value) and returns a boolean indicating if it is
// ok to continue processing the rangeSpec (true == ok, false == quit)
type RangeIteratorCallback func(low, hi rune) (ok bool)

// IterateRangeSpec processes a range specification, calling a callback function
// for each rune-range encountered
func IterateRangeSpec(rangeSpec string, callBack RangeIteratorCallback) {
	for i, l := 0, len(rangeSpec); i < l; {
		var low, hi rune

		var size int

		low, size = utf8.DecodeRuneInString(rangeSpec[i:])

		if low == utf8.RuneError {
			panic("error in range spec - invalid rune incountered at index " + string(i))
		}

		i += size

		// Peek at next byte to see if it is '-'
		// NOTE: This *should* be safe
		if i < l && rangeSpec[i] == '-' {
			i++ // skip '-'

			if i < l {
				hi, size = utf8.DecodeRuneInString(rangeSpec[i:])

				if hi == utf8.RuneError {
					panic("error in range spec - invalid rune incountered at index " + string(i))
				}

				i += size

				if low <= hi {
					if callBack(low, hi) == false {
						return
					}
				} else {
					panic("error in range spec - range not low-to-high")
				}
			} else {
				panic("error in range spec - trailing '-'")
			}
		} else {
			if callBack(low, low) == false {
				return
			}
		}
	}
}

// RangeToBytes converts a range specifier into a byte array suitable for
// the Match*Bytes calls of iNamik/go_lexer
func RangeToBytes(rangeSpec string) []byte {
	bytes := new(bytes.Buffer)

	IterateRangeSpec(rangeSpec, func(low, hi rune) bool {
		for ; low <= hi; low++ {
			_, error := bytes.WriteRune(low)
			if error != nil {
				panic(error)
			}
		}
		return true
	})
	return bytes.Bytes()
}

// RangeToRunes converts a range specifier into a rune array suitable for
// the Match*Runes calls of the iNamik/go_lexer
func RangeToRunes(rangeSpec string) []rune {
	runes := make([]rune, 1)
	IterateRangeSpec(rangeSpec, func(low, hi rune) bool {
		for ; low <= hi; low++ {
			runes = append(runes, low)
		}
		return true
	})
	return runes
}
