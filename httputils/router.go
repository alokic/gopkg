package httputils

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	gmux "github.com/gorilla/mux"
)

var (
	pathParamsRegexp = regexp.MustCompile(`:[\w]+`)
	// ErrNotFound is returned when no route match is found.
	ErrNotFound = errors.New("no matching route was found")
)

type formatConv func(string) string

// ErrorHandler interface for router
type ErrorHandler interface {
	NotFound(http.Handler)
}

// Router interface
type Router interface {
	ErrorHandler
	http.Handler
	Handle(string, string, http.Handler)
	Name() string
}

// CreateRouter is factory for creating various router objects
func CreateRouter(routerName string) Router {
	switch strings.ToLower(routerName) {
	case "gorilla":
		return NewGorilla(gmux.NewRouter())
	case "http":
		return NewNetHTTP(http.NewServeMux())
	default:
		return nil
	}
}

// FormatPath formats a path based on router specific format
func FormatPath(routerName, path string) string {
	var fc formatConv

	switch routerName {
	case "gorilla":
		fc = toGorillaPath
	}

	if fc == nil {
		return ""
	}

	matches := pathParamsRegexp.FindAllString(path, -1)

	for _, m := range matches {
		path = strings.Replace(path, m, fc(m), -1)
	}

	return path
}

// Group is router under a prefix
type Group struct {
	router Router
	prefix string
}

// NetHTTP router.
type NetHTTP struct {
	mux             http.Handler
	notFoundHandler http.Handler
	muxFn           func() *http.ServeMux
}

// NewNetHTTP constructor for net http mux
func NewNetHTTP(mux http.Handler) *NetHTTP {
	n := &NetHTTP{mux: mux}
	n.muxFn = func() *http.ServeMux {
		return n.mux.(*http.ServeMux)
	}
	return n
}

// ServeHTTP - yay proxy is http.Handler
func (c *NetHTTP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.mux.ServeHTTP(w, r)
}

// Handle a path with method
func (c *NetHTTP) Handle(method, path string, handler http.Handler) {
	allowed := []string{"head", "get", "post", "put", "patch", "delete", "options"}
	m := strings.ToLower(method)

	for _, am := range allowed {
		if m != am {
			continue
		}
		c.muxFn().Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.ToLower(r.Method) != method {
				if c.notFoundHandler == nil {
					WriteError(ErrNotFound, http.StatusNotFound, w)
					return
				}
				handler = c.notFoundHandler
			}
			handler.ServeHTTP(w, r)
		}))
		return
	}
}

// NotFound handler for Gorilla.
func (c *NetHTTP) NotFound(handler http.Handler) {
	c.notFoundHandler = handler
}

// Name returns name of router.
func (c *NetHTTP) Name() string {
	return "http"
}

// SetMuxFn set a func which return mux ... it helps in returning actaul mux from embedded types.
func (c *NetHTTP) SetMuxFn(fn func() *http.ServeMux) {
	c.muxFn = fn
}

// Gorilla router.
type Gorilla struct {
	mux   http.Handler
	muxFn func() *gmux.Router
}

// NewGorilla constructor for Gorilla mux
func NewGorilla(mux http.Handler) *Gorilla {
	g := &Gorilla{mux: mux}
	g.muxFn = func() *gmux.Router {
		return g.mux.(*gmux.Router)
	}
	return g
}

// ServeHTTP - yay proxy is http.Handler
func (c *Gorilla) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.mux.ServeHTTP(w, r)
}

// Handle a path with method
func (c *Gorilla) Handle(method, path string, handler http.Handler) {
	allowed := []string{"head", "get", "post", "put", "patch", "delete", "options"}
	m := strings.ToLower(method)
	for _, am := range allowed {
		if m == am {
			c.muxFn().Handle(path, handler).Methods(m)
			return
		}
	}
}

// NotFound handler for Gorilla.
func (c *Gorilla) NotFound(handler http.Handler) {
	c.muxFn().NotFoundHandler = handler
}

// Name returns name of router.
func (c *Gorilla) Name() string {
	return "gorilla"
}

// SetMuxFn set a func which return mux ... it helps in returning actaul mux from ebbedded types.
func (c *Gorilla) SetMuxFn(fn func() *gmux.Router) {
	c.muxFn = fn
}

func toGorillaPath(p string) string {
	if p[0] != ':' {
		return p
	}
	return "{" + p[1:] + "}"
}
