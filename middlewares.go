package metrics

import (
	"net/http"

	webserver "github.com/randlabs/go-webserver"
	"github.com/randlabs/go-webserver/request"
)

// -----------------------------------------------------------------------------

func (mws *Controller) createAliveMiddleware() webserver.MiddlewareFunc {
	return func(next webserver.HandlerFunc) webserver.HandlerFunc {
		return func(req *request.RequestContext) error {
			// Process the request if we are not shutting down
			if !mws.rp.Acquire() {
				req.Error(http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
				return nil
			}
			defer mws.rp.Release()

			return next(req)
		}
	}
}
