# go-metrics

Health and Metrics web server library for Go

## Usage with example

```golang
package example

import (
	"math/rand"

	metrics "github.com/randlabs/go-metrics"
)

func main() {
	// Create a new health & metrics web server
	srvOpts := metrics.Options{
		Address:        "127.0.0.1",
		Port:           3000,
		HealthCallback: healthCallback, // Setup our health check callback
	}
	mws, err := metrics.CreateMetricsWebServer(srvOpts)
	if err != nil {
		// handle error
	}

	// Create a custom prometheus counter
	err = mws.CreateCounterWithCallback(
		"random_counter", "A random counter",
		func() float64 {
			// Return the counter value.
			// The common scenario is to have a shared set of variables you regularly update with the current
			// state of your application.
			return rand.Float64()
		},
	)

	// Start health & metrics web server
	err = mws.Start()
	if err != nil {
		// handle error
	}

	// your app code may go here

	// Stop health & metrics web server before quitting
	mws.Stop()
}

// Health output is in JSON format. Don't forget to add json tags.
type exampleHealthOutput struct {
	Status  string `json:"status"`
}

// Our health callback routine.
func healthCallback() interface{} {
	return exampleHealthOutput{
		Status: "ok",
	}
}
```

## Lincese
See `LICENSE` file for details.
