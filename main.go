package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

// merge merges the two interfaces (supposedly json objects) passed in according
// to the API.md file in https://github.com/laurolins/nanocube.
func merge(dest, src interface{}) {

	for k, v := range a {

	}
}

// TODO(asubiotto): Validate the request.
func handler(w http.ResponseWriter, r *http.Request) {
	// Our actual response to the request.
	var response interface{}

	// Mutex to protect concurrent modification of response and WaitGroup to
	// wait for all responses.
	var mtx = &sync.Mutex{}
	var wg = &sync.WaitGroup{}

	wg.Add(5)
	// Send off the request to each server that we know exists.
	for i := 1; i <= 5; i++ {
		go func(counter int) {
			defer wg.Done()
			rawResponse, err := http.Get(
				fmt.Sprintf("http://localhost:900%v%v", counter, r.URL.Path),
			)
			if err != nil {
				log.Fatal(err)
			}

			// Read the response and unmarshal into an unknown interface object.
			defer rawResponse.Body.Close()
			content, _ := ioutil.ReadAll(rawResponse.Body)

			var jsonResponse interface{}
			// TODO(asubiotto): Handle error.
			json.Unmarshal(content, &jsonResponse)

			// Update our global response.
			mtx.Lock()
			defer mtx.Unlock()
			if response == nil {
				// First one to respond, we just set the global response to be
				// our response.
				response = jsonResponse
				return
			}

			// Otherwise we merge with what we have already.
			response = merge(&response, jsonResponse)
			mtx.Unlock()
		}(i)
	}

	wg.Wait()
	b, err := json.Marshal(response)
	if err != nil {
		log.Fatal(err)
	}

	w.Write(b)
}

func main() {
	// Start a server to listen for queries and forward them.
	http.HandleFunc("/", handler)
	http.ListenAndServe(":29512", nil)
}
