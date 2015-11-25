package distnano

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

// TODO(asubiotto): Validate the request.
func handler(w http.ResponseWriter, r *http.Request) {
	// Our actual response to the request.
	var response *NanocubeResponse

	// Mutex to protect concurrent modification of response and WaitGroup to
	// wait for all responses.
	var mtx = &sync.Mutex{}
	var wg = &sync.WaitGroup{}

	wg.Add(5)
	// Send off the request to each server that we know exists.
	for i := 1; i <= 5; i++ {
		go func(counter int) {
			defer wg.Done()

			// TODO(asubiotto): Was getting an annoying EOF error. Put the check
			// back in for an error.
			rawResponse, err := http.Get(
				fmt.Sprintf("http://localhost:900%v%v", counter, r.URL.Path),
			)

			if err != nil {
				return
			}

			// Read the response and unmarshal into a NanocubeResponse object.
			defer rawResponse.Body.Close()
			content, _ := ioutil.ReadAll(rawResponse.Body)

			partitionResponse := new(NanocubeResponse)
			err = json.Unmarshal(content, partitionResponse)
			if err != nil {
				log.Fatal(err)
			}

			// Update our global response.
			mtx.Lock()
			defer mtx.Unlock()
			if response == nil {
				// First one to respond, we just set the global response to be
				// our response.
				response = partitionResponse
				return
			}

			// TODO(asubiotto): We can use a more intelligent merging strategy
			// when we have a lot of nodes (in a tree-like fashion).

			// Otherwise we merge with what we have already.
			merge(response, partitionResponse)
		}(i)
	}

	wg.Wait()
	b, err := json.Marshal(response)
	if err != nil {
		log.Fatal(err)
	}

	w.Write(b)
}

func Run() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":29512", nil)
}
