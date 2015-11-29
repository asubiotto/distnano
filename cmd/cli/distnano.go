package main

import (
	"flag"
	"fmt"

	"github.com/asubiotto/distnano"
)

// stringslice implements the flag.Value interface so that we can pass a list
// of addresses to the program.
type stringslice []string

var addrs stringslice
var port int

func (s *stringslice) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *stringslice) Set(v string) error {
	*s = append(*s, v)
	return nil
}

// Example run:
// 	go run cmd/cli/distnano.go -a http://localhost:9001 \
// 		-a http://localhost:9002 -a http://localhost:9003 \
//		-a http://localhost:9004 -a http://localhost:9005
// Where all the addresses specified are running nanocube partitions and
// hosting queries.
func main() {
	usage := "Port that we should listen on"
	flag.IntVar(&port, "port", 29512, usage)
	flag.IntVar(&port, "p", 29512, usage+" (shorthand)")

	usage = "List of addresses that our server should interact with. Note " +
		"that the scheme and host must be specified (e.g. " +
		"http://localhost:9000 not just localhost:9000 or 9000)"
	flag.Var(&addrs, "addrs", usage)
	flag.Var(&addrs, "a", usage+" (shorthand)")

	flag.Parse()

	distnano.Run(port, addrs)
}
