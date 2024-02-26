//go:build windows
// +build windows

package tea

import (
	"context"
	"fmt"
	"io"

	"github.com/erikgeiser/coninput"
	localereader "github.com/mattn/go-localereader"
	"golang.org/x/sys/windows"
)

func readInputs(ctx context.Context, msgs chan<- Msg, input io.Reader) error {
	if coninReader, ok := input.(*conInputReader); ok {
		return readConInputs(ctx, msgs, coninReader.conin)
	}

	return readAnsiInputs(ctx, msgs, localereader.NewReader(input))
}

func readConInputs(ctx context.Context, msgsch chan<- Msg, con windows.Handle) error {
	var ps coninput.ButtonState // keep track of previous mouse state
	for {
		events, err := coninput.ReadNConsoleInputs(con, 16)
		if err != nil {
			return fmt.Errorf("read coninput events: %w", err)
		}

		for _, event := range events {
			var msgs []Msg
			switch e := event.Unwrap().(type) {
			case coninput.KeyEventRecord:
				if !e.KeyDown || e.VirtualKeyCode == coninput.VK_SHIFT {
					continue
				}

				for i := 0; i < int(e.RepeatCount); i++ {

					var mod Mod
					if e.ControlKeyState.Contains(coninput.LEFT_ALT_PRESSED | coninput.RIGHT_ALT_PRESSED) {
						mod |= Alt
					}
					var k KeyMsg
					if e.Char != 0 {
						k = vkCharMap[e.Char]
					} else if vk, ok := vkKeyMap[e.VirtualKeyCode]; ok {
						k = vk
					} else {
						k = KeyMsg{
							Runes: []rune{e.Char},
						}
					}
					k.Mod |= mod
					msgs = append(msgs, k)
				}
			case coninput.WindowBufferSizeEventRecord:
				msgs = append(msgs, WindowSizeMsg{
					Width:  int(e.Size.X),
					Height: int(e.Size.Y),
				})
			case coninput.MouseEventRecord:
				event := mouseEvent(ps, e)
				msgs = append(msgs, event)
				ps = e.ButtonState
			case coninput.FocusEventRecord, coninput.MenuEventRecord:
				// ignore
			default: // unknown event
				continue
			}

			// Send all messages to the channel
			for _, msg := range msgs {
				select {
				case msgsch <- msg:
				case <-ctx.Done():
					err := ctx.Err()
					if err != nil {
						return fmt.Errorf("coninput context error: %w", err)
					}
					return err
				}
			}
		}
	}
}

func mouseEventButton(p, s coninput.ButtonState) (button MouseButton, action MouseAction) {
	btn := p ^ s
	action = MouseActionPress
	if btn&s == 0 {
		action = MouseActionRelease
	}

	if btn == 0 {
		switch {
		case s&coninput.FROM_LEFT_1ST_BUTTON_PRESSED > 0:
			button = MouseButtonLeft
		case s&coninput.FROM_LEFT_2ND_BUTTON_PRESSED > 0:
			button = MouseButtonMiddle
		case s&coninput.RIGHTMOST_BUTTON_PRESSED > 0:
			button = MouseButtonRight
		case s&coninput.FROM_LEFT_3RD_BUTTON_PRESSED > 0:
			button = MouseButtonBackward
		case s&coninput.FROM_LEFT_4TH_BUTTON_PRESSED > 0:
			button = MouseButtonForward
		}
		return
	}

	switch {
	case btn == coninput.FROM_LEFT_1ST_BUTTON_PRESSED: // left button
		button = MouseButtonLeft
	case btn == coninput.RIGHTMOST_BUTTON_PRESSED: // right button
		button = MouseButtonRight
	case btn == coninput.FROM_LEFT_2ND_BUTTON_PRESSED: // middle button
		button = MouseButtonMiddle
	case btn == coninput.FROM_LEFT_3RD_BUTTON_PRESSED: // unknown (possibly mouse backward)
		button = MouseButtonBackward
	case btn == coninput.FROM_LEFT_4TH_BUTTON_PRESSED: // unknown (possibly mouse forward)
		button = MouseButtonForward
	}

	return button, action
}

func mouseEvent(p coninput.ButtonState, e coninput.MouseEventRecord) MouseMsg {
	var mod Mod
	if e.ControlKeyState.Contains(coninput.LEFT_ALT_PRESSED | coninput.RIGHT_ALT_PRESSED) {
		mod |= Alt
	}
	if e.ControlKeyState.Contains(coninput.LEFT_CTRL_PRESSED | coninput.RIGHT_CTRL_PRESSED) {
		mod |= Ctrl
	}
	if e.ControlKeyState.Contains(coninput.SHIFT_PRESSED) {
		mod |= Shift
	}
	ev := MouseMsg{
		X:   int(e.MousePositon.X),
		Y:   int(e.MousePositon.Y),
		Mod: mod,
	}
	switch e.EventFlags {
	case coninput.CLICK, coninput.DOUBLE_CLICK:
		ev.Button, ev.Action = mouseEventButton(p, e.ButtonState)
		if ev.Action == MouseActionRelease {
			ev.Type = MouseRelease
		}
		switch ev.Button {
		case MouseButtonLeft:
			ev.Type = MouseLeft
		case MouseButtonMiddle:
			ev.Type = MouseMiddle
		case MouseButtonRight:
			ev.Type = MouseRight
		case MouseButtonBackward:
			ev.Type = MouseBackward
		case MouseButtonForward:
			ev.Type = MouseForward
		}
	case coninput.MOUSE_WHEELED:
		if e.WheelDirection > 0 {
			ev.Button = MouseButtonWheelUp
			ev.Type = MouseWheelUp
		} else {
			ev.Button = MouseButtonWheelDown
			ev.Type = MouseWheelDown
		}
	case coninput.MOUSE_HWHEELED:
		if e.WheelDirection > 0 {
			ev.Button = MouseButtonWheelRight
			ev.Type = MouseWheelRight
		} else {
			ev.Button = MouseButtonWheelLeft
			ev.Type = MouseWheelLeft
		}
	case coninput.MOUSE_MOVED:
		ev.Button, _ = mouseEventButton(p, e.ButtonState)
		ev.Action = MouseActionMotion
		ev.Type = MouseMotion
	}

	return ev
}

var vkKeyMap = map[coninput.VirtualKeyCode]KeyMsg{
	coninput.VK_RETURN: {Sym: KeyEnter},
	coninput.VK_BACK:   {Sym: KeyBackspace},
	coninput.VK_TAB:    {Sym: KeyTab},
	coninput.VK_SPACE:  {Sym: KeySpace, Runes: []rune{' '}},
	coninput.VK_ESCAPE: {Sym: KeyEscape},
	coninput.VK_UP:     {Sym: KeyUp},
	coninput.VK_DOWN:   {Sym: KeyDown},
	coninput.VK_RIGHT:  {Sym: KeyRight},
	coninput.VK_LEFT:   {Sym: KeyLeft},
	coninput.VK_HOME:   {Sym: KeyHome},
	coninput.VK_END:    {Sym: KeyEnd},
	coninput.VK_PRIOR:  {Sym: KeyPgUp},
	coninput.VK_NEXT:   {Sym: KeyPgDown},
	coninput.VK_DELETE: {Sym: KeyDelete},
	coninput.VK_OEM_4:  {Mod: Ctrl, Runes: []rune{'['}},
}

var vkCharMap = map[rune]KeyMsg{
	'@':    {Mod: Ctrl, Runes: []rune{'@'}},
	'\x01': {Mod: Ctrl, Runes: []rune{'a'}},
	'\x02': {Mod: Ctrl, Runes: []rune{'b'}},
	'\x03': {Mod: Ctrl, Runes: []rune{'c'}},
	'\x04': {Mod: Ctrl, Runes: []rune{'d'}},
	'\x05': {Mod: Ctrl, Runes: []rune{'e'}},
	'\x06': {Mod: Ctrl, Runes: []rune{'f'}},
	'\a':   {Mod: Ctrl, Runes: []rune{'g'}},
	'\b':   {Mod: Ctrl, Runes: []rune{'h'}},
	'\t':   {Mod: Ctrl, Runes: []rune{'i'}},
	'\n':   {Mod: Ctrl, Runes: []rune{'j'}},
	'\v':   {Mod: Ctrl, Runes: []rune{'k'}},
	'\f':   {Mod: Ctrl, Runes: []rune{'l'}},
	'\r':   {Mod: Ctrl, Runes: []rune{'m'}},
	'\x0e': {Mod: Ctrl, Runes: []rune{'n'}},
	'\x0f': {Mod: Ctrl, Runes: []rune{'o'}},
	'\x10': {Mod: Ctrl, Runes: []rune{'p'}},
	'\x11': {Mod: Ctrl, Runes: []rune{'q'}},
	'\x12': {Mod: Ctrl, Runes: []rune{'r'}},
	'\x13': {Mod: Ctrl, Runes: []rune{'s'}},
	'\x14': {Mod: Ctrl, Runes: []rune{'t'}},
	'\x15': {Mod: Ctrl, Runes: []rune{'u'}},
	'\x16': {Mod: Ctrl, Runes: []rune{'v'}},
	'\x17': {Mod: Ctrl, Runes: []rune{'w'}},
	'\x18': {Mod: Ctrl, Runes: []rune{'x'}},
	'\x19': {Mod: Ctrl, Runes: []rune{'y'}},
	'\x1a': {Mod: Ctrl, Runes: []rune{'z'}},
	'\x1b': {Mod: Ctrl, Runes: []rune{']'}},
	'\x1c': {Mod: Ctrl, Runes: []rune{'\\'}},
	'\x1f': {Mod: Ctrl, Runes: []rune{'_'}},
}
