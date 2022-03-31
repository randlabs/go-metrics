package metrics

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	webserver "github.com/randlabs/go-webserver"
	"github.com/randlabs/go-webserver/request"
)

// -----------------------------------------------------------------------------

func (mws *MetricsWebServer) getHealthHandler() webserver.HandlerFunc {
	return func(req *request.RequestContext) error {
		// Get current state from callback
		state := mws.healthCallback()

		// Encode and send output
		req.WriteJSON(state)
		req.Success()

		// Done
		return nil
	}
}

func (mws *MetricsWebServer) getMetricsHandler() webserver.HandlerFunc {
	return webserver.HandlerFromHttpHandler(promhttp.HandlerFor(
		mws.registry,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))
}
