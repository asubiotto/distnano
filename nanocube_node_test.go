package distnano

import (
	"fmt"
	"net/http"
	"testing"
)

// TestNanocubeNodeCrimeConstruction only runs if there are 5 NanocubeNodes
// running on http://localhost:900x where x \in {0, 1, 2, 3, 4}
func TestNanocubeNodeCrimeConstruction(t *testing.T) {
	var addrs = []string{
		"http://localhost:9000",
		"http://localhost:9001",
		"http://localhost:9002",
		"http://localhost:9003",
		"http://localhost:9004",
	}

	// Do a count request to check that the servers are up and responding.
	for _, addr := range addrs {
		url := fmt.Sprintf("%v%v", addr, "/count")
		_, err := http.Get(url)
		if err != nil {
			return
		}
	}

	// No error, let's construct them.
	nodes, err := nanocubeNodesFromAddrs(addrs)
	if err != nil {
		t.Fatalf("%v %v", err, nodes)
	}

	// Check that the binSize is 3600s
	for _, node := range nodes {
		if node.tBin.binSize.Seconds() != 3600 {
			t.Fatalf("Unexpected bin size")
		}
	}

	// TODO(asubiotto): Add a check for the relativeBin
}
