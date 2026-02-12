package server

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"log"
	"net"
	"sync/atomic"
)

type HandlerFn func(w *response.Writer, req *request.Request)

type Server struct {
	listener  net.Listener
	handlerFn HandlerFn
	running   atomic.Bool
}

func Serve(port int, fn HandlerFn) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("error listening on port %d: %v", port, err)
	}
	s := &Server{
		listener:  listener,
		handlerFn: fn,
	}
	s.running.Store(true)
	go s.listen()
	return s, nil
}

func (s *Server) Close() error {
	s.running.Store(false)
	return s.listener.Close()
}

func (s *Server) listen() {
	log.Println("server listening on", s.listener.Addr())
	for s.running.Load() {
		conn, err := s.listener.Accept()
		if !s.running.Load() {
			break
		}
		if err != nil {
			log.Printf("error accepting connection from %s: %v", conn.RemoteAddr(), err)
			continue
		}
		log.Println("accepted connection from", conn.RemoteAddr())
		go s.handle(conn)
	}
	log.Println("server exitting")
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	req, err := request.RequestFromReader(conn)
	w := response.NewWriter(conn)
	if err != nil {
		log.Printf("error parsing request: %v", err)
		msg := err.Error()
		if err := w.WriteStatusLine(response.StatusCodeBadRequest); err != nil {
			log.Printf("error writing error response status line: %v", err)
			return
		}
		if err := w.WriteHeaders(response.GetDefaultHeaders(len(msg))); err != nil {
			log.Printf("error writing error response header: %v", err)
			return
		}
		if err := w.WriteBody([]byte(msg)); err != nil {
			log.Printf("error writing error response header: %v", err)
			return
		}
		return
	}
	s.handlerFn(w, req)
	log.Println("handled connection from", conn.RemoteAddr())
}
