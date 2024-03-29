package trietree

import (
	"context"
	"errors"
	"io"
	"math"
	"sort"
	"unicode/utf8"
)

// STree is static tree. It is optimized for serialization.
type STree struct {
	Nodes  []SNode
	Levels []int
}

// Freeze converts dynamic tree to static tree.
func Freeze(src *DTree) *STree {
	nall := src.Root.CountAll()
	nodes := make([]SNode, nall)
	levels := make([]int, src.lastEdgeID)
	z := 1

	var procNode func(x int, dn *DNode, i int)
	procNode = func(x int, dn *DNode, i int) {
		//fmt.Printf("x=%d dn={Label:%c EdgeID:%d Fail:%p} i=%d\n", x, dn.Label, dn.EdgeID, dn.Failure, i)
		start, end := 0, 0
		if n := dn.CountChild(); n > 0 {
			start, end = z, z+n
			z = end
		}
		nodes[x] = SNode{
			Label:  dn.Label,
			Start:  start,
			End:    end,
			EdgeID: dn.EdgeID,
		}
		if dn.EdgeID > 0 {
			levels[dn.EdgeID-1] = dn.Level
		}
		if dn.Child != nil {
			j := i + 1
			y := start
			dn.Child.eachSiblings(func(dn2 *DNode) {
				procNode(y, dn2, j)
				y++
			})
		}
	}

	procNode(0, &src.Root, 0)
	st := &STree{
		Nodes:  nodes,
		Levels: levels,
	}
	st.fillFailure(0)

	return st
}

func (st *STree) fillFailure(x int) {
	p := &st.Nodes[x]
	if p.Start == 0 {
		return
	}
	for i := p.Start; i < p.End; i++ {
		c := &st.Nodes[i]
		c.Fail = st.nextNode(p.Fail, c.Label)
		if c.Fail == i {
			c.Fail = 0
		}
		st.fillFailure(i)
	}
}

// Scan scans a string to find matched words.
func (st *STree) Scan(s string, r ScanReporter) error {
	return st.ScanContext(context.Background(), s, r)
}

// ScanContext scans a string to find matched words.
// ScanReporter r will receive reports for each characters when scan.
func (st *STree) ScanContext(ctx context.Context, s string, r ScanReporter) error {
	sr := newScanReport(r, len(s))
	curr := 0
	for i, c := range s {
		next := st.nextNode(curr, c)
		// emit a scan event.
		sr.reset(i, c)
		for n := next; n > 0; n = st.Nodes[n].Fail {
			if edge := st.Nodes[n].EdgeID; edge > 0 {
				lv := -1
				if edge-1 < len(st.Levels) {
					lv = st.Levels[edge-1]
				}
				sr.add(edge, lv)
			}
		}
		sr.emit()
		// prepare for next.
		if err := ctx.Err(); err != nil {
			return err
		}
		curr = next
	}
	return nil
}

func (st *STree) nextNode(x int, c rune) int {
	for {
		n := &st.Nodes[x]
		next := st.find(n.Start, n.End, c)
		if next >= 0 {
			return next
		}
		if x == 0 {
			return 0
		}
		x = n.Fail
	}
}

func (st *STree) find(a, b int, c rune) int {
	x := a + sort.Search(b-a, func(n int) bool {
		return st.Nodes[n+a].Label >= c
	})
	if x >= a && x < b && st.Nodes[x].Label == c {
		return x
	}
	return -1
}

// LongestPrefix finds a longest prefix node/edge matches given s string.
func (st *STree) LongestPrefix(s string) (prefix string, edgeID int) {
	last := -1
	ilast := 0
	curr := 0
	for i, r := range s {
		n := st.Nodes[curr]
		next := st.find(n.Start, n.End, r)
		if next < 0 {
			break
		}
		if st.Nodes[next].EdgeID > 0 {
			last = next
			ilast = i
		}
		curr = next
	}
	if last < 0 {
		return "", 0
	}
	n := st.Nodes[last]
	return s[:ilast+utf8.RuneLen(n.Label)], n.EdgeID
}

// Write serializes a tree to io.Writer.
func (st *STree) Write(w io.Writer) error {
	ww := newWriter(w)

	// write nodes.
	ww.writeInt(len(st.Nodes))
	if ww.err != nil {
		return ww.err
	}
	for _, n := range st.Nodes {
		err := n.write(ww)
		if err != nil {
			return err
		}
	}

	// write levels.
	ww.writeInt(len(st.Levels))
	if ww.err != nil {
		return ww.err
	}
	for _, lv := range st.Levels {
		ww.writeInt(lv)
	}
	if ww.err != nil {
		return ww.err
	}

	ww.w.Flush()
	return nil
}

const intSize = 32 << (^uint(0) >> 63)

// Read reads static tree from io.Reader.
func Read(r io.Reader) (*STree, error) {
	rr := newReader(r)

	// read nodes.
	n, err := rr.readInt64()
	if err != nil {
		return nil, err
	}
	// check 32 bit overflow.
	if intSize == 32 && n > math.MaxInt32 {
		return nil, errors.New("too large tree for 32bit architecture")
	}
	nodes := make([]SNode, int(n))
	for i := range nodes {
		err := nodes[i].read(rr)
		if err != nil {
			return nil, err
		}
	}

	// read levels.
	n, err = rr.readInt64()
	if err != nil {
		return nil, err
	}
	if intSize == 32 && n > math.MaxInt32 {
		return nil, errors.New("too large levels for 32bit architecture")
	}
	levels := make([]int, int(n))
	for i := range levels {
		levels[i] = rr.readInt()
	}
	if rr.err != nil {
		return nil, rr.err
	}

	return &STree{
		Nodes:  nodes,
		Levels: levels,
	}, nil
}

// SNode is a node for static tree.
type SNode struct {
	Label  rune
	Start  int // start index for children (inclusive)
	End    int // end index of children (exclusive)
	Fail   int // index to failure node
	EdgeID int
}

func (sn SNode) write(w *writer) error {
	w.writeRune(sn.Label)
	w.writeInt(sn.Start)
	w.writeInt(sn.End)
	w.writeInt(sn.Fail)
	w.writeInt(sn.EdgeID)
	if w.err != nil {
		return w.err
	}
	return nil
}

func (sn *SNode) read(r *reader) error {
	sn.Label = r.readRune()
	sn.Start = r.readInt()
	sn.End = r.readInt()
	sn.Fail = r.readInt()
	sn.EdgeID = r.readInt()
	if r.err != nil {
		return r.err
	}
	return nil
}
