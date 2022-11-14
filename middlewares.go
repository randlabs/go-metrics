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
			if mws.rundownProt.Acquire() {
				err := next(req)
				mws.rundownProt.Release()
				return err
			} else {
				req.Error("", http.StatusServiceUnavailable)
				return nil
			}
		}
	}
}
