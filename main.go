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
	Path     []byte  `json:"path,omitempty"`
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

// TODO(asubiotto): The whole error handling has to be looked at more closely.

// mergeChildren merges the children array by aggregating "val"s at the base
// level and using either path as key or x and y as keys.
func mergeChildren(dest, src []Child) {
	if dest == nil || src == nil {
		if dest == nil {
			dest = src
		}
		return
	}

	destMap := make(map[string]int)
	for _, child := range dest {
		// If there's a path there's a value.
		if child.Val == nil {
			log.Fatal("Child value was nil in dest")
		}
		destMap[string(child.Path)] = *(child.Val)
	}

	for _, child := range src {
		if child.Val == nil {
			log.Fatal("Child value was nil in src")
		}

		if _, e := destMap[string(child.Path)]; e {
			// If it exists, we add the value that src has for this key.
			destMap[string(child.Path)] += *(child.Val)
		} else {
			// Otherwise we add it.
			destMap[string(child.Path)] = *(child.Val)
		}
	}

	// Now that we have the results in destMap, we are going to update dest's
	// children slice.
	for _, child := range dest {
		key := string(child.Path)
		*(child.Val) = destMap[key]
		delete(destMap, key)
	}

	if len(destMap) == 0 {
		return
	}

	// The keys that are left after this are paths that were not in dest to
	// begin with.
	newDest := make([]Child, len(dest)+len(destMap))
	copy(newDest, dest)
	i := len(dest)
	for k, v := range destMap {
		newDest[i] = Child{Path: []byte(k), Val: &v}
		i++
	}
	dest = newDest
}

// merge merges src into dest according to the API.md document found in
// github.com/laurolins/nanocube.
func merge(dest, src *NanocubeResponse) {
	// Little check for layers. TODO(asubiotto): Think about this case a bit
	// more when we have empty responses or we want to return an error.
	if dest.Layers == nil {
		dest.Layers = []string{}
	}

	if dest.Root.Val != nil {
		*(dest.Root.Val) += *(src.Root.Val)
		return
	}

	mergeChildren(dest.Root.Children, src.Root.Children)
}

func unmarshalNanocubeResponse(b []byte, dest *NanocubeResponse) {
	// TODO(asubiotto): Handle error. Do we really need this to be a separate
	// function?
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

			// Read the response and unmarshal into a NanocubeResponse object.
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

func main() {
	// Start a server to listen for queries and forward them.
	http.HandleFunc("/", handler)
	http.ListenAndServe(":29512", nil)
}
