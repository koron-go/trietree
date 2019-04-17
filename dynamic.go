package trietree

import (
	"context"
)

// DTree is dynamic tree.
type DTree struct {
	Root DNode

	lastEdgeID int
}

// DNode is a node of dynamic tree.
type DNode struct {
	Label rune

	Low   *DNode
	High  *DNode
	Child *DNode

	EdgeID int

	Failure *DNode
}

func (dn *DNode) dig(c rune) *DNode {
	p := dn.Child
	if p == nil {
		dn.Child = &DNode{Label: c}
		return dn.Child
	}
	for {
		if c == p.Label {
			return p
		}
		if c < p.Label {
			if p.Low == nil {
				p.Low = &DNode{Label: c}
				return p.Low
			}
			p = p.Low
		} else {
			if p.High == nil {
				p.High = &DNode{Label: c}
				return p.High
			}
			p = p.High
		}
	}
}

// Get obtains an existing child node for rune.
func (dn *DNode) Get(r rune) *DNode {
	p := dn.Child
	for p != nil {
		if r == p.Label {
			return p
		}
		if r < p.Label {
			p = p.Low
		} else {
			p = p.High
		}
	}
	return nil
}

// Put puts an edige for key and emits ID for it. ID will be greater than zero.
func (dt *DTree) Put(k string) int {
	n := &dt.Root
	for _, r := range k {
		n = n.dig(r)
	}
	if n.EdgeID <= 0 {
		dt.lastEdgeID++
		n.EdgeID = dt.lastEdgeID
	}
	return n.EdgeID
}

// Scan is a wrapper for ScanContext with context.Background().
func (dt *DTree) Scan(s string, r ScanReporter) error {
	return dt.ScanContext(context.Background(), s, r)
}

// ScanContext scans a string to find matched words.
// ScanReporter r will receive reports for each characters when scan.
func (dt *DTree) ScanContext(ctx context.Context, s string, r ScanReporter) error {
	sr := newScanReport(r, len(s))
	curr := &dt.Root
	//fmt.Printf("ScanContext: %q\n", s)
	for i, c := range s {
		//fmt.Printf("  i=%d c=%c curr=%p%+[3]v\n", i, c, curr)
		next := dt.nextNode(curr, c)
		//fmt.Printf("    next=%p%+[1]v found=%t isRoot=%t\n", next, found, isRoot)
		sr.reportDynamic(i, c, next)
		if err := ctx.Err(); err != nil {
			return err
		}
		curr = next
	}
	return nil
}

func (dt *DTree) nextNode(curr *DNode, c rune) *DNode {
	root := &dt.Root
	for {
		next := curr.Get(c)
		if next != nil {
			return next
		}
		if curr == root {
			return root
		}
		curr = curr.Failure
		if curr == nil {
			curr = root
		}
	}
}

type procDNode func(*DNode)

func (dn *DNode) eachSiblings(fn procDNode) {
	if dn == nil {
		return
	}
	dn.Low.eachSiblings(fn)
	fn(dn)
	dn.High.eachSiblings(fn)
}

// Get retrieve a node for key, otherwise returns nil.
func (dt *DTree) Get(k string) *DNode {
	n := &dt.Root
	for _, r := range k {
		n = n.Get(r)
		if n == nil {
			return nil
		}
	}
	return n
}

// FillFailure fill Failure field with Aho-Corasick algorithm.
func (dt *DTree) FillFailure() {
	dt.Root.Failure = &dt.Root
	dt.fillFailure(&dt.Root)
	dt.Root.Failure = nil
}

func (dt *DTree) fillFailure(parent *DNode) {
	if parent.Child == nil {
		return
	}
	//fmt.Printf("fillFailure: parent(%p)=%+[1]v\n", parent)
	pf := parent.Failure
	parent.Child.eachSiblings(func(curr *DNode) {
		f := dt.nextNode(pf, curr.Label)
		if f == curr {
			f = &dt.Root
		}
		curr.Failure = f
		//fmt.Printf("  curr(%p)=%+[1]v\n", curr)
		dt.fillFailure(curr)
	})
}

// CountChild counts child nodes.
func (dn *DNode) CountChild() int {
	c := 0
	dn.Child.eachSiblings(func(*DNode) { c++ })
	return c
}

// CountAll counts all descended nodes.
func (dn *DNode) CountAll() int {
	c := 1
	dn.Child.eachSiblings(func(n *DNode) {
		c += n.CountAll()
	})
	return c
}
