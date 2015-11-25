package distnano

import (
	"encoding/json"
	"fmt"
	"log"
)

type Path []uint

type Child struct {
	Path     Path    `json:"path,omitempty"`
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

func (p *Path) asKey() string {
	b, err := json.Marshal(&p)
	if err != nil {
		panic(fmt.Sprintf("Got %v encoding path %v", err, p))
	}
	return string(b)
}

func stringAsPath(s string) Path {
	p := Path{}
	b := []byte(s)
	if err := json.Unmarshal(b, &p); err != nil {
		panic(fmt.Sprintf("Got %v decoding %v", err, b))
	}
	return p
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

	destMap := make(map[string]int)
	for _, child := range dest {
		// If there's a path there's a value.
		if child.Val == nil {
			log.Fatal("Child value was nil in dest")
		}
		destMap[child.Path.asKey()] = *(child.Val)
	}

	for _, child := range src {
		if child.Val == nil {
			log.Fatal("Child value was nil in src")
		}

		key := child.Path.asKey()
		if _, e := destMap[key]; e {
			// If it exists, we add the value that src has for this key.
			destMap[key] += *(child.Val)
		} else {
			// Otherwise we add it.
			destMap[key] = *(child.Val)
		}
	}

	// Now that we have the results in destMap, we are going to update dest's
	// children slice.
	for _, child := range dest {
		key := child.Path.asKey()
		*(child.Val) = destMap[key]
		delete(destMap, key)
	}

	if len(destMap) == 0 {
		return dest
	}

	// The keys that are left after this are paths that were not in dest to
	// begin with.
	newDest := make([]Child, len(dest)+len(destMap))
	copy(newDest, dest)
	i := len(dest)
	for k, v := range destMap {
		val := new(int)
		*val = v
		newDest[i] = Child{Path: stringAsPath(k), Val: val}
		i++
	}
	return newDest
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
