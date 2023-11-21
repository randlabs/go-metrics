package metrics

// -----------------------------------------------------------------------------

import (
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	webserver "github.com/randlabs/go-webserver"
	"github.com/randlabs/go-webserver/middleware"
	"github.com/randlabs/rundown-protection"
)

// -----------------------------------------------------------------------------

// Controller holds details about a metrics monitor instance.
type Controller struct {
	rp                  *rundown_protection.RundownProtection
	server              *webserver.Server
	usingInternalServer bool
	registry            *prometheus.Registry
	healthCallback      HealthCallback
}

// Options specifies metrics controller initialization options.
type Options struct {
	// If Server is provided, use this server instead of creating a new one.
	Server *webserver.Server

	// Server name to use when sending response headers. Defaults to 'metrics-server'.
	Name string

	// Address is the bind address to attach the internal web server.
	Address string

	// Port is the port number the internal web server will use.
	Port uint16

	// TLSConfig optionally provides a TLS configuration for use.
	TLSConfig *tls.Config

	// A callback to call if an error is encountered.
	ListenErrorHandler webserver.ListenErrorHandler

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

const (
	defaultServerName = "metrics-server"
)

// -----------------------------------------------------------------------------

// CreateController initializes and creates a new controller
func CreateController(options Options) (*Controller, error) {
	var err error

	if options.HealthCallback == nil {
		return nil, errors.New("invalid health callback")
	}

	// Create metrics object
	mws := Controller{
		rp:             rundown_protection.Create(),
		healthCallback: options.HealthCallback,
	}

	// Create webserver
	if options.Server != nil {
		mws.server = options.Server
	} else {
		serverName := options.Name
		if len(serverName) == 0 {
			serverName = defaultServerName
		}

		mws.usingInternalServer = true
		mws.server, err = webserver.Create(webserver.Options{
			Name:               serverName,
			Address:            options.Address,
			Port:               options.Port,
			ReadTimeout:        10 * time.Second, // 10 seconds for reading a metrics request
			WriteTimeout:       time.Minute,      // and 1 minute for write
			MaxRequestBodySize: 512,              // Currently, no POST endpoints but leave a small buffer for future requests.
			EnableCompression:  false,
			ListenErrorHandler: options.ListenErrorHandler,
			TLSConfig:          options.TLSConfig,
			MinReqFileDescs:    16,
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
	if len(options.Middlewares) > 0 {
		middlewares = append(middlewares, options.Middlewares...)
	}
	middlewares = append(middlewares, mws.createAliveMiddleware())
	if options.DisableClientCache {
		middlewares = append(middlewares, middleware.DisableClientCache())
	}
	if options.IncludeCORS {
		middlewares = append(middlewares, middleware.DefaultCORS())
	}

	// Create middlewares with authorization
	middlewaresWithAuth := make([]webserver.MiddlewareFunc, len(middlewares))
	copy(middlewaresWithAuth, middlewares)
	if len(options.AccessToken) > 0 {
		middlewaresWithAuth = append(middlewaresWithAuth, middleware.ProtectedWithToken(options.AccessToken))
	}

	// Add health handler to web server
	m := middlewares
	if options.RequestAccessTokenInHealth {
		m = middlewaresWithAuth
	}
	mws.server.GET("/health", mws.getHealthHandler(), m...)
	mws.server.HEAD("/health", mws.getHealthHandler(), m...)

	// Add metrics handler to web server
	mws.server.GET("/metrics", mws.getMetricsHandler(), middlewaresWithAuth...)

	// Add debug profiles handler to web server
	if options.EnableDebugProfiles {
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
	mws.rp.Wait()

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
