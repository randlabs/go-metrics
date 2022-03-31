package metrics

// -----------------------------------------------------------------------------

import (
	"crypto/tls"
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	webserver "github.com/randlabs/go-webserver"
	"github.com/randlabs/go-webserver/middleware"
)

// -----------------------------------------------------------------------------

// MetricsWebServer is the metrics web server object
type MetricsWebServer struct {
	httpsrv        *webserver.Server
	registry       *prometheus.Registry
	healthCallback HealthCallback
	accessToken    []byte
}

// Options specifies metrics web server initialization options.
type Options struct {
	// Address is the bind address to attach the server listener.
	Address string

	// Port is the port number the server will listen.
	Port uint16

	// TLSConfig optionally provides a TLS configuration for use.
	TLSConfig *tls.Config

	// AccessToken is an optional access token required to access the status endpoints.
	AccessToken string

	// HealthCallback indicates a function that returns an object that will be returned as a JSON output.
	HealthCallback HealthCallback

	// Include Cache-Control headers in response.
	IncludeNoCache bool

	// Include CORS headers in response.
	IncludeCORS    bool
}

// HealthCallback indicates a function that returns an object that will be returned as a JSON output.
type HealthCallback func() interface{}

// -----------------------------------------------------------------------------

// CreateMetricsWebServer initializes and creates a new web server
func CreateMetricsWebServer(opts Options) (*MetricsWebServer, error) {
	var err error

	if opts.HealthCallback == nil {
		return nil, errors.New("invalid health callback")
	}

	// Create metrics object
	mws := MetricsWebServer{
		healthCallback: opts.HealthCallback,
		accessToken:    []byte(opts.AccessToken),
	}

	// Create webserver
	mws.httpsrv, err = webserver.Create(webserver.Options{
		Name:              "metrics-server",
		Address:           opts.Address,
		Port:              opts.Port,
		EnableCompression: false,
		TLSConfig:         opts.TLSConfig,
	})
	if err != nil {
		mws.Stop()
		return nil, fmt.Errorf("unable to create metrics web server [err=%v]", err)
	}

	// Create Prometheus handler
	err = mws.createPrometheusRegistry()
	if err != nil {
		mws.Stop()
		return nil, err
	}

	// Add middlewares
	if opts.IncludeNoCache {
		mws.httpsrv.Use(middleware.DisableCacheControl())
	}
	if opts.IncludeCORS {
		mws.httpsrv.Use(middleware.DefaultCORS())
	}

	// Add webserver handlers
	mws.httpsrv.GET("/health", mws.protectedHandler(mws.getHealthHandler()))
	mws.httpsrv.GET("/metrics", mws.protectedHandler(mws.getMetricsHandler()))
	mws.httpsrv.ServeDebugProfiler("/debug/pprof", mws.checkAccessToken)

	// Done
	return &mws, nil
}

// Start starts the metrics web server
func (mws *MetricsWebServer) Start() error {
	if mws.httpsrv == nil {
		return errors.New("metrics webserver not initialized")
	}
	return mws.httpsrv.Start()
}

// Stop shuts down the metrics web server
func (mws *MetricsWebServer) Stop() {
	// Stop web server
	if mws.httpsrv != nil {
		mws.httpsrv.Stop()
		mws.httpsrv = nil
	}
	mws.registry = nil

	mws.cleanAccessToken()
	mws.healthCallback = nil
}

// Registry returns the prometheus registry object
func (mws *MetricsWebServer) Registry() *prometheus.Registry {
	return mws.registry
}

// -----------------------------------------------------------------------------

func (mws *MetricsWebServer) cleanAccessToken() {
	tokenLen := len(mws.accessToken)
	for idx := 0; idx < tokenLen; idx++ {
		mws.accessToken[idx] = 0
	}
}
