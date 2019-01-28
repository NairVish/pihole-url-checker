package main

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

// for sort.Sort-ing BlacklistPriority objects
type BLPriorities []BlacklistPriority

func (blp *BLPriorities) Len() int      { return len(*blp) }
func (blp *BLPriorities) Swap(i, j int) { (*blp)[i], (*blp)[j] = (*blp)[j], (*blp)[i] }
func (blp *BLPriorities) Less(i, j int) bool {
	return (*blp)[i].Priority < (*blp)[j].Priority
}
