package main

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var piholeListRoot = "/etc/pihole"
var ipAddrRegex = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)

type ListResult struct {
	LineNumber   int
	ListFileName string
	ListURL      string
}

type TotalResult struct {
	QueryURL     string
	AllBLMatches []ListResult
}

func getAllListURLs() []string {
	list_urls_filename := filepath.Join(piholeListRoot, "adlists.list")
	full_url_list, err := os.Open(list_urls_filename)
	if err != nil {
		log.Fatal(err)
	}
	defer full_url_list.Close()

	url_list := make([]string, 25)
	file_scanner := bufio.NewScanner(full_url_list)
	for file_scanner.Scan() {
		this_url := file_scanner.Text()
		if !strings.HasPrefix(this_url, "#") || len(this_url) == 0 { // if the URL isn't commented out or an empty line...
			url_list = append(url_list, file_scanner.Text())
		}
	}

	if err := file_scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return url_list
}

func main() {

}
