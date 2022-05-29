package metrics_test

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/randlabs/go-metrics"
)

// -----------------------------------------------------------------------------

type State struct {
	System string `json:"system"`
}

// -----------------------------------------------------------------------------

func TestWebServer(t *testing.T) {
	// Create a new health & metrics controller with a web server
	srvOpts := metrics.Options{
		Address:             "127.0.0.1",
		Port:                3000,
		HealthCallback:      healthCallback,
		EnableDebugProfiles: true,
		IncludeCORS:         true,
		DisableClientCache:  true,
	}
	mc, err := metrics.CreateController(srvOpts)
	if err != nil {
		t.Errorf("unable to create web server [%v]", err)
		return
	}

	// Create some custom counters
	err = mc.CreateCounterWithCallback(
		"random_counter", "A random counter",
		func() float64 {
			return rand.Float64()
		},
	)
	if err == nil {
		err = mc.CreateCounterVecWithCallback(
			"random_counter_vec", "A random counter vector", []string{"set", "value"},
			metrics.VectorMetric{
				{
					Values: []string{"Set A", "Value 1"},
					Handler: func() float64 {
						return rand.Float64()
					},
				},
				{
					Values: []string{"Set A", "Value 2"},
					Handler: func() float64 {
						return rand.Float64()
					},
				},
				{
					Values: []string{"Set A", "Value 3"},
					Handler: func() float64 {
						return rand.Float64()
					},
				},
			},
		)
	}

	if err == nil {
		err = mc.CreateGaugeVecWithCallback(
			"random_gauge_vec", "A random gauge vector", []string{"set", "value"},
			metrics.VectorMetric{
				{
					Values: []string{"Set A", "Value 1"},
					Handler: func() float64 {
						return rand.Float64()
					},
				},
				{
					Values: []string{"Set A", "Value 2"},
					Handler: func() float64 {
						return rand.Float64()
					},
				},
				{
					Values: []string{"Set A", "Value 3"},
					Handler: func() float64 {
						return rand.Float64()
					},
				},
			},
		)
	}
	if err != nil {
		t.Errorf("unable to create metric handlers [%v]", err)
	}

	// Start server
	err = mc.Start()
	if err != nil {
		t.Errorf("unable to start web server [%v]", err)
		return
	}

	// Open default browser
	openBrowser("http://" + srvOpts.Address + ":" + strconv.Itoa(int(srvOpts.Port)) + "/metrics")

	// Wait for CTRL+C or timeout
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	select {
	case <-c:
		// NOTE: By default, tests cannot last more than 10 minutes.
	case <-time.After(5 * time.Minute):
	}

	fmt.Println("Shutting down...")

	// Stop web server
	mc.Destroy()
}

// -----------------------------------------------------------------------------

func healthCallback() interface{} {
	return State{
		System: "all services running",
	}
}

func openBrowser(url string) {
	switch runtime.GOOS {
	case "linux":
		_ = exec.Command("xdg-open", url).Start()
	case "windows":
		_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		_ = exec.Command("open", url).Start()
	}
}
