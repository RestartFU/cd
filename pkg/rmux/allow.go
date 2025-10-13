package rmux

import (
	"io"
	"net/http"
)

type AllowFunc func(*http.Request) bool

func allowMiddleware(allow AllowFunc) func(http.Handler) http.Handler {
	return func(f http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !allow(r) {
				w.WriteHeader(http.StatusForbidden)
				_, _ = io.WriteString(w, "Forbidden")
				return
			}
			f.ServeHTTP(w, r)
		})
	}
}
