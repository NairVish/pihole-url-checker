package main

import (
	"fmt"
	"os"
	"regexp"
)

var piholeListRoot = "/etc/pihole"
var ipAddrRegex = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)

// TODO: Whitelists (whitelist.txt) and wildcard/regex blocking (regex.list).

func main() {
	err := os.Chdir(piholeListRoot)
	logFatalIfError(err)

	if len(os.Args) != 2 {
		logFatalIfError(fmt.Errorf("USAGE: %s <url_to_check>", os.Args[0]))
	}

	fmt.Printf("QUERY: %s\nSearching. This may take a while depending on the number and size of your blocklists...\n", os.Args[1])
	search_results := searchForURLInAllLists(os.Args[1])
	fmt.Println(stringifyResults(search_results))
}
