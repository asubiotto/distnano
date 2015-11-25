package distnano

import (
	"encoding/json"
	"fmt"
	"log"
)

type Child struct {
	Path     []uint  `json:"path,omitempty"`
	X        *int    `json:"x,omitempty"`
	Y        *int    `json:"y,omitempty"`
	Val      *int    `json:"val,omitempty"`
	Children []Child `json:"children,omitempty"`
}

type Root struct {
	Val      *int    `json:"val,omitempty"`
	Children []Child `json:"children,omitempty"`
}

type NanocubeResponse struct {
	Layers []string `json:"layers"`
	Root   Root     `json:"root"`
}

// asKey is a function that encodes c's path or x and y values as a string to
// be used as a key in a map. This is necessary to produce quick aggregations
// on keys.
func (c Child) asKey() string {
	// Want to ignore val.
	c.Val = nil
	b, err := json.Marshal(c)
	if err != nil {
		log.Fatalf("Got %v encoding child %v", err, c)
	}
	return string(b)
}

// stringAsChild decodes the string into a Child by unmarshaling the string and
// returning as a result either a Child with its Path set or its X and Y set.
// This is useful to us once we have finished aggregating using the map and
// want to convert the map back to a []Child.
func stringAsChild(s string) Child {
	c := Child{}
	b := []byte(s)
	if err := json.Unmarshal(b, &c); err != nil {
		log.Fatalf("Got %v decoding %v", err, b)
	}
	return c
}

// TODO(asubiotto): The whole error handling has to be looked at more closely.

// mergeChildren merges the children array by aggregating "val"s at the base
// level and using either path as key or x and y as keys.
func mergeChildren(dest, src []Child) []Child {
	if dest == nil || src == nil {
		if dest == nil {
			return src
		}
		return dest
	}

	// Store pointers to children accesible by key.
	destMap := make(map[string]*Child)
	for i, child := range dest {
		destMap[child.asKey()] = &dest[i]
	}

	// For each child in src, merge with the destMap.
	for i, child := range src {
		key := child.asKey()

		// If the key already exists in the map, we want to merge with that
		// key. Otherwise, we simply add it.
		if _, e := destMap[key]; e {
			// If the child doesn't have a Val, it necessarily has something in
			// its []Child field. We therefore initiate a recursive merge.
			// Otherwise, the semantics of merging is to add to the value.
			if child.Val == nil {
				destMap[key].Children =
					mergeChildren(destMap[key].Children, child.Children)
			} else {
				*(destMap[key].Val) += *(child.Val)
			}
		} else {
			destMap[key] = &src[i]
		}
	}

	// We now want to create a new []Child slice from our [string]*Child map.
	result := make([]Child, len(destMap))
	i := 0

	for _, v := range destMap {
		result[i] = *v
		i++
	}

	return result
}

// merge merges src into dest according to the API.md document found in
// github.com/laurolins/nanocube.
func merge(dest, src *NanocubeResponse) {
	// Little check for layers. TODO(asubiotto): Think about this case a bit
	// more when we have empty responses or we want to return an error.
	if dest.Layers == nil {
		dest.Layers = []string{}
	}

	if dest.Root.Val != nil {
		*(dest.Root.Val) += *(src.Root.Val)
		return
	}

	b, _ := json.Marshal(dest)
	b1, _ := json.Marshal(src)
	fmt.Printf("Merging children %v and %v\n\n", string(b), string(b1))
	dest.Root.Children = mergeChildren(dest.Root.Children, src.Root.Children)
	b, _ = json.Marshal(dest)
	fmt.Printf("Got: %v\n\n\n\n", string(b))
}
