package router

import (
	"net/http"
	"strings"

	gmux "github.com/gorilla/mux"
	"github.com/alokic/gopkg/httputils"
	"github.com/alokic/gopkg/tracer"
	gorillatrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gorilla/mux"
	httptrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"
)

// Gorilla router struct
type Gorilla struct {
	*httputils.Gorilla
	mux *gorillatrace.Router
}

// NetHTTP router struct
type NetHTTP struct {
	*httputils.NetHTTP
	mux *httptrace.ServeMux
}

// CreateRouter factory to create tracing enabled router.
func CreateRouter(name, serviceName string, opts ...tracer.StartSpanOption) httputils.Router {
	switch strings.ToLower(name) {
	case "gorilla":
		return newGorilla(serviceName, opts...)
	case "http":
		return newNetHTTP(serviceName, opts...)
	default:
		return nil
	}
}

func newGorilla(serviceName string, opts ...tracer.StartSpanOption) *Gorilla {
	mux := gorillatrace.NewRouter(
		[]gorillatrace.RouterOption{gorillatrace.WithServiceName(serviceName), gorillatrace.WithSpanOptions(opts...)}...,
	)

	g := &Gorilla{Gorilla: httputils.NewGorilla(mux), mux: mux}
	g.Gorilla.SetMuxFn(func() *gmux.Router {
		return g.mux.Router
	})
	return g
}

func newNetHTTP(serviceName string, opts ...tracer.StartSpanOption) *NetHTTP {
	mux := httptrace.NewServeMux(
		[]httptrace.MuxOption{httptrace.WithServiceName(serviceName)}...,
	)

	g := &NetHTTP{NetHTTP: httputils.NewNetHTTP(mux), mux: mux}
	g.NetHTTP.SetMuxFn(func() *http.ServeMux {
		return g.mux.ServeMux
	})
	return &NetHTTP{NetHTTP: httputils.NewNetHTTP(mux), mux: mux}
}
