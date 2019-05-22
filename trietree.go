package trietree

// ScanEvent is an event which detected in Scan/ScanContext.
type ScanEvent struct {
	Index int
	Label rune
	Nodes []ScanNode
}

// ScanNode is scanned node information.
type ScanNode struct {
	ID    int
	Level int
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
	r  ScanReporter
	ev ScanEvent

	nodesBuf  []ScanNode
	nodesCurr []ScanNode
}

func newScanReport(r ScanReporter, n int) *scanReport {
	return &scanReport{
		r:        r,
		nodesBuf: make([]ScanNode, n),
	}
}

func (sr *scanReport) reset(i int, c rune) {
	sr.ev.Index = i
	sr.ev.Label = c
	sr.nodesCurr = sr.nodesBuf[:0]
}

func (sr *scanReport) add(id int, level int) {
	sr.nodesCurr = append(sr.nodesCurr, ScanNode{ID: id, Level: level})
}

func (sr *scanReport) emit() {
	nodes := sr.nodesCurr
	if len(nodes) == 0 {
		nodes = nil
	}
	sr.ev.Nodes = nodes
	sr.r.ScanReport(sr.ev)
}
