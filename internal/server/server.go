package server

import (
	"fmt"
	"net"
)

type Server struct {
	listener net.Listener
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	server := &Server{listener: listener}
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
	conn.Write([]byte("HTTP/1.1 200 OK \r\nContent-Type: text/plain \r\nHello World!\r\n\r\n"))
	conn.Close()
}
