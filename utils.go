package main

import (
	"log"
	"regexp"
)

var defaultPiholeListRoot = "/etc/pihole"
var ipAddrRegex = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)

func logFatalIfError(err error) {
	if err != nil {
		// debug.PrintStack()
		log.Fatal(err)
	}
}
