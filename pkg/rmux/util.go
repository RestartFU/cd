package rmux

import (
	"io"
	"net/http"

	"github.com/restartfu/cd/pkg/restutil"
)

func wrapHandler[T any](f func(T) *Response) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		q, err := restutil.ExtractHTTPQuery[T](r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		resp := f(q)
		if resp != nil {
			w.WriteHeader(resp.StatusCode)
			_, _ = io.WriteString(w, resp.Message)
			return
		}
	}
}
