package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

type Child struct {
	Path     []int   `json:"path,omitempty"`
	X        *int    `json:"x,omitempty"`
	Y        *int    `json:"y,omitempty"`
	Val      *int    `json:"val,omitempty"`
	Children []Child `json:"children,omitempty`
}

type Root struct {
	Val      *int    `json:"val,omitempty"`
	Children []Child `json:"children,omitempty"`
}

type NanocubeResponse struct {
	Layers []string `json:"layers"`
	Root   Root     `json:"root"`
}

// merge merges the two interfaces (supposedly json objects) passed in according
// to the API.md file in https://github.com/laurolins/nanocube.
func merge(dest, src *NanocubeResponse) {
	// Little check for layers.
	if dest.Layers == nil {
		dest.Layers = []string{}
	}

	if dest.Root.Val != nil {
		*(dest.Root.Val) += *(src.Root.Val)
		return
	}
}

func unmarshalNanocubeResponse(b []byte, dest *NanocubeResponse) {
	// TODO(asubiotto): Handle error.
	err := json.Unmarshal(b, dest)
	if err != nil {
		log.Fatal(err)
	}
}

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

			// Read the response and unmarshal into an unknown interface object.
			defer rawResponse.Body.Close()
			content, _ := ioutil.ReadAll(rawResponse.Body)

			partitionResponse := new(NanocubeResponse)
			unmarshalNanocubeResponse(content, partitionResponse)

			// Update our global response.
			mtx.Lock()
			defer mtx.Unlock()
			if response == nil {
				// First one to respond, we just set the global response to be
				// our response.
				response = partitionResponse
				return
			}

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

func main() {
	// Start a server to listen for queries and forward them.
	http.HandleFunc("/", handler)
	http.ListenAndServe(":29512", nil)
}
