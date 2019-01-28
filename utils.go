package main

import (
	"log"
)

func logFatalIfError(err error) {
	if err != nil {
		// debug.PrintStack()
		log.Fatal(err)
	}
}
