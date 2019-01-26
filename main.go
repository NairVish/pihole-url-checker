package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var piholeListRoot = "/etc/pihole"
var ipAddrRegex = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)

// TODO: Whitelists and wildcard/regex blocking.

type ListResult struct {
	LineNumber   int
	ListFileName string
	ListURL      string
}

type TotalResult struct {
	QueryURL     string
	AllBLMatches []ListResult
}

func logFatalIfError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func getAllListURLs() []string {
	list_urls_filename := filepath.Join(piholeListRoot, "adlists.list")
	full_url_list, err := os.Open(list_urls_filename)
	logFatalIfError(err)
	defer full_url_list.Close()

	url_list := make([]string, 25)
	url_list = append(url_list, "black.list")
	file_scanner := bufio.NewScanner(full_url_list)
	for file_scanner.Scan() {
		this_url := file_scanner.Text()
		if !strings.HasPrefix(this_url, "#") || len(this_url) > 0 { // if the URL isn't commented out or an empty line...
			url_list = append(url_list, file_scanner.Text())
		}
	}

	err = file_scanner.Err()
	logFatalIfError(err)

	return url_list
}

func getAllBlocklistFileNames() []string {
	filenames, err := ioutil.ReadDir(piholeListRoot)
	logFatalIfError(err)

	filelist := make([]string, 25)
	filelist = append(filelist, "black.list")
	for _, f := range filenames {
		if f.IsDir() {
			continue
		}

		if strings.HasPrefix(f.Name(), "list.") && strings.HasSuffix(f.Name(), ".domains") {
			filelist = append(filelist, f.Name())
		}
	}
	sort.Strings(filelist)

	return filelist
}

func searchForURLInAllLists(query string) *TotalResult {
	all_list_urls := getAllListURLs()
	all_list_filenames := getAllBlocklistFileNames()
	if len(all_list_urls) != len(all_list_filenames) {
		log.Fatal(fmt.Printf("len(all_list_urls) [%d] != len(all_list_filenames) [%d]", len(all_list_urls), len(all_list_filenames)))
	}

	all_matches := make([]ListResult, 2)
	for i := range all_list_urls {
		this_list_filename := all_list_filenames[i]
		this_list_file, err := os.Open(this_list_filename)
		logFatalIfError(err)

		// TODO

		this_list_file.Close()
	}

	return &TotalResult{QueryURL: query, AllBLMatches: all_matches}
}

func main() {
	err := os.Chdir(piholeListRoot)
	logFatalIfError(err)
}
