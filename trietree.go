package trietree

// ScanEvent is an event which detected in Scan/ScanContext.
type ScanEvent struct {
	Index int
	Label rune
	IDs   []int
}

// ScanReporter receive reports of scan.
type ScanReporter interface {
	ScanReport(ev ScanEvent)
}

// ScanReportFunc is a utility type to implements ScanReporter.
type ScanReportFunc func(ev ScanEvent)

// ScanReport implements a method of ScanReporter.
func (f ScanReportFunc) ScanReport(ev ScanEvent) {
	f(ev)
}

type scanReport struct {
	r   ScanReporter
	ids []int
}

func newScanReport(r ScanReporter, n int) *scanReport {
	return &scanReport{
		r:   r,
		ids: make([]int, n),
	}
}

func (sr *scanReport) idsOrNil( ids []int) []int{
	if len(ids) == 0 {
		return nil
	}
	return ids
}

func (sr *scanReport) reportDynamic(i int, c rune, n *DNode) {
	ids := sr.ids[:0]
	for n != nil {
		if n.EdgeID > 0 {
			ids = append(ids, n.EdgeID)
		}
		n = n.Failure
	}
	sr.r.ScanReport(ScanEvent{
		Index: i,
		Label: c,
		IDs:   sr.idsOrNil(ids),
	})
}

func (sr *scanReport) reportStatic(i int, c rune, n int, nodes []SNode) {
	ids := sr.ids[:0]
	for n > 0 {
		edge := nodes[n].EdgeID
		if edge > 0 {
			ids = append(ids, edge)
		}
		n = nodes[n].Fail
	}
	sr.r.ScanReport(ScanEvent{
		Index: i,
		Label: c,
		IDs:   sr.idsOrNil(ids),
	})
}
