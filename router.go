package httplib

import (
	"strings"
)

type Handler interface {
	ServeHTTP(conn Response, r *Request)
}

type HandlerFunc func(w Response, req *Request)

type route struct {
	method   string
	pattern  string
	segments []string // pre-split pattern: "/users/{id}" → ["users", "{id}"]
	handler  HandlerFunc
}

type Router struct {
	routes []route
}

func NewRouter() *Router {
	return &Router{
		routes: make([]route, 0),
	}
}

func (r *Router) GET(path string, handler HandlerFunc) {

	r.routes = append(r.routes, route{
		method:   "GET",
		pattern:  path,
		segments: strings.Split(path, "/"),
		handler:  handler,
	})
}

func (r *Router) POST(path string, handler HandlerFunc) {
	r.routes = append(r.routes, route{
		method:   "POST",
		pattern:  path,
		segments: strings.Split(path, "/"),
		handler:  handler,
	})
}

func (r *Router) PUT(path string, handler HandlerFunc) {
	r.routes = append(r.routes, route{
		method:   "PUT",
		pattern:  path,
		segments: strings.Split(path, "/"),
		handler:  handler,
	})
}

func (r *Router) PATCH(path string, handler HandlerFunc) {
	r.routes = append(r.routes, route{
		method:   "PATCH",
		pattern:  path,
		segments: strings.Split(path, "/"),
		handler:  handler,
	})
}

func (r *Router) DELETE(path string, handler HandlerFunc) {
	r.routes = append(r.routes, route{
		method:   "DELETE",
		pattern:  path,
		segments: strings.Split(path, "/"),
		handler:  handler,
	})
}

func (r *Router) ServeHTTP(w Response, req *Request) {
	resp := NewResponse()

	for _, route := range r.routes {
		if route.method != req.Method {
			continue
		}
		params, ok := match(route.segments, req.URL.Path)
		if !ok {
			continue
		}
		req.params = params
		route.handler(w, req)
		return
	}

	resp.StatusCode = StatusNotFound
	resp.ReasonPhrase = "Not Found"
	resp.Write(*w.Conn)
}

func match(routeSegments []string, requestPath string) (map[string]string, bool) {

	reqSegments := strings.Split(requestPath, "/")

	if len(routeSegments) != len(reqSegments) {
		return nil, false
	}

	params := make(map[string]string)
	for i, seg := range routeSegments {
		if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {

			key := seg[1 : len(seg)-1]
			params[key] = reqSegments[i]

		} else if seg != reqSegments[i] {

			return nil, false
		}
	}
	return params, true
}
