package connlimit

import (
	"net/http"

	"github.com/urfave/negroni"
)

// MaxAllowed limits the number of connections by using a
// channel n connections deep.
func MaxAllowed(n int) negroni.HandlerFunc {
	sem := make(chan struct{}, n)
	acquire := func() { sem <- struct{}{} }
	release := func() { <-sem }
	return negroni.HandlerFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		acquire() // before request
		next(w, r)
		release() // after request
	})
}
