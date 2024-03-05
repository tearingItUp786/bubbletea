package tea

import (
	"github.com/charmbracelet/x/exp/term/input"
)

// MouseMsg contains information about a mouse event and are sent to a programs
// update function when mouse activity occurs. Note that the mouse must first
// be enabled in order for the mouse events to be received.
type MouseMsg = input.MouseEvent

// Mouse event actions.
const (
	MouseActionPress   = input.MouseActionPress
	MouseActionRelease = input.MouseActionRelease
	MouseActionMotion  = input.MouseActionMotion
)

// Mouse event buttons
//
// This is based on X11 mouse button codes.
//
//	1 = left button
//	2 = middle button (pressing the scroll wheel)
//	3 = right button
//	4 = turn scroll wheel up
//	5 = turn scroll wheel down
//	6 = push scroll wheel left
//	7 = push scroll wheel right
//	8 = 4th button (aka browser backward button)
//	9 = 5th button (aka browser forward button)
//	10
//	11
//
// Other buttons are not supported.
const (
	MouseButtonNone       = input.MouseButtonNone
	MouseButtonLeft       = input.MouseButtonLeft
	MouseButtonMiddle     = input.MouseButtonMiddle
	MouseButtonRight      = input.MouseButtonRight
	MouseButtonWheelUp    = input.MouseButtonWheelUp
	MouseButtonWheelDown  = input.MouseButtonWheelDown
	MouseButtonWheelLeft  = input.MouseButtonWheelLeft
	MouseButtonWheelRight = input.MouseButtonWheelRight
	MouseButtonBackward   = input.MouseButtonBackward
	MouseButtonForward    = input.MouseButtonForward
	MouseButton10         = input.MouseButton10
	MouseButton11         = input.MouseButton11
)
