package distnano

import (
	"encoding/json"
	"log"
	"sort"
)

// asKey is a function that encodes c's path or x and y values as a string to
// be used as a key in a map. This is necessary to produce quick aggregations
// on keys.
func (c Child) asKey() string {
	// Want to ignore values that are being aggregated over.
	c.Val = nil
	c.Children = nil

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
func mergeNanocubeResponse(dest, src *NanocubeResponse) {
	if dest.Layers == nil {
		dest.Layers = []string{}
	}

	if dest.Root.Val != nil {
		*(dest.Root.Val) += *(src.Root.Val)
		return
	}

	dest.Root.Children = mergeChildren(dest.Root.Children, src.Root.Children)
}

func mergeFields(dest, src []Field) []Field {
	if len(dest) != len(src) {
		log.Fatal("Schema fields not the same length")
	}

	for i := 0; i < len(dest); i++ {
		// Merge the valname maps together.
		for k, v := range src[i].Valnames {
			dest[i].Valnames[k] += v
		}

		keys := make(sort.StringSlice, 0, len(dest[i].Valnames))
		for k, _ := range dest[i].Valnames {
			keys = append(keys, k)
		}

		// The semantics of the schema call is to respond with the index of the
		// key in sorted order as the value.
		keys.Sort()
		for j := 0; j < len(keys); j++ {
			dest[i].Valnames[keys[j]] = j
		}
	}

	return dest
}

func mergeSchemaResponse(dest, src *SchemaResponse) {
	for _, md := range dest.Metadata {
		// Every node will have a different filename, so this key value pair is
		// not necessary.
		if md.Key == "name" {
			md.Value = ""
		}
	}

	dest.Fields = mergeFields(dest.Fields, src.Fields)
}
