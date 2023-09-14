package tea

import (
	"fmt"
	"log"

	"github.com/erikgeiser/coninput"
	"golang.org/x/sys/windows"
)

func enableWindowsConInput(p *Program) (func() error, error) {
	con, err := windows.GetStdHandle(windows.STD_INPUT_HANDLE)
	if err != nil {
		return nil, fmt.Errorf("get stdin handle: %w", err)
	}

	p.conInput = con

	var originalConsoleMode uint32

	err = windows.GetConsoleMode(con, &originalConsoleMode)
	if err != nil {
		return nil, fmt.Errorf("get console mode: %w", err)
	}

	log.Println("Input mode:", coninput.DescribeInputMode(originalConsoleMode))

	newConsoleMode := coninput.AddInputModes(
		0,
		windows.ENABLE_WINDOW_INPUT,
		windows.ENABLE_MOUSE_INPUT,
		// windows.ENABLE_PROCESSED_INPUT,
		windows.ENABLE_EXTENDED_FLAGS,
		// windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING,
	)

	log.Println("Setting mode to:", coninput.DescribeInputMode(newConsoleMode))

	err = windows.SetConsoleMode(con, newConsoleMode)
	if err != nil {
		return nil, fmt.Errorf("set console mode: %w", err)
	}

	cancelEvent, err := windows.CreateEvent(nil, 0, 0, nil)
	if err != nil {
		return nil, fmt.Errorf("create stop event: %w", err)
	}

	p.cancelEvent = cancelEvent

	return func() error {
		log.Println("Resetting input mode to:", coninput.DescribeInputMode(originalConsoleMode))

		if err := windows.CloseHandle(p.cancelEvent); err != nil {
			return fmt.Errorf("close stop event: %w", err)
		}

		resetErr := windows.SetConsoleMode(con, originalConsoleMode)
		if err == nil && resetErr != nil {
			return fmt.Errorf("reset console mode: %w", resetErr)
		}

		return nil
	}, nil
}

func (p *Program) handleConInput() chan struct{} {
	ch := make(chan struct{})

	go func() {
		defer func() {
			windows.SetEvent(p.cancelEvent)
			close(ch)
		}()
		for {
			select {
			case <-p.ctx.Done():
				return
			default:
				if p.ctx.Err() != nil {
					return
				}

				if err := waitForInput(p.conInput, p.cancelEvent); err != nil {
					log.Printf("wait for input: %s", err)
					return
				}

				n, err := coninput.ReadConsoleInput(p.conInput, p.inputEvents)
				if err != nil {
					log.Printf("read input events: %s", err)
					return
				}

				log.Printf("Read %d events:\n", n)
				for _, event := range p.inputEvents[:n] {
					log.Println("  ", event)
					switch e := event.Unwrap().(type) {
					case coninput.WindowBufferSizeEventRecord:
						p.msgs <- WindowSizeMsg{
							Width:  int(e.Size.X),
							Height: int(e.Size.Y),
						}
					}
				}

			}
		}
	}()

	return ch
}

var errCanceled = fmt.Errorf("read cancelled")

func waitForInput(conin, cancel windows.Handle) error {
	event, err := windows.WaitForMultipleObjects([]windows.Handle{conin, cancel}, false, windows.INFINITE)
	switch {
	case windows.WAIT_OBJECT_0 <= event && event < windows.WAIT_OBJECT_0+2:
		if event == windows.WAIT_OBJECT_0+1 {
			return errCanceled
		}

		if event == windows.WAIT_OBJECT_0 {
			return nil
		}

		return fmt.Errorf("unexpected wait object is ready: %d", event-windows.WAIT_OBJECT_0)
	case windows.WAIT_ABANDONED <= event && event < windows.WAIT_ABANDONED+2:
		return fmt.Errorf("abandoned")
	case event == uint32(windows.WAIT_TIMEOUT):
		return fmt.Errorf("timeout")
	case event == windows.WAIT_FAILED:
		return fmt.Errorf("failed")
	default:
		return fmt.Errorf("unexpected error: %w", error(err))
	}
}

type overlappedReader windows.Handle

// Read performs an overlapping read fom a windows.Handle.
func (r overlappedReader) Read(data []byte) (int, error) {
	hevent, err := windows.CreateEvent(nil, 0, 0, nil)
	if err != nil {
		return 0, fmt.Errorf("create event: %w", err)
	}

	overlapped := windows.Overlapped{HEvent: hevent}

	var n uint32

	err = windows.ReadFile(windows.Handle(r), data, &n, &overlapped)
	if err != nil && err != windows.ERROR_IO_PENDING {
		return int(n), err
	}

	err = windows.GetOverlappedResult(windows.Handle(r), &overlapped, &n, true)
	if err != nil {
		return int(n), nil
	}

	return int(n), nil
}
