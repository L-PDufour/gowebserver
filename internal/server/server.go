package server

import (
	"fmt"
	"net"

	"gowebserver/internal/request"
	"gowebserver/internal/response"
)

type Server struct {
	listener net.Listener
	handler  Handler
}

type Handler func(w *response.Writer, req *request.Request)

func Serve(port int, h Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	server := &Server{listener: listener, handler: h}
	go server.listen()
	return server, nil
}

func (s *Server) Close() error {
	err := s.listener.Close()
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) listen() {

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go s.handle(conn) // ← goroutine here instead
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	r, err := request.RequestFromReader(conn)
	if err != nil {
		w := response.NewWriter(conn)
		w.WriteStatusLine(response.StatusBadRequest)
		h := response.GetDefaultHeaders(len(err.Error()))
		h.Set("Content-Type", "text/html")
		w.WriteHeaders(h)
		w.WriteBody([]byte(err.Error()))
		return
	}
	w := response.NewWriter(conn)
	s.handler(w, r)
}
