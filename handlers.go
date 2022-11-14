package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	webserver "github.com/randlabs/go-webserver"
	"github.com/randlabs/go-webserver/request"
)

// -----------------------------------------------------------------------------

const (
	strContentTypeTextPlain       = "text/plain"
	strContentTypeApplicationJSON = "application/json"
)

func (mws *Controller) getHealthHandler() webserver.HandlerFunc {
	return func(req *request.RequestContext) error {
		// Get current health status from callback
		status := mws.healthCallback()

		// Send output
		if isJSON(status) {
			req.SetResponseHeader("Content-Type", strContentTypeApplicationJSON)
		} else {
			req.SetResponseHeader("Content-Type", strContentTypeTextPlain)
		}

		if !req.IsHead() {
			_, _ = req.WriteString(status)
		} else {
			req.SetResponseHeader("Content-Length", strconv.FormatUint(uint64(int64(len(status))), 10))
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

// -----------------------------------------------------------------------------

func isJSON(s string) bool {
	// An official (?) method but a plain text is also considered a valid JSON
	// var js interface{}
	// return json.Unmarshal([]byte(s), &js) == nil

	// Our quick approach
	startIdx := 0
	endIdx := len(s)

	// Skip blanks at the beginning and the end
	for startIdx < endIdx && isBlank(s[startIdx]) {
		startIdx += 1
	}
	for endIdx > startIdx && isBlank(s[endIdx-1]) {
		endIdx -= 1
	}

	return startIdx < endIdx &&
		((s[startIdx] == '{' && s[endIdx-1] == '}') ||
			(s[startIdx] == '[' && s[endIdx-1] == ']'))
}

func isBlank(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n'
}
