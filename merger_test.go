package distnano

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"
)

// ByKey implements sort.Interface for []Child based on the Child asKey()
// method.
type ByKey []Child

var mergeNormal = []struct {
	dest, src, expected []byte
}{
	{
		[]byte(`{ "layers":[  ], "root":{ "val":23588 } }`),
		[]byte(`{ "layers":[  ], "root":{ "val":26412 } }`),
		[]byte(`{"layers":[],"root":{"val":50000}}`),
	},
}

var mergeWithPath = []struct {
	dest, src, expected []byte
}{
	{
		// First triplet tests when src is bigger than dest.
		[]byte(`{"layers":["multi-target:time"],"root":{"children":[{"path":[0],"val":293}]}}`),
		[]byte(`{"layers":["multi-target:time"],"root":{"children":[{"path":[0],"val":664},{"path":[1],"val":595},{"path":[2],"val":602},{"path":[3],"val":582},{"path":[4],"val":583},{"path":[5],"val":593},{"path":[6],"val":55}]}}`),
		[]byte(`{"layers":["multi-target:time"],"root":{"children":[{"path":[0],"val":957},{"path":[1],"val":595},{"path":[2],"val":602},{"path":[3],"val":582},{"path":[4],"val":583},{"path":[5],"val":593},{"path":[6],"val":55}]}}`),
	},
}

var mergeWithXY = []struct {
	dest, src, expected []byte
}{
	{
		[]byte(`{"layers":["anchor:location"],"root":{"children":[{"x":6,"y":131,"val":7280},{"x":7,"y":130,"val":33},{"x":6,"y":130,"val":2606}]}}`),
		[]byte(`{"layers":["anchor:location"],"root":{"children":[{"x":6,"y":131,"val":7321},{"x":7,"y":130,"val":57},{"x":6,"y":130,"val":2581},{"x":5,"y":131,"val":122}]}}`),
		[]byte(`{"layers":["anchor:location"],"root":{"children":[{"x":5,"y":131,"val":122},{"x":6,"y":130,"val":5187},{"x":6,"y":131,"val":14601},{"x":7,"y":130,"val":90}]}}`),
	},
	{
		[]byte(`{"layers":["anchor:location"],"root":{"children":[{"x":5,"y":131,"val":217},{"x":6,"y":131,"val":29436},{"x":7,"y":130,"val":166},{"x":6,"y":130,"val":10182}]}}`),
		[]byte(`{"layers":["anchor:location"],"root":{"children":[{"x":6,"y":131,"val":7384},{"x":7,"y":130,"val":58},{"x":6,"y":130,"val":2502},{"x":5,"y":131,"val":55}]}}`),
		[]byte(`{"layers":["anchor:location"],"root":{"children":[{"x":5,"y":131,"val":272},{"x":6,"y":130,"val":12684},{"x":6,"y":131,"val":36820},{"x":7,"y":130,"val":224}]}}`),
	},
}

var mergeRecursiveXY = []struct {
	dest, src, expected []byte
}{
	{
		[]byte(`{"layers":["anchor:location","multi-target:time"],"root":{"children":[{"x":1,"y":0,"children":[{"path":[0],"val":5},{"path":[2],"val":4},{"path":[3],"val":7},{"path":[4],"val":10}]},{"x":0,"y":0,"children":[{"path":[0],"val":1},{"path":[2],"val":4},{"path":[3],"val":3},{"path":[4],"val":1}]},{"x":1,"y":1,"children":[{"path":[1],"val":1},{"path":[3],"val":2},{"path":[4],"val":1}]},{"x":0,"y":1,"children":[{"path":[1],"val":3},{"path":[2],"val":4},{"path":[3],"val":2},{"path":[4],"val":2}]},{"x":1,"y":1,"children":[{"path":[0],"val":2},{"path":[2],"val":1},{"path":[3],"val":5},{"path":[4],"val":1}]},{"x":0,"y":1,"children":[{"path":[0],"val":1},{"path":[2],"val":1},{"path":[3],"val":1},{"path":[4],"val":4}]},{"x":1,"y":0,"children":[{"path":[0],"val":1},{"path":[1],"val":1},{"path":[2],"val":1},{"path":[3],"val":4},{"path":[4],"val":12}]},{"x":0,"y":0,"children":[{"path":[1],"val":1},{"path":[3],"val":1}]}]}}`),
		[]byte(`{"layers":["anchor:location","multi-target:time"],"root":{"children":[{"x":1,"y":1,"children":[{"path":[0],"val":5},{"path":[1],"val":2},{"path":[2],"val":1},{"path":[4],"val":1}]},{"x":0,"y":1,"children":[{"path":[0],"val":1},{"path":[1],"val":3},{"path":[2],"val":1},{"path":[3],"val":3},{"path":[4],"val":1}]},{"x":1,"y":0,"children":[{"path":[0],"val":1},{"path":[1],"val":3},{"path":[2],"val":6},{"path":[3],"val":4},{"path":[4],"val":11}]},{"x":0,"y":0,"children":[{"path":[0],"val":3},{"path":[1],"val":1},{"path":[3],"val":3},{"path":[4],"val":2}]}]}}`),
		[]byte(`{"layers":["anchor:location","multi-target:time"],"root":{"children":[{"x":1,"y":1,"children":[{"path":[0],"val":5},{"path":[1],"val":2},{"path":[2],"val":1},{"path":[4],"val":1}]},{"x":1,"y":0,"children":[{"path":[0],"val":1},{"path":[1],"val":3},{"path":[2],"val":6},{"path":[3],"val":4},{"path":[4],"val":11}]},{"x":0,"y":1,"children":[{"path":[0],"val":1},{"path":[1],"val":3},{"path":[2],"val":1},{"path":[3],"val":3},{"path":[4],"val":1}]},{"x":0,"y":0,"children":[{"path":[0],"val":3},{"path":[1],"val":1},{"path":[3],"val":3},{"path":[4],"val":2}]}]}}`),
	},
}

func mustMarshalNanocubeResponse(t *testing.T, ncr *NanocubeResponse) []byte {
	b, err := json.Marshal(ncr)
	if err != nil {
		t.Fatalf("Unexpected error: %v\n", err)
	}
	return b
}

func mustUnmarshalSlice(t *testing.T, b []byte) *NanocubeResponse {
	ncr := new(NanocubeResponse)
	err := json.Unmarshal(b, ncr)
	if err != nil {
		t.Fatalf("Unexpected error: %v\n", err)
	}
	return ncr
}

func (s ByKey) Len() int           { return len(s) }
func (s ByKey) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByKey) Less(i, j int) bool { return s[i].asKey() < s[j].asKey() }

// ncrPrepare sorts all of dest's []Child slices so that reflect.DeepEqual can
// be used.
func ncrPrepare(dest *NanocubeResponse) {
	if dest.Root.Children != nil {
		childPrepare(dest.Root.Children)
	}
}

// childPrepare sorts a []Child slice recursively traversin any addition
// []Child slices.
func childPrepare(children []Child) {
	for _, child := range children {
		if child.Children != nil {
			childPrepare(child.Children)
		}
	}

	sort.Sort(ByKey(children))
}

func TestMergeNormal(t *testing.T) {
	for _, e := range mergeNormal {
		dest := mustUnmarshalSlice(t, e.dest)
		merge(dest, mustUnmarshalSlice(t, e.src))
		ans := mustMarshalNanocubeResponse(t, dest)
		if !reflect.DeepEqual(ans, e.expected) {
			t.Fatalf("%v was not equal to %v", string(ans), string(e.expected))
		}
	}
}

func TestMergeWithPath(t *testing.T) {
	for _, e := range mergeWithPath {
		dest := mustUnmarshalSlice(t, e.dest)
		merge(dest, mustUnmarshalSlice(t, e.src))

		ncrPrepare(dest)
		ans := mustMarshalNanocubeResponse(t, dest)
		if !reflect.DeepEqual(ans, e.expected) {
			t.Fatalf("%v was not equal to %v", string(ans), string(e.expected))
		}
	}
}

func TestMergeWithXY(t *testing.T) {
	for _, e := range mergeWithXY {
		dest := mustUnmarshalSlice(t, e.dest)
		merge(dest, mustUnmarshalSlice(t, e.src))

		ncrPrepare(dest)
		ans := mustMarshalNanocubeResponse(t, dest)
		if !reflect.DeepEqual(ans, e.expected) {
			t.Fatalf("%v was not equal to %v", string(ans), string(e.expected))
		}
	}
}
