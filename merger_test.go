package distnano

import (
	"reflect"
	"sort"
	"testing"
)

// ByKey implements sort.Interface for []Child based on the Child asKey()
// method.
type ByKey []Child

// TODO(asubiotto): Specify these as actual response objects, it seems that
// JSON is an unnecessary layer in this test suite.
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

var mergeSchema = []struct {
	dest, src, expected []byte
}{
	{
		[]byte(`{ "fields":[ { "name":"location", "type":"nc_dim_quadtree_25", "valnames":{  } }, { "name":"crime", "type":"nc_dim_cat_1", "valnames":{ "CRIM_SEXUAL_ASSAULT":6, "OFFENSE_INVOLVING_CHILDREN":16, "CRIMINAL_TRESPASS":5, "SEX_OFFENSE":21, "INTIMIDATION":11, "KIDNAPPING":12, "PROSTITUTION":18, "PUBLIC_PEACE_VIOLATION":19, "WEAPONS_VIOLATION":24, "ASSAULT":1, "MOTOR_VEHICLE_THEFT":14, "GAMBLING":8, "OTHER_OFFENSE":17, "DECEPTIVE_PRACTICE":7, "NARCOTICS":15, "CRIMINAL_DAMAGE":4, "LIQUOR_LAW_VIOLATION":13, "BATTERY":2, "THEFT":23, "ARSON":0, "STALKING":22 } }, { "name":"time", "type":"nc_dim_time_2", "valnames":{  } }, { "name":"count", "type":"nc_var_uint_4", "valnames":{  } } ], "metadata":[ { "key":"tbin", "value":"2014-01-01_00:00:00_3600s" }, { "key":"location__origin", "value":"degrees_mercator_quadtree25" }, { "key":"name", "value":"xac" } ] }`),
		[]byte(`{ "fields":[ { "name":"location", "type":"nc_dim_quadtree_25", "valnames":{  } }, { "name":"crime", "type":"nc_dim_cat_1", "valnames":{ "CRIM_SEXUAL_ASSAULT":6, "OFFENSE_INVOLVING_CHILDREN":16, "CRIMINAL_TRESPASS":5, "INTIMIDATION":10, "KIDNAPPING":12, "ROBBERY":20, "HOMICIDE":9, "BURGLARY":3, "WEAPONS_VIOLATION":24, "ASSAULT":1, "MOTOR_VEHICLE_THEFT":14, "ARSON":1, "STALKING":0, "INTERFERENCE_WITH_PUBLIC_OFFICER":10 } }, { "name":"time", "type":"nc_dim_time_2", "valnames":{  } }, { "name":"count", "type":"nc_var_uint_4", "valnames":{  } } ], "metadata":[ { "key":"tbin", "value":"2014-01-01_00:00:00_3600s" }, { "key":"location__origin", "value":"degrees_mercator_quadtree25" }, { "key":"name", "value":"xad" } ] }`),
		[]byte(`{ "fields":[ { "name":"location", "type":"nc_dim_quadtree_25", "valnames":{  } }, { "name":"crime", "type":"nc_dim_cat_1", "valnames":{ "CRIMINAL_TRESPASS":5, "KIDNAPPING":12, "ASSAULT":1, "LIQUOR_LAW_VIOLATION":13, "BURGLARY":3, "SEX_OFFENSE":21, "PUBLIC_PEACE_VIOLATION":19, "ARSON":0, "HOMICIDE":9, "CRIM_SEXUAL_ASSAULT":6, "OFFENSE_INVOLVING_CHILDREN":16, "INTIMIDATION":11, "PROSTITUTION":18, "WEAPONS_VIOLATION":24, "NARCOTICS":15, "ROBBERY":20, "STALKING":22, "MOTOR_VEHICLE_THEFT":14, "GAMBLING":8, "OTHER_OFFENSE":17, "DECEPTIVE_PRACTICE":7, "CRIMINAL_DAMAGE":4, "BATTERY":2, "THEFT":23, "INTERFERENCE_WITH_PUBLIC_OFFICER":10} }, { "name":"time", "type":"nc_dim_time_2", "valnames":{  } }, { "name":"count", "type":"nc_var_uint_4", "valnames":{  } } ], "metadata":[ { "key":"tbin", "value":"2014-01-01_00:00:00_3600s" }, { "key":"location__origin", "value":"degrees_mercator_quadtree25" }, { "key":"name", "value":"xac" } ] }`),
	},
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
		dest, src := new(NanocubeResponse), new(NanocubeResponse)

		dest.Unmarshal(e.dest)
		src.Unmarshal(e.src)
		dest.Merge(src)

		ncrPrepare(dest)
		ans, _ := dest.Marshal()
		if !reflect.DeepEqual(ans, e.expected) {
			t.Fatalf("%v was not equal to %v", string(ans), string(e.expected))
		}
	}
}

func TestMergeWithPath(t *testing.T) {
	for _, e := range mergeWithPath {
		dest, src := new(NanocubeResponse), new(NanocubeResponse)

		dest.Unmarshal(e.dest)
		src.Unmarshal(e.src)
		dest.Merge(src)

		ncrPrepare(dest)
		ans, _ := dest.Marshal()
		if !reflect.DeepEqual(ans, e.expected) {
			t.Fatalf("%v was not equal to %v", string(ans), string(e.expected))
		}
	}
}

func TestMergeWithXY(t *testing.T) {
	for _, e := range mergeWithXY {
		dest, src := new(NanocubeResponse), new(NanocubeResponse)

		dest.Unmarshal(e.dest)
		src.Unmarshal(e.src)
		dest.Merge(src)

		ncrPrepare(dest)
		ans, _ := dest.Marshal()
		if !reflect.DeepEqual(ans, e.expected) {
			t.Fatalf("%v was not equal to %v", string(ans), string(e.expected))
		}
	}
}

func TestMergeSchema(t *testing.T) {
	for _, e := range mergeSchema {
		dest, src := new(SchemaResponse), new(SchemaResponse)
		expected := new(SchemaResponse)

		dest.Unmarshal(e.dest)
		src.Unmarshal(e.src)

		// We unmarshal expected so that map equality is not order dependent.
		expected.Unmarshal(e.expected)

		dest.Merge(src)
		if !reflect.DeepEqual(*dest, *expected) {
			t.Fatalf("%v was not equal to %v", *dest, *expected)
		}
	}
}
