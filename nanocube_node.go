package distnano

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type TBin struct {
	startTime time.Time
	binSize   time.Duration
}

// ByTime implements sort.Interface for []*TBin based on the TBin's startTime.
type ByTime []*TBin

type NanocubeNode struct {
	addr        string
	tBin        *TBin
	relativeBin int
}

type SpecialTimeQuery struct {
	queryOne     string
	queryTwo     string
	bucketOffset int
	node         *NanocubeNode
}

func (t ByTime) Len() int           { return len(t) }
func (t ByTime) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t ByTime) Less(i, j int) bool { return t[i].startTime.Before(t[j].startTime) }

// nanocubeNodesFromAddrs converts addrs to NanocubeNodes by querying the
// schema of each address and creating NanocubeNodes accordingly.
func nanocubeNodesFromAddrs(addrs []string) ([]*NanocubeNode, error) {
	// Get each schema and the tbin string to determine each node's notion of
	// time.
	tbins := make([]*TBin, len(addrs), len(addrs))
	for i, addr := range addrs {
		response, err := (&NanocubeNode{addr: addr}).Query("/schema")
		if err != nil {
			return nil, err
		}

		for _, md := range (response.(*SchemaResponse)).Metadata {
			if md.Key == "tbin" {
				tbin := newTBin(md.Value)
				if tbin == nil {
					return nil, errors.New("error creating time bin")
				}
				tbins[i] = tbin
			}
		}
	}

	sort.Sort(ByTime(tbins))

	nanocubeNodes := make([]*NanocubeNode, len(addrs), len(addrs))
	var firstNode *NanocubeNode = nil
	for i, addr := range addrs {
		nanocubeNodes[i] = newNanocubeNode(addr, tbins[i])
		if firstNode == nil {
			firstNode = nanocubeNodes[i]
		}

		// Get number of bins since the firstNode's startTime.
		nanocubeNodes[i].relativeBin =
			int(nanocubeNodes[i].tBin.startTime.Sub(
				firstNode.tBin.startTime,
			).Hours() / firstNode.tBin.binSize.Hours())
	}

	return nanocubeNodes, nil
}

// splitTBString returns the time (start) in tbstring and the Duration that
// specifies the size of a time bin.
func newTBin(tbstring string) *TBin {
	// layout is needed by the time module to interpret the format of our time
	// strings.
	layout := "2006-01-02_15:04:05"
	splitString := strings.Split(tbstring, "_")
	if len(splitString) != 3 {
		return nil
	}

	t, err := time.Parse(
		layout,
		fmt.Sprintf("%v_%v", splitString[0], splitString[1]),
	)

	if err != nil {
		return nil
	}

	d, err := time.ParseDuration(splitString[2])
	if err != nil {
		return nil
	}

	tbin := new(TBin)
	tbin.startTime = t
	tbin.binSize = d
	return tbin
}

func newNanocubeNode(addr string, tBin *TBin) *NanocubeNode {
	nn := new(NanocubeNode)
	nn.addr = addr
	nn.tBin = tBin
	return nn
}

// Query does some preprocessing of the url before querying the NanocubeNode.
func (n *NanocubeNode) Query(url string) (JSONResponse, error) {
	var spTimeQuery *SpecialTimeQuery

	// If this is a time query, we have to convert it from an absolute to a
	// relative time.
	if strings.Contains(url, "interval") {
		var queryOutsideRange bool
		url, queryOutsideRange, spTimeQuery = n.mustAbsToRelTimeQuery(url)
		if queryOutsideRange {
			return new(NanocubeResponse), nil
		}
	}

	if spTimeQuery != nil {
		return spTimeQuery.query()
	}

	return n.query(url)
}

// query queries both urls of SpecialTimeQuery and merges them together to
// return one NanocubeResponse.
func (s *SpecialTimeQuery) query() (*NanocubeResponse, error) {
	// Query the second query first. This is the query that has the same
	// resolution as the global query and fits nicely into the global buckets.
	response, err := s.node.query(s.queryTwo)
	if err != nil {
		return nil, err
	}

	responseTwo := response.(*NanocubeResponse)

	// If our n.relativeBin is not divisible by the number of tbins in a global
	// bucket.
	if s.queryOne != "" {
		response, err = s.node.query(s.queryOne)
		if err != nil {
			return nil, err
		}

		responseOne := response.(*NanocubeResponse)

		if len(responseOne.Root.Children) > 0 {
			// And now merge the responses: add everything in the queryOne
			// bucket to the first bucket of queryTwo.
			for _, child := range responseTwo.Root.Children {
				// Shift over to make space for the "0" bucket.
				child.Path[0] += 1
			}

			// Add the 0 bucket.
			responseTwo.Root.Children = append(
				responseTwo.Root.Children,
				Child{
					Path: []uint{0},
					Val:  responseOne.Root.Children[0].Val,
				},
			)
		}
	}

	// Adjust for bucket offset.
	for _, child := range responseTwo.Root.Children {
		child.Path[0] += uint(s.bucketOffset)
	}

	return responseTwo, nil
}

// query queries the NanocubeNode by sending an HTTP GET request to the url
// endpoint. Examples are: "/count", "/schema".
func (n *NanocubeNode) query(url string) (JSONResponse, error) {
	// Note what kind of request this is.
	schemaRequest := strings.HasPrefix(url, "/schema")
	var response JSONResponse
	if schemaRequest {
		response = new(SchemaResponse)
	} else {
		response = new(NanocubeResponse)
	}

	rawResponse, err := http.Get(fmt.Sprintf("%v%v", n.addr, url))
	if err != nil {
		return nil, err
	}

	defer rawResponse.Body.Close()
	content, _ := ioutil.ReadAll(rawResponse.Body)

	err = response.Unmarshal(content)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// absToRelTimeQuery converts a query from a distribution-agnostic client to
// a query that takes into account a NanocubeNode's relativeBin.
// The return values are:
//
// 	string: the relative query
//	bool: whether the query lies outside the NanocubeNode's range
//
// The special time query is for queries where n.relativeBin falls into a
// bucket. It will have queryOne set to what the query should be to query the
// NanocubeNode until the start of the next bucket and then the query from the
// start of that bucket until the end global bucket.
//
// With this information, the caller can reconstruct a global answer with
// arbitrary relative offsets.
func (n *NanocubeNode) mustAbsToRelTimeQuery(query string) (
	string,
	bool,
	*SpecialTimeQuery,
) {
	timeQueries := []*regexp.Regexp{
		regexp.MustCompile("mt_interval_sequence[(][0-9]*,[0-9]*,[0-9]*[)]"),
		regexp.MustCompile("interval[(][0-9]*,[0-9]*[)]"),
	}

	queryOutsideRange := true

	relative := query
	for i, timeQuery := range timeQueries {
		for _, substr := range timeQuery.FindAllString(query, -1) {
			switch i {
			case 0:
				// mt_interval_sequence case.
				abstbins := mustSplitAndGetInts(substr, "mt_interval_sequence")
				reltbins := make([]int, len(abstbins), len(abstbins))

				reltbins[1] = abstbins[1]

				// Semantics are: first value is start time bin, second value
				// is how many time bins are in a bucket, third value is how
				// many buckets to get.
				startOffset := abstbins[0] - n.relativeBin
				endOffset := startOffset + (abstbins[1] * abstbins[2])

				if startOffset >= 0 {
					reltbins[0] = startOffset
				} else if endOffset > 0 {
					// Create a SpecialTimeQuery. First query goes to
					spTimeQuery := new(SpecialTimeQuery)

					numTBinsToStart := abstbins[1] -
						((n.relativeBin - abstbins[0]) % abstbins[1])

					if numTBinsToStart == abstbins[1] {
						// Case where n.relativBin is divisible by abstbins[1].
						numTBinsToStart = 0
					}

					// How many buckets are we going to miss?
					bucketOffset := (n.relativeBin - abstbins[0]) / abstbins[1]

					spTimeQuery.queryTwo = strings.Replace(
						relative,
						substr,
						fmt.Sprintf(
							"mt_interval_sequence(%v,%v,%v)",
							// Start from the start of the next bucket.
							strconv.Itoa(0+numTBinsToStart),
							// And use the same bucket resolution.
							strconv.Itoa(abstbins[1]),
							// And subtract however many buckets we can't
							// answer plus the one that will be answered by
							// our first partial query.
							strconv.Itoa(abstbins[2]-(bucketOffset+1)),
						),
						-1,
					)

					if numTBinsToStart > 0 {
						spTimeQuery.queryOne = strings.Replace(
							relative,
							substr,
							fmt.Sprintf(
								"mt_interval_sequence(%v,%v,%v)",
								// Start from the start of our NanocubeNode.
								strconv.Itoa(0),
								strconv.Itoa(numTBinsToStart),
								// Get one multiple of what we have above.
								strconv.Itoa(1),
							),
							-1,
						)
					}

					spTimeQuery.bucketOffset = bucketOffset
					spTimeQuery.node = n

					return "", false, spTimeQuery
				}

				reltbins[2] = int((endOffset - reltbins[0]) / reltbins[1])

				queryOutsideRange = endOffset < 0

				relative = strings.Replace(
					relative,
					substr,
					fmt.Sprintf(
						"mt_interval_sequence(%v,%v,%v)",
						strconv.Itoa(reltbins[0]),
						strconv.Itoa(reltbins[1]),
						strconv.Itoa(reltbins[2]),
					),
					-1,
				)
			case 1:
				// interval case.
				abstbins := mustSplitAndGetInts(substr, "interval")
				reltbins := make([]int, len(abstbins), len(abstbins))

				// Semantics are: first value is start time bin, second value
				// is end time bin.
				startOffset := abstbins[0] - n.relativeBin
				endOffset := abstbins[1] - n.relativeBin

				if startOffset > 0 {
					reltbins[0] = startOffset
				} else {
					reltbins[0] = 0
				}

				reltbins[1] = endOffset
				queryOutsideRange = endOffset < 0

				relative = strings.Replace(
					relative,
					substr,
					fmt.Sprintf(
						"interval(%v,%v)",
						strconv.Itoa(reltbins[0]),
						strconv.Itoa(reltbins[1]),
					),
					-1,
				)
			}
		}
	}

	return relative, queryOutsideRange, nil
}

func mustSliceAtoi(slice []string) []int {
	result := make([]int, len(slice), len(slice))
	for j, a := range slice {
		if i, err := strconv.Atoi(a); err != nil {
			panic(err)
		} else {
			result[j] = i
		}
	}

	return result
}

// mustSplitAndGetInts is a utility function to be used solely for strings of
// the form sep(0,...,9) to get a list of integers separated by commas.
func mustSplitAndGetInts(str, sep string) []int {
	split := strings.Split(str, sep)

	// Trim the parentheses then split by comma to get the three
	// values.
	return mustSliceAtoi(
		strings.Split(strings.Trim(split[1], "()"), ","),
	)
}
