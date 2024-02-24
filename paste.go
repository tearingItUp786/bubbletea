package tea

import (
	"fmt"
	"unicode/utf8"
)

// PasteMsg represents a bracketed paste event.
type PasteMsg string

// String implements Event.
func (e PasteMsg) String() string {
	return fmt.Sprintf("paste: %q", string(e))
}

func parseBracketedPaste(p []byte, buf *[]byte) Msg {
	switch string(p) {
	case "\x1b[200~":
		*buf = []byte{}
	case "\x1b[201~":
		var paste []rune
		for len(*buf) > 0 {
			r, w := utf8.DecodeRune(*buf)
			if r != utf8.RuneError {
				*buf = (*buf)[w:]
			}
			paste = append(paste, r)
		}
		*buf = nil
		return PasteMsg(paste)
	}
	return nil
}
