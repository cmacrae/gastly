<p align="center">
  <img src="https://i.imgur.com/TVwMiNN.png">
</p>

## Features
- Automatic proxy retrieval/setup
- Automatic HTTP retries, with configurable behavior
- Prometheus metrics for requests performed by `gastly`

## Example
### Implementation
```go
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cmacrae/gastly"
)

// Serve Prometheus metrics on port 3000
func init() {
	go gastly.ServeMetrics(3000)
}

func main() {
	// Set up proxies
	p, err := gastly.NewProvider(os.Getenv("GHOST_PROXIES_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	// Configure retry behavior
	retryOptions := gastly.RetryOptions{
		Max:             3,
		WaitMaxSecs:     6,
		WaitMinSecs:     1,
		BackoffStepSecs: 2,
	}

	// For demonstration: every second, pick a random proxy from the account
	// use it to GET icanhazip.com
	for range time.NewTicker(1 * time.Second).C {
		resp, err := p.Get("http://icanhazip.com", nil, retryOptions)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		code := resp.StatusCode
		fmt.Println(fmt.Sprintf("%v%v - %v\n", string(body), code, http.StatusText(code)))
	}
}
```

### Output
```
$ ./example
123.45.678.90
200 - OK

90.123.45.678
200 - OK

45.12.90.453
200 - OK

459.12.3.45
200 - OK

90.123.45.678
200 - OK

123.45.678.90
200 - OK
```

### Metrics
```
$ curl -s localhost:3000/metrics | fgrep gastly
# HELP gastly_external_http_requests_total How many external HTTP requests processed, partitioned by status code and HTTP method.
# TYPE gastly_external_http_requests_total counter
gastly_external_http_requests_total{code="200",method="GET"} 6
```
