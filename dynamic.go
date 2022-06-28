package trietree

import (
	"context"
	"unicode/utf8"
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
	Level  int

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
	level := 0
	for _, r := range k {
		n = n.dig(r)
		level++
	}
	if n.EdgeID <= 0 {
		dt.lastEdgeID++
		n.EdgeID = dt.lastEdgeID
	}
	n.Level = level
	return n.EdgeID
}

// Scan scans a string to find matched words.
func (dt *DTree) Scan(s string, r ScanReporter) error {
	return dt.ScanContext(context.Background(), s, r)
}

// ScanContext scans a string to find matched words.
// ScanReporter r will receive reports for each characters when scan.
func (dt *DTree) ScanContext(ctx context.Context, s string, r ScanReporter) error {
	sr := newScanReport(r, len(s))
	curr := &dt.Root
	for i, c := range s {
		next := dt.nextNode(curr, c)
		// emit a scan event.
		sr.reset(i, c)
		for n := next; n != nil; n = n.Failure {
			if n.EdgeID > 0 {
				sr.add(n.EdgeID, n.Level)
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

// LongestPrefix finds a longest prefix against given s string.
func (dt *DTree) LongestPrefix(s string) (prefix string, edgeID int) {
	var last *DNode
	ilast := 0
	curr := &dt.Root
	for i, r := range s {
		next := curr.Get(r)
		if next == nil {
			break
		}
		if next.EdgeID > 0 {
			last = next
			ilast = i
		}
		curr = next
	}
	if last == nil {
		return "", 0
	}
	return s[:ilast+utf8.RuneLen(last.Label)], last.EdgeID
}
