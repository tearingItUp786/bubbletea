package main

import (
	"fmt"
	"io"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/term/ansi/ctrl"
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
		if msg.Action != tea.KeyPress {
			break
		}
		switch m.prevKey.String() {
		case "q":
			if msg.String() == "q" {
				cmd = tea.Quit
			}
		case "r":
			switch msg.String() {
			case "b":
				execute(sys.RequestBackgroundColor)
			case "d":
				execute(ctrl.RequestPrimaryDeviceAttributes)
			case "k":
				execute(kitty.Request)
			}
		case "k":
			switch msg.String() {
			case "0":
				m.kittyFlags = 0
			case "1":
				m.kittyFlags |= kitty.DisambiguateEscapeCodes
			case "2":
				m.kittyFlags |= kitty.ReportEventTypes
			case "3":
				m.kittyFlags |= kitty.ReportAlternateKeys
			case "4":
				m.kittyFlags |= kitty.ReportAllKeys
			case "5":
				m.kittyFlags |= kitty.ReportAssociatedKeys
			}
			execute(kitty.Push(m.kittyFlags))
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
	defer execute(kitty.Push(0))
	p := tea.NewProgram(model{})
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func execute(seq string) {
	io.WriteString(os.Stdout, seq)
}
