package tea

import (
	"fmt"
)

// UnknownSequenceMsg represents an unknown event.
type UnknownSequenceMsg string

// String implements Event.
func (e UnknownSequenceMsg) String() string {
	return fmt.Sprintf("unknown event: %q", string(e))
}
