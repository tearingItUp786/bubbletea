package tea

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/charmbracelet/x/exp/term/input"
	"github.com/muesli/cancelreader"
	"golang.org/x/term"
)

func (p *Program) initTerminal() error {
	if err := p.initInput(); err != nil {
		return err
	}

	p.renderer.hideCursor()
	return nil
}

// restoreTerminalState restores the terminal to the state prior to running the
// Bubble Tea program.
func (p *Program) restoreTerminalState() error {
	if p.renderer != nil {
		p.renderer.disableBracketedPaste()
		p.renderer.showCursor()
		p.renderer.disableKeyboardEnhancement() // kitty keyboard protocol
		p.disableMouse()

		if p.renderer.altScreen() {
			p.renderer.exitAltScreen()

			// give the terminal a moment to catch up
			time.Sleep(time.Millisecond * 10) //nolint:gomnd
		}
	}

	return p.restoreInput()
}

// restoreInput restores the tty input to its original state.
func (p *Program) restoreInput() error {
	if p.ttyInput != nil && p.previousTtyInputState != nil {
		if err := term.Restore(int(p.ttyInput.Fd()), p.previousTtyInputState); err != nil {
			return fmt.Errorf("error restoring console: %w", err)
		}
	}
	if p.ttyOutput != nil && p.previousOutputState != nil {
		if err := term.Restore(int(p.ttyOutput.Fd()), p.previousOutputState); err != nil {
			return fmt.Errorf("error restoring console: %w", err)
		}
	}
	return nil
}

// initInputReader (re)commences reading inputs.
func (p *Program) initInputReader() error {
	// Initialize the input reader.
	// This need to be done after the terminal has been initialized and set to
	// raw mode.
	// On Windows, this will change the console mode to enable mouse and window
	// events.
	drv, err := input.NewDriver(p.input, os.Getenv("TERM"), 0)
	if err != nil {
		return err
	}

	p.inputReader = drv
	p.readLoopDone = make(chan struct{})
	go p.readLoop()

	return nil
}

func readInputs(ctx context.Context, msgs chan<- Msg, reader *input.Driver) error {
	var readEvents [16]input.Event
	for {
		n, err := reader.ReadInput(readEvents[:])
		if err != nil {
			return err
		}

		events := readEvents[:n]
		for _, e := range events {
			select {
			case msgs <- e:
			case <-ctx.Done():
				err := ctx.Err()
				if err != nil {
					err = fmt.Errorf("found context error while reading input: %w", err)
				}
				return err
			}
		}
	}
}

func (p *Program) readLoop() {
	defer close(p.readLoopDone)

	err := readInputs(p.ctx, p.msgs, p.inputReader)
	if !errors.Is(err, io.EOF) && !errors.Is(err, cancelreader.ErrCanceled) {
		select {
		case <-p.ctx.Done():
		case p.errs <- err:
		}
	}
}

// waitForReadLoop waits for the cancelReader to finish its read loop.
func (p *Program) waitForReadLoop() {
	select {
	case <-p.readLoopDone:
	case <-time.After(500 * time.Millisecond): //nolint:gomnd
		// The read loop hangs, which means the input
		// cancelReader's cancel function has returned true even
		// though it was not able to cancel the read.
	}
}

// checkResize detects the current size of the output and informs the program
// via a WindowSizeMsg.
func (p *Program) checkResize() {
	if p.ttyOutput == nil {
		// can't query window size
		return
	}

	w, h, err := term.GetSize(int(p.ttyOutput.Fd()))
	if err != nil {
		select {
		case <-p.ctx.Done():
		case p.errs <- err:
		}

		return
	}

	p.Send(WindowSizeMsg{
		Width:  w,
		Height: h,
	})
}
