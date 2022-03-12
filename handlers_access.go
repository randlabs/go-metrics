package metrics

import (
	"crypto/subtle"
	"strings"

	webserver "github.com/randlabs/go-webserver"
	"github.com/valyala/fasthttp"
)

// -----------------------------------------------------------------------------
// Private methods

func (mws *MetricsWebServer) checkAccessToken(ctx *webserver.RequestCtx) bool {
	if len(mws.accessToken) == 0 {
		return true
	}

	var token []byte

	// Get X-Access-Token header
	header := ctx.Request.Header.Peek("X-Access-Token")
	if len(header) > 0 {
		token = header
	} else {
		// If no token provided, try with Authorization: Bearer XXX header
		header = ctx.Request.Header.Peek("Authorization")
		if len(header) > 0 {
			auth := strings.SplitN(string(header), " ", 2)
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

func (mws *MetricsWebServer) protectedHandler(handler fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		// Check access token
		if mws.checkAccessToken(ctx) {
			handler(ctx)
		} else {
			webserver.EnableCORS(ctx)
			webserver.DisableCache(ctx)
			webserver.SendAccessDenied(ctx, "403 forbidden")
		}
	}
}
