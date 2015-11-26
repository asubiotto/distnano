package distnano

import (
	"encoding/json"
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
	result := make([]Child, 0, len(destMap))

	for _, v := range destMap {
		// Note that result is not reallocated on each iteration because its
		// capacity was specified.
		result = append(result, *v)
	}

	return result
}

// merge merges src into dest according to the API.md document found in
// github.com/laurolins/nanocube.
func merge(dest, src *NanocubeResponse) {
	if dest.Layers == nil {
		dest.Layers = []string{}
	}

	if dest.Root.Val != nil {
		*(dest.Root.Val) += *(src.Root.Val)
		return
	}

	dest.Root.Children = mergeChildren(dest.Root.Children, src.Root.Children)
}
