package httplib

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
)

type Request struct {
	Method  string
	Path    string
	Version string
	Headers Headers
	Body    io.Reader
	Proto   string
	URL     *url.URL
	ctx     context.Context
}

func (s *Server) ParseRequest(conn net.Conn) (*Request, error) {

	reader := bufio.NewReader(conn)

	req := &Request{}

	// var buffer []byte
	// buffer,_=io.ReadAll(reader)

	// fmt.Println(string(buffer))

	// 1. Request Line
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, NewError(StatusInternalServerError, err.Error())
	}

	requestLine = strings.TrimSpace(requestLine)
	if requestLine == "" {
		return nil, NewError(StatusBadRequest, "empty request line")
	}

	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		return nil, NewError(StatusBadRequest, "invalid request line format")
	}

	req.Method, req.Path, req.Proto = parts[0], parts[1], parts[2]

	host := req.Headers.Get("Host")
	rawURL := "http://" + host + req.Path

	req.URL, err = url.Parse(rawURL)

	// 2. Headers

	req.Headers = NewHeaders()

	for {

		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, NewError(StatusInternalServerError, err.Error())
		}

		line = strings.TrimSpace(line)

		if line == "" {
			break // stop on empty line
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue // skping invalid Headers
		}

		key := textproto.CanonicalMIMEHeaderKey(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])

		req.Headers.Add(key, val)
	}

	if req.Headers.Get("Host") == "" {
		return nil, NewError(StatusBadRequest, "missing mandatory Host header")
	}

	// 3. Body

	if req.Headers.Get("Transfer-Encoding") == "chunked" {

		fmt.Println(req.Headers.Get("Transfer-Encoding"))
		var bodyBuffer []byte

		for {
			sizeLine, err := reader.ReadString('\n')
			if err != nil {
				return nil, &HttpError{StatusCode: StatusBadRequest, Message: "Bad Request: Failed to read chunk size"}
			}

			sizeLine = strings.TrimSpace(sizeLine)

			chunkSize, err := strconv.ParseInt(sizeLine, 16, 64)
			if err != nil {
				return nil, &HttpError{StatusCode: StatusBadRequest, Message: "Bad Request: Invalid chunk size format"}
			}

			if chunkSize == 0 {

				reader.ReadString('\n')
				break
			}

			chunkData := make([]byte, chunkSize)
			_, err = io.ReadFull(reader, chunkData)
			if err != nil {
				return nil, &HttpError{StatusCode: StatusBadRequest, Message: "Bad Request: Failed to read chunk data"}
			}

			bodyBuffer = append(bodyBuffer, chunkData...)

			reader.ReadString('\n')
		}

		req.Body = io.NopCloser(bytes.NewReader(bodyBuffer))
		return req, nil
	}

	contentLengthStr := req.Headers.Get("Content-Length")
	if contentLengthStr == "" {
		return req, nil // returning nil body if content-length not provided
	}

	contentLength, err := strconv.Atoi(contentLengthStr)
	if err != nil {
		return nil, NewError(StatusBadRequest, "Invalid Content-Length")
	}

	if contentLength > s.MaxBodySize {
		return nil, NewError(StatusRequestEntityTooLarge, "Payload Too Large")
	}

	body := make([]byte, contentLength)

	n, err := io.ReadFull(reader, body)
	if err != nil {
		return nil, NewError(StatusInternalServerError, fmt.Sprintf("failed to read body: %v (read %d of %d)", err, n, contentLength))
	}

	req.Body = io.NopCloser(bytes.NewReader(body))

	return req, nil
}

func (r Request) Cookie(key string) string {

	if c := r.Headers.Get("Cookie"); c != "" {

		cookies := r.Headers.Values("Cookie")
		for _, keyValue := range cookies {
			parts := strings.SplitN(keyValue, ":", 2)
			if len(parts) != 2 {
				continue // skping invalid
			}
			if parts[0] == key {
				return parts[1]
			}
		}

		return ""
	}

	return ""
}

func (r Request) Cookies(key string) []string {

	if c := r.Headers.Get("Cookie"); c != "" {

		var res []string

		cookies := r.Headers.Values("Cookie")
		for _, keyValue := range cookies {
			parts := strings.SplitN(keyValue, ":", 2)
			if len(parts) != 2 {
				continue // skping invalid
			}
			if parts[0] == key {
				res = append(res, parts[1])
			}
		}

		return res
	}

	return nil
}

func (r *Request) Context() context.Context {

	if r.ctx != nil {
		return r.ctx
	}

	return context.Background()
}

func (r *Request) WithContext(ctx context.Context) *Request {
	if ctx == nil {
		panic("nil context")
	}

	req2 := new(Request)

	*req2 = *r

	req2.ctx = ctx
	return req2
}
