# go-metrics

Health and Metrics controller and web server library for Go

## Usage with example

```golang
package example

import (
	"math/rand"

	"github.com/randlabs/go-metrics"
)

func main() {
	// Create a new health & metrics controller with a web server
	srvOpts := metrics.Options{
		Address:             "127.0.0.1",
		Port:                3000,
		HealthCallback:      healthCallback, // Setup our health check callback
		EnableDebugProfiles: true,
		IncludeCORS:         true,
		DisableClientCache:  true,
	}
	mc, err := metrics.CreateController(srvOpts)
	if err != nil {
		// handle error
	}

	// Create a custom prometheus counter
	err = mc.CreateCounterWithCallback(
		"random_counter", "A random counter",
		func() float64 {
			// Return the counter value.
			// The common scenario is to have a shared set of variables you regularly update with the current
			// state of your application.
			return rand.Float64()
		},
	)

	// Start health & metrics web server
	err = mc.Start()
	if err != nil {
		// handle error
	}

	// your app code may go here

	// Stop health & metrics web server before quitting
	mc.Destroy()
}

// Health output will be in JSON format.
type exampleHealthOutput struct {
	Status  string `json:"status"`
}

// Our health callback routine.
func healthCallback() string {
	state := exampleHealthOutput{
		Status: "ok",
	}

	j, _ := json.Marshal(state)
	return string(j)
}
```

## License

See [LICENSE](/LICENSE) file for details.
