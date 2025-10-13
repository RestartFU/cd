package rmux

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Router struct {
	*mux.Router
}

func NewRouter() *Router {
	r := &Router{
		Router: mux.NewRouter(),
	}

	return r
}
func (r *Router) Allow(f AllowFunc) {
	r.Use(allowMiddleware(f))
}

func HandleRoute[T any](r *Router, route string, f func(T) *Response) {
	r.HandleFunc(route, wrapHandler(f))
}

func (r *Router) Start(addr string) error {
	return http.ListenAndServe(addr, r)
}
