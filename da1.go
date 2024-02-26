package tea

import (
	"fmt"
)

// PrimaryDeviceAttrsMsg represents a primary device attributes event.
type PrimaryDeviceAttrsMsg []uint

// String implements input.Event.
func (e PrimaryDeviceAttrsMsg) String() string {
	s := "DA1"
	if len(e) > 0 {
		s += fmt.Sprintf(": %v", []uint(e))
	}
	return s
}
