package distnano

import (
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var nodes []*NanocubeNode

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

	wg.Add(len(nodes))
	// Send off the request to each server that we know exists.
	for _, v := range nodes {
		go func(node *NanocubeNode) {
			defer wg.Done()

			partitionResponse, err := node.Query(r.URL.Path)
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
		}(v)
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
	var err error
	nodes, err = nanocubeNodesFromAddrs(addresses)
	if err != nil {
		log.Fatalf(
			"Could not create nanocube nodes from provided addresses: %v\n",
			addresses,
		)
	}

	http.HandleFunc("/", handler)
	log.Println("Starting server on port", port)
	go func() {
		<-time.After(2 * time.Second)
		log.Println("Server started")
	}()
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}
