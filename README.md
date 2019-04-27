<p align="center">
  <img width="460" height="300" src="https://i.imgur.com/TVwMiNN.png">
</p>

## Example
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

func main() {
	p, err := gastly.NewProvider(os.Getenv("GHOST_PROXIES_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	for range time.NewTicker(1 * time.Second).C {
		resp, err := p.Get("http://icanhazip.com", nil)
		if err != nil {
			log.Fatal(err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		defer resp.Body.Close()

		code := resp.StatusCode
		fmt.Println(fmt.Sprintf("%v%v - %v\n", string(body), code, http.StatusText(code)))
	}
}
```
