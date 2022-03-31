package metrics

import (
	"crypto/subtle"
	"strings"

	webserver "github.com/randlabs/go-webserver"
	"github.com/randlabs/go-webserver/request"
)

// -----------------------------------------------------------------------------
// Private methods

func (mws *MetricsWebServer) checkAccessToken(req *request.RequestContext) bool {
	if len(mws.accessToken) == 0 {
		return true
	}

	var token []byte

	// Get X-Access-Token header
	header := req.RequestHeader("X-Access-Token")
	if len(header) > 0 {
		token = []byte(header)
	} else {
		// If no token provided, try with Authorization: Bearer XXX header
		header = req.RequestHeader("Authorization")
		if len(header) > 0 {
			auth := strings.SplitN(header, " ", 2)
			if len(auth) == 2 && strings.EqualFold("Bearer", auth[0]) {
				token = []byte(auth[1])
			}
		}
	}

	//Check token
	if len(token) > 0 && subtle.ConstantTimeCompare(mws.accessToken, token) != 0 {
		return true
	}

	// Deny access
	return false
}

func (mws *MetricsWebServer) protectedHandler(handler webserver.HandlerFunc) webserver.HandlerFunc {
	return func(req *request.RequestContext) error {
		// Check access token
		if !mws.checkAccessToken(req) {
			req.AccessDenied("403 forbidden")
			return nil
		}
		return handler(req)
	}
}
