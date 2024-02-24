package main

import (
	"fmt"
	"io"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/term/ansi/kitty"
	"github.com/charmbracelet/x/exp/term/ansi/sys"
)

type model struct {
	prevKey    tea.KeyMsg
	kittyFlags int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.prevKey.String() {
		case "q":
			if msg.String() == "q" {
				cmd = tea.Quit
			}
		case "r":
			switch msg.String() {
			case "b":
				execute(sys.RequestBackgroundColor)
			}
		case "k":
			switch msg.String() {
			case "0":
				m.kittyFlags = 0
				execute(kitty.Disable(m.kittyFlags))
			case "1":
				if m.kittyFlags&kitty.DisambiguateEscapeCodes == 0 {
					m.kittyFlags |= kitty.DisambiguateEscapeCodes
					execute(kitty.Enable(m.kittyFlags))
				} else {
					m.kittyFlags &^= kitty.DisambiguateEscapeCodes
					execute(kitty.Disable(m.kittyFlags))
				}
			case "2":
				if m.kittyFlags&kitty.ReportEventTypes == 0 {
					m.kittyFlags |= kitty.ReportEventTypes
					execute(kitty.Enable(m.kittyFlags))
				} else {
					m.kittyFlags &^= kitty.ReportEventTypes
					execute(kitty.Disable(m.kittyFlags))
				}
			case "3":
				if m.kittyFlags&kitty.ReportAlternateKeys == 0 {
					m.kittyFlags |= kitty.ReportAlternateKeys
					execute(kitty.Enable(m.kittyFlags))
				} else {
					m.kittyFlags &^= kitty.ReportAlternateKeys
					execute(kitty.Disable(m.kittyFlags))
				}
			case "4":
				if m.kittyFlags&kitty.ReportAllKeys == 0 {
					m.kittyFlags |= kitty.ReportAllKeys
					execute(kitty.Enable(m.kittyFlags))
				} else {
					m.kittyFlags &^= kitty.ReportAllKeys
					execute(kitty.Disable(m.kittyFlags))
				}
			case "5":
				if m.kittyFlags&kitty.ReportAssociatedKeys == 0 {
					m.kittyFlags |= kitty.ReportAssociatedKeys
					execute(kitty.Enable(m.kittyFlags))
				} else {
					m.kittyFlags &^= kitty.ReportAssociatedKeys
					execute(kitty.Disable(m.kittyFlags))
				}

			}
		}
		m.prevKey = msg
	}
	switch msg := msg.(type) {
	case string:
		return m, tea.Batch(tea.Println(msg), cmd)
	case fmt.Stringer:
		return m, tea.Batch(tea.Println(msg.String()), cmd)
	}
	return m, cmd
}

func (m model) View() string {
	return "Type any key and it will be printed to the terminal. Press qq to quit."
}

func main() {
	p := tea.NewProgram(model{})
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func execute(seq string) {
	io.WriteString(os.Stdout, seq)
}
