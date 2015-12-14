package distnano

import "testing"

var addrs = [5]string{
	"http://localhost:9000",
	"http://localhost:9001",
	"http://localhost:9002",
	"http://localhost:9003",
	"http://localhost:9004",
}

// TestNanocubeNodeCrimeConstruction only runs if there are 5 NanocubeNodes
// running on http://localhost:900x where x \in {0, 1, 2, 3, 4}
var TestNanocubeNodeCrimeConstruction(t *testing.T) {
	// Do a count request to check that the servers are up and responding.
	for _, addr := range addrs {
			url := fmt.Sprintf("%v%v", addr, "/count")
	}
}
