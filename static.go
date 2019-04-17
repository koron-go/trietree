package trietree

import (
	"context"
	"errors"
	"io"
	"math"
	"sort"
)

// STree is static tree. It is optimized for serialization.
type STree struct {
	Nodes []SNode
}

// Freeze converts dynamic tree to static tree.
func Freeze(src *DTree) *STree {
	nall := src.Root.CountAll()
	nodes := make([]SNode, nall)
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
	st := &STree{Nodes: nodes}
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

// Scan is a wrapper for ScanContext with context.Background().
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
		sr.reportStatic(i, c, next, st.Nodes)
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

// Write serializes a tree to io.Writer.
func (st *STree) Write(w0 io.Writer) error {
	w := newWriter(w0)
	w.writeInt(len(st.Nodes))
	if w.err != nil {
		return w.err
	}
	for _, n := range st.Nodes {
		err := n.write(w)
		if err != nil {
			return err
		}
	}
	w.w.Flush()
	return nil
}

const intSize = 32 << (^uint(0) >> 63)

// Read reads static tree from io.Reader.
func Read(r0 io.Reader) (*STree, error) {
	r := newReader(r0)
	n, err := r.readInt64()
	if err != nil {
		return nil, err
	}
	// check 32 bit overflow.
	if intSize == 32 && n > math.MaxInt32 {
		return nil, errors.New("too large tree for 32bit architecture")
	}
	nodes := make([]SNode, int(n))
	for i := range nodes {
		err := nodes[i].read(r)
		if err != nil {
			return nil, err
		}
	}
	return &STree{
		Nodes: nodes,
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
