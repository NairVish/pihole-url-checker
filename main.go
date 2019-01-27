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
	LineText     string
}

type TotalResult struct {
	QueryURL       string
	ExactBLMatches []ListResult
	ApprxBLMatches []ListResult
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

	exact_matches := make([]ListResult, 2)
	approx_matches := make([]ListResult, 2)
	for i := range all_list_urls {
		this_list_filename := all_list_filenames[i]
		this_list_file, err := os.Open(this_list_filename)
		logFatalIfError(err)
		file_scanner := bufio.NewScanner(this_list_file)

		// scan file
		for file_scanner.Scan() {
			this_entry := file_scanner.Text()
			orig_entry := file_scanner.Text()
			if strings.HasPrefix(this_entry, "#") || len(this_entry) == 0 {
				continue
			}

			if ipAddrRegex.MatchString(this_entry) {
				this_entry = ipAddrRegex.ReplaceAllLiteralString(this_entry, "")
				this_entry = strings.TrimSpace(this_entry)
			}

			if query == this_entry {
				exact_matches = append(exact_matches, ListResult{LineNumber: i, ListFileName: this_list_filename, ListURL: all_list_urls[i], LineText: orig_entry})
			} else if strings.Contains(this_entry, query) {
				approx_matches = append(approx_matches, ListResult{LineNumber: i, ListFileName: this_list_filename, ListURL: all_list_urls[i], LineText: orig_entry})
			}
		}

		err = file_scanner.Err()
		logFatalIfError(err)
		this_list_file.Close()
	}

	return &TotalResult{QueryURL: query, ExactBLMatches: exact_matches, ApprxBLMatches: approx_matches}
}

func stringifyResults(result *TotalResult) string {
	if len(result.ExactBLMatches) == 0 && len(result.ApprxBLMatches) == 0 {
		return fmt.Sprintf("No results found in existing, active blocklists for %s.", result.QueryURL)
	}

	str := fmt.Sprintf("RESULTS:\nQuery: %s\n", result.QueryURL)
	if len(result.ExactBLMatches) > 0 {
		str += "\nEXACT MATCHES:\n"
		for _, m := range result.ExactBLMatches {
			str += fmt.Sprintf("%s (%s)\n\tLine %d\n\tEntry: %s\n", m.ListFileName, m.ListURL, m.LineNumber, m.LineText)
		}
	}
	if len(result.ApprxBLMatches) > 0 {
		str += "\nAPPROXIMATE MATCHES (may not result in blocks of the query):\n"
		for _, m := range result.ApprxBLMatches {
			str += fmt.Sprintf("%s (%s)\n\tLine %d\n\tEntry: %s\n", m.ListFileName, m.ListURL, m.LineNumber, m.LineText)
		}
	}

	return str
}

func main() {
	err := os.Chdir(piholeListRoot)
	logFatalIfError(err)

	if len(os.Args) != 2 {
		logFatalIfError(fmt.Errorf("USAGE: %s <url_to_check>", os.Args[0]))
	}

	search_results := searchForURLInAllLists(os.Args[1])
	fmt.Println(stringifyResults(search_results))
}
