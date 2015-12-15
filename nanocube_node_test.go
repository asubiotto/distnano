package distnano

import (
	"reflect"
	"testing"
)

var absToRel = []struct {
	query        string
	result       string
	outsideRange bool
	spt          *SpecialTimeQuery
	relativeBin  int
}{
	// Tests a query that falls fully into a NanocubeNode's range.
	{
		"/count.r(\"time\",mt_interval_sequence(480,24,10))",
		"/count.r(\"time\",mt_interval_sequence(36,24,10))",
		false,
		nil,
		444,
	},
	// Tests an interval query (simpler of the two time queries).
	{
		"/count.r(\"time\",interval(480,720))",
		"/count.r(\"time\",interval(36,276))",
		false,
		nil,
		444,
	},
	// Tests a query that falls only partially into a NanocubeNode's range and
	// would need to be readjusted according to bucketOffset.
	{
		"/count.r(\"time\",mt_interval_sequence(408,24,13))",
		"",
		false,
		&SpecialTimeQuery{
			queryOne:     "/count.r(\"time\",mt_interval_sequence(0,12,1))",
			queryTwo:     "/count.r(\"time\",mt_interval_sequence(12,24,11))",
			bucketOffset: 1,
			node:         nil,
		},
		444,
	},
	// Tests a time query over the whole dataset for a variety of relativeBins.
	{
		"/count.r(\"time\",mt_interval_sequence(0,524289,8192))",
		"/count.r(\"time\",mt_interval_sequence(0,524289,8192))",
		false,
		nil,
		0,
	},
	{
		"/count.r(\"time\",mt_interval_sequence(0,524289,8192))",
		"",
		false,
		&SpecialTimeQuery{
			queryOne:     "/count.r(\"time\",mt_interval_sequence(0,523545,1))",
			queryTwo:     "/count.r(\"time\",mt_interval_sequence(523545,524289,8191))",
			bucketOffset: 0,
			node:         nil,
		},
		744,
	},
	{
		"/count.r(\"time\",mt_interval_sequence(480,24,10))",
		"/count.r(\"time\",mt_interval_sequence(0,24,-32))",
		true,
		nil,
		1488,
	},
}

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
	for _, e := range absToRel {
		node := &NanocubeNode{relativeBin: e.relativeBin}
		result, outsideRange, spTimeQuery := node.mustAbsToRelTimeQuery(e.query)

		if spTimeQuery != nil {
			spTimeQuery.node = nil
		}

		if result != e.result {
			t.Fatalf(
				"Expected %v, got %v for query %v\n",
				e.result,
				result,
				e,
			)
		} else if outsideRange != e.outsideRange {
			t.Fatalf(
				"Expected %v, got %v for query %v\n",
				e.outsideRange,
				outsideRange,
				e,
			)
		} else if !reflect.DeepEqual(spTimeQuery, e.spt) {
			t.Fatalf(
				"Expected %v, got %v for query %v\n",
				e.spt,
				spTimeQuery,
				e,
			)
		}
	}
}
