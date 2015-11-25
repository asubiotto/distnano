package distnano

import (
	"encoding/json"
	"reflect"
	"testing"
)

var countNormal = []struct {
	dest, src, expected []byte
}{
	{
		[]byte(`{ "layers":[  ], "root":{ "val":23588 } }`),
		[]byte(`{ "layers":[  ], "root":{ "val":26412 } }`),
		[]byte(`{"layers":[],"root":{"val":50000}}`),
	},
}

var countWithPath = []struct {
	dest, src, expected []byte
}{
	{
		// First triplet tests when src is bigger than dest.
		[]byte(`{"layers":["multi-target:time"],"root":{"children":[{"path":[0],"val":293}]}}`),
		[]byte(`{"layers":["multi-target:time"],"root":{"children":[{"path":[0],"val":664},{"path":[1],"val":595},{"path":[2],"val":602},{"path":[3],"val":582},{"path":[4],"val":583},{"path":[5],"val":593},{"path":[6],"val":55}]}}`),
		[]byte(`{"layers":["multi-target:time"],"root":{"children":[{"path":[0],"val":957},{"path":[1],"val":595},{"path":[2],"val":602},{"path":[3],"val":582},{"path":[4],"val":583},{"path":[5],"val":593},{"path":[6],"val":55}]}}`),
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

func TestCountNormal(t *testing.T) {
	for _, e := range countNormal {
		dest := mustUnmarshalSlice(t, e.dest)
		merge(dest, mustUnmarshalSlice(t, e.src))
		ans := mustMarshalNanocubeResponse(t, dest)
		if !reflect.DeepEqual(ans, e.expected) {
			t.Fatalf("%v was not equal to %v", string(ans), string(e.expected))
		}
	}
}

func TestCountWithPath(t *testing.T) {
	for _, e := range countWithPath {
		dest := mustUnmarshalSlice(t, e.dest)
		merge(dest, mustUnmarshalSlice(t, e.src))
		ans := mustMarshalNanocubeResponse(t, dest)
		if !reflect.DeepEqual(ans, e.expected) {
			t.Fatalf("%v was not equal to %v", string(ans), string(e.expected))
		}
	}
}
