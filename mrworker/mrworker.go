package main

import (
	"log"
	"plugin"

	"github.com/shortykevich/map-reduce/mr"
	"github.com/shortykevich/map-reduce/mrapps"
)

func main() {
	// if len(os.Args) != 2 {
	// 	fmt.Fprintf(os.Stderr, "Usage: mrworker xxx.so\n")
	// 	os.Exit(1)
	// }

	// mapf, reducef := loadPlugin(os.Args[1])

	mapf, reducef := mrapps.Map, mrapps.Reduce
	mr.Worker(mapf, reducef)
}

// load the application Map and Reduce functions
// from a plugin file, e.g. ../mrapps/wc.so
func loadPlugin(filename string) (func(string, string) []mr.KeyValue, func(string, []string) string) {
	p, err := plugin.Open(filename)
	if err != nil {
		log.Fatalf("cannot load plugin %v", filename)
	}
	xmapf, err := p.Lookup("Map")
	if err != nil {
		log.Fatalf("cannot find Map in %v", filename)
	}
	mapf := xmapf.(func(string, string) []mr.KeyValue)
	xreducef, err := p.Lookup("Reduce")
	if err != nil {
		log.Fatalf("cannot find Reduce in %v", filename)
	}
	reducef := xreducef.(func(string, []string) string)

	return mapf, reducef
}
