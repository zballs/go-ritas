package types

import (
	"errors"
	log "github.com/inconshreveable/log15"
	"io"
	"net"
	"sync/atomic"
)

type Server struct {
	Addr string
	net.Listener

	Listening uint32
	Started   uint32
	Stopped   uint32

	Msgs chan *Message
	Errs chan error

	log.Logger
}

func NewServer(addr string) *Server {
	return &Server{
		Addr:   addr,
		Msgs:   make(chan *Message),
		Errs:   make(chan error),
		Logger: log.New("module", "server"),
	}
}

func (s *Server) Listen() error {

	l, err := net.Listen("tcp", s.Addr)

	if err != nil {
		return err
	}

	s.Listener = l

	swapped := atomic.CompareAndSwapUint32(&s.Listening, 0, 1)

	if !swapped {
		return errors.New("Error: already listening")
	}

	return nil
}

func (s *Server) Start() error {

	var err error

	if atomic.LoadUint32(&s.Listening) == 0 {

		err = s.Listen()

		if err != nil {
			return err
		}
	}

	if atomic.LoadUint32(&s.Stopped) == 1 {
		return errors.New("Error: not starting; already stopped")
	}

	swapped := atomic.CompareAndSwapUint32(&s.Started, 0, 1)

	if !swapped {
		return errors.New("Error: not starting; already started")
	}

	return nil
}

func (s *Server) Stop() error {

	swapped := atomic.CompareAndSwapUint32(&s.Stopped, 0, 1)

	if !swapped {
		return errors.New("Error: not stopping; already stopped")
	}

	s.Close()

	return nil
}

func (s *Server) Reset() error {

	swapped := atomic.CompareAndSwapUint32(&s.Stopped, 1, 0)

	if !swapped {
		return errors.New("Error: not resetting; not stopped")
	}

	atomic.CompareAndSwapUint32(&s.Started, 1, 0)

	return nil
}

func (s *Server) IsRunning() bool {
	return atomic.LoadUint32(&s.Started) == 1 && atomic.LoadUint32(&s.Stopped) == 0
}

func (s *Server) ReadRoutine() {

	for {

		if !s.IsRunning() {
			return
		}

		// Accept conn
		conn, err := s.Accept()

		if err != nil {
			s.Errs <- err
			continue
		}

		// Read one message from conn
		msg, err := ReadMessage(conn)

		if err != nil {

			if err != io.EOF {
				s.Errs <- err
			}

			continue
		}

		s.Msgs <- msg

		// Close conn
		conn.Close()
	}
}
