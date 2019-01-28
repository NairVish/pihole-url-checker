package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"os"
)

func main() {
	// create new parser object
	parser := argparse.NewParser(os.Args[0], "Checks if the given URL/query is present in any (active) Pi-hole blocklists.")
	// define query and pi-hole root folder arguments and parse given arguments
	q := parser.String("q", "query", &argparse.Options{Required: true, Help: "Query URL to search"})
	r := parser.String("r", "root", &argparse.Options{Required: false, Help: "Pi-hole's root folder", Default: defaultPiholeListRoot})
	err := parser.Parse(os.Args)
	logFatalIfError(err)

	so := NewSearchObj(*r)
	fmt.Printf("QUERY: %s\nSearching. This may take a while depending on the number and size of your blocklists...\n", *q)
	so.SearchForURLInAllLists(*q)
	fmt.Println(so.StringifyResults())
}
