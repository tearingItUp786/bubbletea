package tea

import (
	"fmt"

	"github.com/charmbracelet/x/exp/term/ansi"
)

// UnknownCsiMsg represents an unknown CSI sequence event.
type UnknownCsiMsg struct {
	ansi.CsiSequence
}

// String implements input.Event.
func (e UnknownCsiMsg) String() string {
	return fmt.Sprintf("unknown CSI sequence: %q", e.CsiSequence)
}

// UnknownOscMsg represents an unknown OSC sequence event.
type UnknownOscMsg struct {
	ansi.OscSequence
}

// String implements input.Event.
func (e UnknownOscMsg) String() string {
	return fmt.Sprintf("unknown OSC sequence: %q", e.OscSequence)
}
