package httplib

import "net"

type Handler interface {
	ServeHTTP(conn net.Conn, 	r *Request)
}

type HandlerFunc func(w net.Conn, req *Request)

type Router struct {
	routes map[string]map[string]HandlerFunc
}

func NewRouter() *Router {
	return &Router{
		routes: map[string]map[string]HandlerFunc{
			"GET":    make(map[string]HandlerFunc),
			"POST":   make(map[string]HandlerFunc),
			"PUT":    make(map[string]HandlerFunc),
			"PATCH":  make(map[string]HandlerFunc),
			"DELETE": make(map[string]HandlerFunc),
		},
	}
}

func (r *Router) GET(path string, handler HandlerFunc) {
	r.routes["GET"][path] = handler
}

func (r *Router) POST(path string, handler HandlerFunc) {
	r.routes["POST"][path] = handler
}

func (r *Router) PUT(path string, handler HandlerFunc) {
	r.routes["PUT"][path] = handler
}

func (r *Router) PATCH(path string, handler HandlerFunc) {
	r.routes["PATCH"][path] = handler
}

func (r *Router) DELETE(path string, handler HandlerFunc) {
	r.routes["DElETE"][path] = handler
}

func (r *Router) ServeHTTP(conn net.Conn, req *Request) {
	resp := NewResponse()

	methodRoutes, exists := r.routes[req.Method]
	if !exists {
		resp.StatusCode = StatusMethodNotAllowed
		resp.ReasonPhrase = "Method Not Allowed"
		resp.Write(conn)
		return
	}

	handler, exists := methodRoutes[req.URL.Path]
	if !exists {
		resp.StatusCode = StatusNotFound
		resp.ReasonPhrase = "Not Found"
		resp.Write(conn)
		return
	}

	handler(conn, req)
}
