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

type SearchObj struct {
	piholeListRootFolder string
	blocklistURLs        []string
	blocklistFileNames   []string
	FinalResult          *TotalResult
}

func NewSearchObj(piholeListRoot string) *SearchObj {
	so := SearchObj{piholeListRootFolder: piholeListRoot}
	so.populateAllListURLs()
	so.populateAllBlocklistFileNames()

	if len(so.blocklistURLs) != len(so.blocklistFileNames) {
		log.Fatal(fmt.Printf("len(all_list_urls) [%d] != len(all_list_filenames) [%d]", len(so.blocklistURLs), len(so.blocklistFileNames)))
	}

	return &so
}

func (so *SearchObj) populateAllListURLs() {
	list_urls_filename := filepath.Join(so.piholeListRootFolder, "adlists.list")
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

	so.blocklistURLs = url_list
}

func (so *SearchObj) populateAllBlocklistFileNames() {
	filenames, err := ioutil.ReadDir(so.piholeListRootFolder)
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
	sort.Sort(&filelist)

	final_flist := make([]string, 0)
	for _, bp := range filelist {
		final_flist = append(final_flist, bp.Filename)
	}

	so.blocklistFileNames = final_flist
}

func (so *SearchObj) SearchForURLInAllLists(query string) {
	exact_matches := make([]ListResult, 0)
	approx_matches := make([]ListResult, 0)
	for i := range so.blocklistURLs {
		this_list_filename := so.blocklistFileNames[i]
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
				exact_matches = append(exact_matches, ListResult{LineNumber: j, ListFileName: this_list_filename, ListURL: so.blocklistURLs[i], LineText: orig_entry})
			} else if strings.Contains(this_entry, query) {
				approx_matches = append(approx_matches, ListResult{LineNumber: j, ListFileName: this_list_filename, ListURL: so.blocklistURLs[i], LineText: orig_entry})
			}
		}

		err = file_scanner.Err()
		logFatalIfError(err)
		this_list_file.Close()
	}

	so.FinalResult = &TotalResult{QueryURL: query, ExactBLMatches: exact_matches, ApprxBLMatches: approx_matches}
}

func (so *SearchObj) StringifyResults() string {
	if len(so.FinalResult.ExactBLMatches) == 0 && len(so.FinalResult.ApprxBLMatches) == 0 {
		return "No results found in existing, active blocklists."
	}

	str := fmt.Sprintf("\nFound %d exact matches and %d approximate matches.\n", len(so.FinalResult.ExactBLMatches), len(so.FinalResult.ApprxBLMatches))
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
