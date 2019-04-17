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
	r  ScanReporter
	ev ScanEvent

	idBuf  []int
	idCurr []int
}

func newScanReport(r ScanReporter, n int) *scanReport {
	return &scanReport{
		r:     r,
		idBuf: make([]int, n),
	}
}

func (sr *scanReport) reset(i int, c rune) {
	sr.ev.Index = i
	sr.ev.Label = c
	sr.idCurr = sr.idBuf[:0]
}

func (sr *scanReport) addID(id int) {
	sr.idCurr = append(sr.idCurr, id)
}

func (sr *scanReport) emit() {
	ids := sr.idCurr
	if len(ids) == 0 {
		ids = nil
	}
	sr.ev.IDs = ids
	sr.r.ScanReport(sr.ev)
}
