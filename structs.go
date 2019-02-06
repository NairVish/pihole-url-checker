package main

// ListResult holds a single search match.
type ListResult struct {
	LineNumber   int    // the line number in the file where the match was made
	ListFileName string // the file name
	ListURL      string // the source URL of the file
	LineText     string // the original text of the line where the match was made
}

// TotalResult represents the final search result.
type TotalResult struct {
	QueryURL       string       // the original query URL
	ExactBLMatches []ListResult // a list of all exact matches
	ApprxBLMatches []ListResult // a list of all approximate matches
}

// BlacklistPriority represents a single blocklist and is used to sort blocklists according to their position
// 	in the list of blocklists (by the numbers in their filenames).
type BlacklistPriority struct {
	Priority int
	Filename string
}

// for sort.Sort-ing BlacklistPriority objects
type BLPriorities []BlacklistPriority

func (blp *BLPriorities) Len() int      { return len(*blp) }
func (blp *BLPriorities) Swap(i, j int) { (*blp)[i], (*blp)[j] = (*blp)[j], (*blp)[i] }
func (blp *BLPriorities) Less(i, j int) bool {
	return (*blp)[i].Priority < (*blp)[j].Priority
}
