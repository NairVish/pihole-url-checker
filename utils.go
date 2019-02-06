package main

import (
	"log"
	"regexp"
)

var defaultPiholeListRoot = "/etc/pihole"                           // the default Pi-hole root folder
var ipAddrRegex = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`) // regex corresponding to an IP address

// logFatalIfError is a helper function that takes in an error and executes log.Fatal if the error is not nil.
func logFatalIfError(err error) {
	if err != nil {
		// debug.PrintStack()
		log.Fatal(err)
	}
}
