package metrics

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	webserver "github.com/randlabs/go-webserver"
	"github.com/valyala/fasthttp"
)

// -----------------------------------------------------------------------------

func (mws *MetricsWebServer) getHealthHandler() fasthttp.RequestHandler {
	return func (ctx *webserver.RequestCtx) {
		webserver.EnableCORS(ctx)
		webserver.DisableCache(ctx)

		// Get current state from callback
		state := mws.healthCallback()

		// Encode and send output
		webserver.SendJSON(ctx, state)
	}
}

func (mws *MetricsWebServer) getMetricsHandler() fasthttp.RequestHandler {
	return webserver.FastHttpHandlerFromHttpHandler(promhttp.HandlerFor(
		mws.registry,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))
}
