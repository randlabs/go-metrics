package metrics

// -----------------------------------------------------------------------------

import (
	"crypto/tls"
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	webserver "github.com/randlabs/go-webserver"
	"github.com/randlabs/go-webserver/middleware"
	rp "github.com/randlabs/rundown-protection"
)

// -----------------------------------------------------------------------------

// Controller holds details about a metrics monitor instance.
type Controller struct {
	rundownProt         *rp.RundownProtection
	server              *webserver.Server
	usingInternalServer bool
	registry            *prometheus.Registry
	healthCallback      HealthCallback
}

// Options specifies metrics controller initialization options.
type Options struct {
	// If Server is provided, use this server instead of creating a new one.
	Server *webserver.Server

	// Address is the bind address to attach the internal web server.
	Address string

	// Port is the port number the internal web server will use.
	Port uint16

	// TLSConfig optionally provides a TLS configuration for use.
	TLSConfig *tls.Config

	// AccessToken is an optional access token required to access the status endpoints.
	AccessToken string

	// If RequestAccessTokenInHealth is enabled, access token checked also in '/health' endpoint.
	RequestAccessTokenInHealth bool

	// HealthCallback is a function that returns an object which, in turn, will be converted to JSON format.
	HealthCallback HealthCallback

	// Expose debugging profiles /debug/pprof endpoint.
	EnableDebugProfiles bool

	// Include Cache-Control headers in response to disable client-side caching.
	DisableClientCache bool

	// Include CORS headers in response.
	IncludeCORS bool

	// Middlewares additional set of middlewares for the endpoints.
	Middlewares []webserver.MiddlewareFunc
}

// HealthCallback indicates a function that returns a string that will be returned as the output.
type HealthCallback func() string

// -----------------------------------------------------------------------------

// CreateController initializes and creates a new controller
func CreateController(opts Options) (*Controller, error) {
	var err error

	if opts.HealthCallback == nil {
		return nil, errors.New("invalid health callback")
	}

	// Create metrics object
	mws := Controller{
		rundownProt:    rp.Create(),
		healthCallback: opts.HealthCallback,
	}

	// Create webserver
	if opts.Server != nil {
		mws.server = opts.Server
	} else {
		mws.usingInternalServer = true
		mws.server, err = webserver.Create(webserver.Options{
			Name:              "metrics-server",
			Address:           opts.Address,
			Port:              opts.Port,
			EnableCompression: false,
			TLSConfig:         opts.TLSConfig,
		})
		if err != nil {
			mws.Destroy()
			return nil, fmt.Errorf("unable to create metrics web server [err=%v]", err)
		}
	}

	// Create Prometheus handler
	err = mws.createPrometheusRegistry()
	if err != nil {
		mws.Destroy()
		return nil, err
	}

	// Add middlewares
	middlewares := make([]webserver.MiddlewareFunc, 0)
	if len(opts.Middlewares) > 0 {
		middlewares = append(middlewares, opts.Middlewares...)
	}
	middlewares = append(middlewares, mws.createAliveMiddleware())
	if opts.DisableClientCache {
		middlewares = append(middlewares, middleware.DisableClientCache())
	}
	if opts.IncludeCORS {
		middlewares = append(middlewares, middleware.DefaultCORS())
	}

	// Create middlewares with authorization
	middlewaresWithAuth := make([]webserver.MiddlewareFunc, len(middlewares))
	copy(middlewaresWithAuth, middlewares)
	if len(opts.AccessToken) > 0 {
		middlewaresWithAuth = append(middlewaresWithAuth, middleware.ProtectedWithToken(opts.AccessToken))
	}

	// Add health handler to web server
	if opts.RequestAccessTokenInHealth {
		mws.server.GET("/health", mws.getHealthHandler(), middlewaresWithAuth...)
	} else {
		mws.server.GET("/health", mws.getHealthHandler(), middlewares...)
	}

	// Add metrics handler to web server
	mws.server.GET("/metrics", mws.getMetricsHandler(), middlewaresWithAuth...)

	// Add debug profiles handler to web server
	if opts.EnableDebugProfiles {
		mws.server.ServeDebugProfiles("/debug/pprof", middlewaresWithAuth...)
	}

	// Done
	return &mws, nil
}

// Start starts the monitor's internal web server
func (mws *Controller) Start() error {
	if mws.server == nil {
		return errors.New("metrics monitor web server not initialized")
	}
	if !mws.usingInternalServer {
		return errors.New("cannot start an external web server")
	}
	return mws.server.Start()
}

// Destroy destroys the monitor and stops the internal web server
func (mws *Controller) Destroy() {
	// Initiate shutdown
	mws.rundownProt.Wait()

	// Cleanup
	if mws.server != nil {
		// Stop the internal web server if running
		if mws.usingInternalServer {
			mws.server.Stop()
		}
		mws.server = nil
	}
	mws.registry = nil
	mws.healthCallback = nil
}

// Registry returns the prometheus registry object
func (mws *Controller) Registry() *prometheus.Registry {
	return mws.registry
}
