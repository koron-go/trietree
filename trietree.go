package trietree

// Reporter will receive report of Scan/ScanCotext.
type Reporter interface {
	// Report is calleback when scan each runes.
	// i is index, c is a rune, id is EdgeID which returned when added and
	// viaRoot will be true when reseted match status (node traverse was
	// resetted to root).
	Report(i int, c rune, viaRoot bool, ids []int)
}

// ReportFunc is a utility type to implement Reporter.
type ReportFunc func(int, rune, bool, []int)

// Report implements a method for Reporter.
func (f ReportFunc) Report(i int, c rune, viaRoot bool, ids []int) {
	f(i, c, viaRoot, ids)
}

type reportWrapper struct {
	r     Reporter
	idbuf []int
}

func newReportWrapper(r Reporter, n int) *reportWrapper {
	return &reportWrapper{r: r, idbuf: make([]int, 0, n)}
}

func (rw *reportWrapper) reportDynamic(i int, c rune, viaRoot bool, n *DNode) {
	ids := rw.idbuf[:0]
	for n != nil {
		if n.EdgeID > 0 {
			ids = append(ids, n.EdgeID)
		}
		n = n.Failure
	}
	if len(ids) == 0 {
		ids = nil
	}
	rw.r.Report(i, c, viaRoot, ids)
}

func (rw *reportWrapper) reportStatic(i int, c rune, viaRoot bool, n int, nodes []SNode) {
	ids := rw.idbuf[:0]
	for n > 0 {
		edge := nodes[n].EdgeID
		if edge > 0 {
			ids = append(ids, edge)
		}
		n = nodes[n].Fail
	}
	if len(ids) == 0 {
		ids = nil
	}
	rw.r.Report(i, c, viaRoot, ids)
}
