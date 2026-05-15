package httplib

import (
	"fmt"
	"net"
)

type Response struct {
	StatusCode   int    // def 200 OK
	Proto        string // HTTP/1.1
	ReasonPhrase string
	Header       Headers
	Body         []byte
}

func NewResponse() *Response {
	return &Response{
		Proto:      "HTTP/1.1",
		Header:     NewHeaders(),
		StatusCode: StatusOK, // default status code 200 OK
	}
}

func (r *Response) Write(conn net.Conn) error {

	if _, err := fmt.Fprintf(conn, "%s %d %s\r\n", r.Proto, r.StatusCode, r.ReasonPhrase); err != nil {
		return err
	}

	if len(r.Body) > 0 {
		r.Header.Set("Content-Length", fmt.Sprintf("%d", len(r.Body)))
	}

	for key, val := range r.Header {
		for _, values := range val {
			if _, err := fmt.Fprintf(conn, "%s: %s\r\n", key, values); err != nil {
				return err
			}
		}
	}

	if _, err := fmt.Fprintf(conn, "\r\n"); err != nil {
		return err
	}

	if len(r.Body) > 0 {
		if _, err := conn.Write(r.Body); err != nil {
			return err
		}
	}

	return nil
}

func (r *Response) SetHeader(key, val string) {
	r.Header.Set(key, val)
}

func (r *Response) AddHeader(key, val string) {
	r.Header.Add(key, val)
}
