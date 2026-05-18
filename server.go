package httplib

import (
	"fmt"
	"net"
	"time"
)

type Server struct {
	Addr        string
	Handler     Handler
	MaxBodySize int // in bytes
	ReadTimeout time.Duration
}

func (s *Server) ListenAndServe() error {

	addr := s.Addr
	if addr == "" {
		addr = ":http"
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	fmt.Printf("listens on addr: %s", ln.Addr().String())
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("[ERROR] Accept connection failed: %v\n", err)
			continue
		}

		go s.handleConnection(conn)
	}

}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	if s.ReadTimeout > 0 {
		conn.SetReadDeadline(time.Now().Add(s.ReadTimeout))
	}

	resp := NewResponse()
	req, err := s.ParseRequest(conn)
	if err != nil {
		if httpErr, ok := err.(*HttpError); ok {
			resp.StatusCode = httpErr.StatusCode
			resp.ReasonPhrase = httpErr.Message
			resp.Write(conn)
		} else {
			resp.StatusCode = StatusBadRequest
			resp.ReasonPhrase = "Bad request"
			resp.Write(conn)
			fmt.Println(err.Error())
		}
		return
	}

	conn.SetReadDeadline(time.Time{})
	resp.Conn = &conn

	s.Handler.ServeHTTP(*resp, req)
}
