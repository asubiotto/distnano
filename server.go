package distnano

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
)

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	log.Println("Handling request for url path: ", r.URL.Path)

	// Our actual response to the request.
	var response JSONResponse

	// Mutex to protect concurrent modification of response and WaitGroup to
	// wait for all responses.
	var mtx = &sync.Mutex{}
	var wg = &sync.WaitGroup{}

	// Note what kind of request this is.
	schemaRequest := strings.HasPrefix(r.URL.Path, "/schema")
	wg.Add(5)
	// Send off the request to each server that we know exists.
	for i := 1; i <= 5; i++ {
		go func(counter int) {
			defer wg.Done()

			rawResponse, err := http.Get(
				fmt.Sprintf("http://localhost:900%v%v", counter, r.URL.Path),
			)

			if err != nil {
				log.Fatalf(
					"Cannot continue, error getting %v from node %v\n",
					r.URL.Path,
					i,
				)
			}

			// Read the response and unmarshal into a NanocubeResponse object.
			defer rawResponse.Body.Close()
			content, _ := ioutil.ReadAll(rawResponse.Body)

			var partitionResponse JSONResponse
			if schemaRequest {
				partitionResponse = new(SchemaResponse)
			} else {
				partitionResponse = new(NanocubeResponse)
			}

			err = partitionResponse.Unmarshal(content)
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
			response.Merge(partitionResponse)
		}(i)
	}

	wg.Wait()
	b, err := response.Marshal()
	if err != nil {
		log.Fatal(err)
	}

	w.Write(b)
}

func Run() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":29512", nil))
}
