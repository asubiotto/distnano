package distnano

import "testing"

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
		_, err := (&NanocubeNode{addr: addr}).Query("/schema")
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

func TestAbsoluteToRelativeTimeQuery(t *testing.T) {
	node := &NanocubeNode{relativeBin: 444}
	result, _, _ := node.mustAbsToRelTimeQuery(
		"/count.r(\"time\",mt_interval_sequence(480,24,10))",
	)

	if result != "/count.r(\"time\",mt_interval_sequence(36,24,10))" {
		t.Fatal("Unexcepted result", result)
	}

	result, _, _ = node.mustAbsToRelTimeQuery(
		"/count.r(\"time\",interval(480,720))",
	)

	if result != "/count.r(\"time\",interval(36,276))" {
		t.Fatal("Unexcepted result", result)
	}

	result, _, bucketOffset := node.mustAbsToRelTimeQuery(
		"/count.r(\"time\",mt_interval_sequence(408,24,13))",
	)

	if result != "/count.r(\"time\",mt_interval_sequence(0,24,10))" {
		t.Fatal("Unexpected result", result)
	} else if bucketOffset != 3 {
		t.Fatal("Unexpected bucket offset", bucketOffset)
	}
}
