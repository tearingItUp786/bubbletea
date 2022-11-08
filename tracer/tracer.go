package tracer

import (
	"encoding/json"
	"net"
	"sync"
	"time"
)

type Tracer interface {
	Command(c Command)
	Message(m Message)
}

type Event struct {
	Timestamp time.Time
	Message   Message
	Command   Command
}

type Message struct {
	Model string
	Type  string
	Msg   string
}

type Command struct {
	ID       int
	Started  time.Time
	Finished time.Time
	Type     string
	Msg      string
}

func NewCommand() Command {
	mtx.Lock()
	defer mtx.Unlock()
	id++

	return Command{
		ID:      id,
		Started: time.Now(),
	}
}

var (
	ch  = make(chan Event)
	mtx sync.Mutex
	id  int
)

type RemoteTracer struct {
}

func NewRemoteTracer() (*RemoteTracer, error) {
	listen, err := net.Listen("tcp", ":13337")
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			conn, err := listen.Accept()
			if err != nil {
				return
			}
			go handleTraceConn(conn)
		}
	}()

	t := &RemoteTracer{}
	return t, nil
}

func (rt *RemoteTracer) Command(c Command) {
	ev := Event{
		Timestamp: time.Now(),
		Command:   c,
	}

	ch <- ev
}

func (rt *RemoteTracer) Message(m Message) {
	ev := Event{
		Timestamp: time.Now(),
		Message:   m,
	}

	ch <- ev
}

func handleTraceConn(conn net.Conn) {
	for ev := range ch {
		b, _ := json.Marshal(ev)
		_, _ = conn.Write(b)
		_, err := conn.Write([]byte("\n"))
		if err != nil {
			return
		}
	}

	// close conn
	_ = conn.Close()
}
