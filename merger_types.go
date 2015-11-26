package distnano

import (
	"encoding/json"
	"log"
	"reflect"
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

type Field struct {
	Name     string         `json:"name,omitempty"`
	Type     string         `json:"type,omitempty"`
	Valnames map[string]int `json:"valnames"`
}

type Metadata struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type SchemaResponse struct {
	Fields   []Field    `json:"fields"`
	Metadata []Metadata `json:"metadata"`
}

type JSONResponse interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
	Merge(JSONResponse)
}

func (n *NanocubeResponse) Marshal() ([]byte, error) {
	return json.Marshal(n)
}

func (n *NanocubeResponse) Unmarshal(b []byte) error {
	return json.Unmarshal(b, n)
}

func (n *NanocubeResponse) Merge(j JSONResponse) {
	ncr, ok := j.(*NanocubeResponse)
	if !ok {
		log.Fatal(
			"Tried to merge NanocubeResponse with unexpected type %v\n",
			reflect.TypeOf(j),
		)
	}

	mergeNanocubeResponse(n, ncr)
}

func (s *SchemaResponse) Marshal() ([]byte, error) {
	return json.Marshal(s)
}

func (s *SchemaResponse) Unmarshal(b []byte) error {
	return json.Unmarshal(b, s)
}

func (s *SchemaResponse) Merge(j JSONResponse) {
	scr, ok := j.(*SchemaResponse)
	if !ok {
		log.Fatal(
			"Tried to merge SchemaResponse with unexpected type %v\n",
			reflect.TypeOf(j),
		)
	}

	mergeSchemaResponse(s, scr)
}
