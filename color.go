package tea

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

// FgColorMsg represents a foreground color change event.
type FgColorMsg struct{ color.Color }

// String implements input.Event.
func (e FgColorMsg) String() string {
	r, g, b, a := e.RGBA()
	r >>= 8
	g >>= 8
	b >>= 8
	a >>= 8
	return fmt.Sprintf("FgColor: [%02x]#%02x%02x%02x", a, r, g, b)
}

// BgColorMsg represents a background color change event.
type BgColorMsg struct{ color.Color }

// String implements input.Event.
func (e BgColorMsg) String() string {
	r, g, b, a := e.RGBA()
	r >>= 8
	g >>= 8
	b >>= 8
	a >>= 8
	return fmt.Sprintf("BgColor: [%02x]#%02x%02x%02x", a, r, g, b)
}

// CursorColorMsg represents a cursor color change event.
type CursorColorMsg struct{ color.Color }

// String implements input.Event.
func (e CursorColorMsg) String() string {
	r, g, b, a := e.RGBA()
	r >>= 8
	g >>= 8
	b >>= 8
	a >>= 8
	return fmt.Sprintf("CursorColor: [%02x]#%02x%02x%02x", a, r, g, b)
}

func xParseColor(s string) color.Color {
	switch {
	case strings.HasPrefix(s, "rgb:"):
		parts := strings.Split(s[4:], "/")
		if len(parts) != 3 {
			return color.Black
		}

		r, _ := strconv.ParseUint(parts[0], 16, 32)
		g, _ := strconv.ParseUint(parts[1], 16, 32)
		b, _ := strconv.ParseUint(parts[2], 16, 32)

		return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
	case strings.HasPrefix(s, "rgba:"):
		parts := strings.Split(s[5:], "/")
		if len(parts) != 4 {
			return color.Black
		}

		r, _ := strconv.ParseUint(parts[0], 16, 32)
		g, _ := strconv.ParseUint(parts[1], 16, 32)
		b, _ := strconv.ParseUint(parts[2], 16, 32)
		a, _ := strconv.ParseUint(parts[3], 16, 32)

		return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
	}
	return color.Black
}
