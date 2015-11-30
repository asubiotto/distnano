package distnano

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// addrs is a string slice that will hold all the addrs of our child nodes.
var addrs []string

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	log.Println("Handling request for url path:", r.URL.Path)

	// Our actual response to the request.
	var response JSONResponse

	// This global error variable will indicate whether our goroutines
	// encountered any errors.
	var glerr error
	var glerrMtx = &sync.Mutex{}

	// Mutex to protect concurrent modification of response and WaitGroup to
	// wait for all responses.
	var mtx = &sync.Mutex{}
	var wg = &sync.WaitGroup{}

	// Note what kind of request this is.
	schemaRequest := strings.HasPrefix(r.URL.Path, "/schema")
	wg.Add(len(addrs))
	// Send off the request to each server that we know exists.
	for _, e := range addrs {
		go func(addr string) {
			defer wg.Done()

			url := fmt.Sprintf("%v%v", addr, r.URL.Path)

			rawResponse, err := http.Get(url)
			if err != nil {
				glerrMtx.Lock()
				defer glerrMtx.Unlock()
				glerr = err
				return
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
				glerrMtx.Lock()
				defer glerrMtx.Unlock()
				glerr = err
				return
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
		}(e)
	}

	wg.Wait()
	if response == nil || glerr != nil {
		w.Write([]byte("error"))
		return
	}

	b, err := response.Marshal()
	if err != nil {
		log.Fatal(err)
	}

	w.Write(b)
}

func Run(port int, addresses []string) {
	addrs = addresses
	http.HandleFunc("/", handler)
	log.Println("Starting server on port", port)
	go func() {
		<-time.After(2 * time.Second)
		log.Println("Server started")
	}()
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}
