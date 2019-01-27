package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"sort"
	"strconv"
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

type BlacklistPriority struct {
	Priority int
	Filename string
}

// for sort.Sort
type BLPriorities []BlacklistPriority

func (blp BLPriorities) Len() int      { return len(blp) }
func (blp BLPriorities) Swap(i, j int) { blp[i], blp[j] = blp[j], blp[i] }
func (blp BLPriorities) Less(i, j int) bool {
	return blp[i].Priority < blp[j].Priority
}

func logFatalIfError(err error) {
	if err != nil {
		debug.PrintStack()
		log.Fatal(err)
	}
}

func getAllListURLs() []string {
	list_urls_filename := filepath.Join(piholeListRoot, "adlists.list")
	full_url_list, err := os.Open(list_urls_filename)
	logFatalIfError(err)
	defer full_url_list.Close()

	url_list := make([]string, 0)
	url_list = append(url_list, "black.list")
	file_scanner := bufio.NewScanner(full_url_list)
	for file_scanner.Scan() {
		this_url := strings.TrimSpace(file_scanner.Text())
		if !strings.HasPrefix(this_url, "#") && len(this_url) > 0 { // if the URL isn't commented out or an empty line...
			url_list = append(url_list, this_url)
		}
	}

	err = file_scanner.Err()
	logFatalIfError(err)

	return url_list
}

func getAllBlocklistFileNames() []string {
	filenames, err := ioutil.ReadDir(piholeListRoot)
	logFatalIfError(err)

	filelist := make(BLPriorities, 0)
	filelist = append(filelist, BlacklistPriority{Priority: -1, Filename: "black.list"})
	for _, f := range filenames {
		fname := strings.TrimSpace(f.Name())

		if f.IsDir() || len(fname) == 0 {
			continue
		}

		if strings.HasPrefix(fname, "list.") && strings.HasSuffix(fname, ".domains") {
			num, err := strconv.Atoi(strings.Split(fname, ".")[1])
			logFatalIfError(err)
			filelist = append(filelist, BlacklistPriority{Priority: num, Filename: fname})
		}
	}

	// using sort.Sort instead of sort.Slice b/c the default go version on Raspbian is 1.7.4 (< 1.8)
	sort.Sort(filelist)

	final_flist := make([]string, 0)
	for _, bp := range filelist {
		final_flist = append(final_flist, bp.Filename)
	}

	return final_flist
}

func searchForURLInAllLists(query string) *TotalResult {
	all_list_urls := getAllListURLs()
	all_list_filenames := getAllBlocklistFileNames()
	if len(all_list_urls) != len(all_list_filenames) {
		log.Fatal(fmt.Printf("len(all_list_urls) [%d] != len(all_list_filenames) [%d]", len(all_list_urls), len(all_list_filenames)))
	}

	exact_matches := make([]ListResult, 0)
	approx_matches := make([]ListResult, 0)
	for i := range all_list_urls {
		this_list_filename := all_list_filenames[i]
		this_list_file, err := os.Open(this_list_filename)
		logFatalIfError(err)
		file_scanner := bufio.NewScanner(this_list_file)

		// scan file
		for j := 0; file_scanner.Scan(); j += 1 {
			this_entry := strings.TrimSpace(file_scanner.Text())
			orig_entry := strings.TrimSpace(file_scanner.Text())
			if strings.HasPrefix(this_entry, "#") || len(this_entry) == 0 {
				continue
			}

			if ipAddrRegex.MatchString(this_entry) {
				this_entry = ipAddrRegex.ReplaceAllLiteralString(this_entry, "")
				this_entry = strings.TrimSpace(this_entry)
			}

			if query == this_entry {
				exact_matches = append(exact_matches, ListResult{LineNumber: j, ListFileName: this_list_filename, ListURL: all_list_urls[i], LineText: orig_entry})
			} else if strings.Contains(this_entry, query) {
				approx_matches = append(approx_matches, ListResult{LineNumber: j, ListFileName: this_list_filename, ListURL: all_list_urls[i], LineText: orig_entry})
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

	str := fmt.Sprintf("\nFound %d exact matches and %d approximate matches.\n", len(result.ExactBLMatches), len(result.ApprxBLMatches))
	if len(result.ExactBLMatches) > 0 {
		str += "\nEXACT MATCHES:\n"
		for i, m := range result.ExactBLMatches {
			str += fmt.Sprintf("%02d) %s (%s)\n\tLine %d\n\tEntry: %s\n", i, m.ListFileName, m.ListURL, m.LineNumber, m.LineText)
		}
	}
	if len(result.ApprxBLMatches) > 0 {
		str += "\nAPPROXIMATE MATCHES (may result in blocks of the query):\n"
		for i, m := range result.ApprxBLMatches {
			str += fmt.Sprintf("%02d) %s (%s)\n\tLine %d\n\tEntry: %s\n", i, m.ListFileName, m.ListURL, m.LineNumber, m.LineText)
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

	fmt.Printf("QUERY: %s\nSearching. This may take a while depending on the number and size of your blocklists...\n", os.Args[1])
	search_results := searchForURLInAllLists(os.Args[1])
	fmt.Println(stringifyResults(search_results))
}
