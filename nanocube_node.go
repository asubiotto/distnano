package distnano

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"
)

type TBin struct {
	startTime time.Time
	binSize   time.Duration
}

// ByTime implements sort.Interface for []TBin based on the TBin's startTime.
type ByTime []TBin

type NanocubeNode struct {
	addr        string
	tBin        TBin
	relativeBin int
}

func (t ByTime) Len() int           { return len(t) }
func (t ByTime) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t ByTime) Less(i, j int) bool { return t[i].startTime.Before(t[j].startTime) }

// nanocubeNodesFromAddrs converts addrs to NanocubeNodes by querying the
// schema of each address and creating NanocubeNodes accordingly.
func nanocubeNodesFromAddrs(addrs []string) ([]*NanocubeNode, error) {
	// Get each schema and the tbin string to determine each node's notion of
	// time.
	tbins := make([]TBin, len(addrs), len(addrs))
	for i, addr := range addrs {
		url := fmt.Sprintf("%v%v", addr, "/schema")
		rawResponse, err := http.Get(url)
		if err != nil {
			return nil, err
		}

		defer rawResponse.Body.Close()
		content, _ := ioutil.ReadAll(rawResponse.Body)

		response := new(SchemaResponse)
		err = response.Unmarshal(content)
		if err != nil {
			return nil, err
		}

		for _, md := range response.Metadata {
			if md.Key == "tbin" {
				tbin := newTBin(md.Value)
				if tbin == nil {
					return nil, errors.New("error creating time bin")
				}
				tbins[i] = *tbin
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

func newNanocubeNode(addr string, tBin TBin) *NanocubeNode {
	nn := new(NanocubeNode)
	nn.addr = addr
	nn.tBin = tBin
	return nn
}
