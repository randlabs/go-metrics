package metrics

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	webserver "github.com/randlabs/go-webserver"
	"github.com/randlabs/go-webserver/request"
)

// -----------------------------------------------------------------------------

func (mws *Controller) getHealthHandler() webserver.HandlerFunc {
	return func(req *request.RequestContext) error {
		// Get current health status from callback
		status := mws.healthCallback()

		// Send output
		if len(status) > 0 {
			req.WriteString(status)
		}
		req.Success()

		// Done
		return nil
	}
}

func (mws *Controller) getMetricsHandler() webserver.HandlerFunc {
	return webserver.HandlerFromHttpHandler(promhttp.HandlerFor(
		mws.registry,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))
}
