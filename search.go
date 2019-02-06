package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// SearchObj represents the search itself, including its methods and members.
type SearchObj struct {
	piholeListRootFolder string       // the Pi-hole root folder specified
	blocklistURLs        []string     // source URLs of all active blocklists
	blocklistFileNames   []string     // downloaded filenames of all active blocklists
	FinalResult          *TotalResult // the final search result
}

// NewSearchObj instantiates a new SearchObj by populating all of the list URLs and file names.
func NewSearchObj(piholeListRoot string) *SearchObj {
	so := SearchObj{piholeListRootFolder: piholeListRoot}
	so.populateAllListURLs()
	so.populateAllBlocklistFileNames()

	if len(so.blocklistURLs) != len(so.blocklistFileNames) {
		log.Fatal(fmt.Printf("len(all_list_urls) [%d] != len(all_list_filenames) [%d]", len(so.blocklistURLs), len(so.blocklistFileNames)))
	}

	return &so
}

// SearchObj.populateAllListURLs reads "adlists.list" and stores the URLs of all active blocklists.
func (so *SearchObj) populateAllListURLs() {
	list_urls_filename := filepath.Join(so.piholeListRootFolder, "adlists.list")
	full_url_list, err := os.Open(list_urls_filename)
	logFatalIfError(err)
	defer full_url_list.Close()

	url_list := make([]string, 0)
	url_list = append(url_list, "black.list") // append the user-defined blacklist first
	file_scanner := bufio.NewScanner(full_url_list)
	for file_scanner.Scan() { // for each line in the URL list...
		this_url := strings.TrimSpace(file_scanner.Text())          // trim any surrounding whitespace
		if !strings.HasPrefix(this_url, "#") && len(this_url) > 0 { // if the URL isn't commented out or an empty line...
			url_list = append(url_list, this_url)
		}
	}

	// if the file_scanner encountered an error, it will have exited the for loop prematurely and released an error
	err = file_scanner.Err()
	logFatalIfError(err)

	so.blocklistURLs = url_list
}

// SearchObj.populateAllBlocklistFileNames parses and stores the filename for each active blocklist.
func (so *SearchObj) populateAllBlocklistFileNames() {
	filenames, err := ioutil.ReadDir(so.piholeListRootFolder) // read/list the contents of the root directory
	logFatalIfError(err)

	filelist := make(BLPriorities, 0)
	filelist = append(filelist, BlacklistPriority{Priority: -1, Filename: "black.list"}) // append the user-defined blacklist first
	for _, f := range filenames {
		fname := strings.TrimSpace(f.Name())

		if f.IsDir() || len(fname) == 0 {
			continue
		}

		// downloaded pi-hole blocklists have a name of the format: list.##.SOURCE_URL.domains
		// the number corresponds to its position in adlists.list...we will use this later to sort the listnames
		//	which can't be sorted normally because the numbers in the file name are not zero-padded
		if strings.HasPrefix(fname, "list.") && strings.HasSuffix(fname, ".domains") {
			num, err := strconv.Atoi(strings.Split(fname, ".")[1])
			logFatalIfError(err)
			filelist = append(filelist, BlacklistPriority{Priority: num, Filename: fname})
		}
	}

	// using sort.Sort instead of sort.Slice b/c the default go version on Raspbian is 1.7.4 (< 1.8)
	sort.Sort(&filelist) // sort by priority (i.e., the number in the filename)

	// extract just the sorted filenames from the resulting slice
	final_flist := make([]string, 0)
	for _, bp := range filelist {
		final_flist = append(final_flist, bp.Filename)
	}

	so.blocklistFileNames = final_flist
}

// SearchObj.SearchForURLInAllLists executes a search for `query` in all active blocklists.
func (so *SearchObj) SearchForURLInAllLists(query string) {
	exact_matches := make([]ListResult, 0)
	approx_matches := make([]ListResult, 0)
	for i := range so.blocklistURLs { // for each blocklist URL...
		// start reading its corresponding file in the pi-hole root
		this_list_filename := filepath.Join(so.piholeListRootFolder, so.blocklistFileNames[i])
		this_list_file, err := os.Open(this_list_filename)
		logFatalIfError(err)
		file_scanner := bufio.NewScanner(this_list_file)

		// scan the file line-by-line (TODO: Any way to speed this up?)
		for j := 0; file_scanner.Scan(); j += 1 {
			this_entry := strings.TrimSpace(file_scanner.Text())
			orig_entry := strings.TrimSpace(file_scanner.Text())
			// ignore if the whole line is a comment (TODO: Handle lines where comments are at the end of the line, need URL still)
			if strings.HasPrefix(this_entry, "#") || len(this_entry) == 0 {
				continue
			}

			// if the line starts with an ip address, trim it off
			if ipAddrRegex.MatchString(this_entry) {
				this_entry = ipAddrRegex.ReplaceAllLiteralString(this_entry, "")
				this_entry = strings.TrimSpace(this_entry)
			}

			// if the part of the line that's left matches the query exactly. append as an exact match
			if query == this_entry {
				exact_matches = append(exact_matches, ListResult{LineNumber: j, ListFileName: so.blocklistFileNames[i], ListURL: so.blocklistURLs[i], LineText: orig_entry})
			} else if strings.Contains(this_entry, query) {
				// else if the part of the line that's left contains the query (maybe as part of a subdomain),
				// 	append as an approximate match
				approx_matches = append(approx_matches, ListResult{LineNumber: j, ListFileName: so.blocklistFileNames[i], ListURL: so.blocklistURLs[i], LineText: orig_entry})
			}
		}

		err = file_scanner.Err()
		logFatalIfError(err)
		this_list_file.Close()
	}

	so.FinalResult = &TotalResult{QueryURL: query, ExactBLMatches: exact_matches, ApprxBLMatches: approx_matches}
}

// SearchObj.StringifyResults converts the results into string form for eventual printing.
func (so *SearchObj) StringifyResults() string {
	if len(so.FinalResult.ExactBLMatches) == 0 && len(so.FinalResult.ApprxBLMatches) == 0 {
		return "No results found in existing, active blocklists."
	}

	str := fmt.Sprintf("\nFound %d exact match(es) and %d approximate match(es).\n", len(so.FinalResult.ExactBLMatches), len(so.FinalResult.ApprxBLMatches))
	if len(so.FinalResult.ExactBLMatches) > 0 {
		str += "\nEXACT MATCHES:\n"
		for i, m := range so.FinalResult.ExactBLMatches {
			str += fmt.Sprintf("%02d) %s (%s)\n\tLine %d\n\tEntry: %s\n", i, m.ListFileName, m.ListURL, m.LineNumber, m.LineText)
		}
	}
	if len(so.FinalResult.ApprxBLMatches) > 0 {
		str += "\nAPPROXIMATE MATCHES (may potentially contribute to blocks of the query):\n"
		for i, m := range so.FinalResult.ApprxBLMatches {
			str += fmt.Sprintf("%02d) %s (%s)\n\tLine %d\n\tEntry: %s\n", i, m.ListFileName, m.ListURL, m.LineNumber, m.LineText)
		}
	}

	return str
}
