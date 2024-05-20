package trietree

import (
	"unicode/utf8"
)

// Prediction is identifier of a key.
type Prediction struct {
	Start int // Start is start index of key in query.
	End   int // End is end index of key in query.
	ID    int // ID is for edge node identifier.
}

// PredictIter is the iterator of Prediction.
type PredictIter func() *Prediction

// PredictIter returns an iterator function PredictIter, which enumerates
// Prediction: key suggestions that match the query in the tree.
func (dt *DTree) PredictIter(query string) PredictIter {
	return predictIter[*DNode](dt, query)
}

// PredictIter returns an iterator function PredictIter, which enumerates
// Prediction: key suggestions that match the query in the tree.
func (st *STree) PredictIter(query string) PredictIter {
	return predictIter[int](st, query)
}

type predictableTree[T comparable] interface {
	root() T
	nextNode(T, rune) T
	nodeId(T) int
	nodeLevel(T) int
	nodeFail(T) T
}

// methods DTree satisfies predictableTree[*DNode]
func (dt *DTree) root() *DNode             { return &dt.Root }
func (dt *DTree) nodeId(n *DNode) int      { return n.EdgeID }
func (dt *DTree) nodeLevel(n *DNode) int   { return n.Level }
func (dt *DTree) nodeFail(n *DNode) *DNode { return n.Failure }

// methods STree satisfies predictableTree[int]
func (st *STree) root() int           { return 0 }
func (st *STree) nodeId(n int) int    { return st.Nodes[n].EdgeID }
func (st *STree) nodeLevel(n int) int { return st.Levels[st.nodeId(n)-1] }
func (st *STree) nodeFail(n int) int  { return st.Nodes[n].Fail }

type traverser[T comparable] struct {
	tree  predictableTree[T]
	query string
	pivot T
	index int
}

func newTraverser[T comparable](tree predictableTree[T], query string) traverser[T] {
	return traverser[T]{
		tree:  tree,
		query: query,
		pivot: tree.root(),
		index: 0,
	}
}

// next consumes a rune from query, and determine next node to travese tree.
// this returns next node, and tail index of last parsed rune in query.
func (tr *traverser[T]) next() (node T, end int, valid bool) {
	var zero T
	if tr.query == "" {
		return zero, 0, false
	}
	r, sz := utf8.DecodeRuneInString(tr.query)
	if sz == 0 {
		return zero, 0, false
	}
	tr.query = tr.query[sz:]
	tr.index += sz
	tr.pivot = tr.tree.nextNode(tr.pivot, r)
	return tr.pivot, tr.index, true
}

func (tr *traverser[T]) close() {
	tr.query = ""
}

// trailingIndex returns the index of the n'th character from the end of string s.
func trailingIndex(s string, n int) int {
	x := len(s)
	for n > 0 && x > 0 {
		_, sz := utf8.DecodeLastRuneInString(s[:x])
		if sz == 0 {
			break
		}
		x -= sz
		n--
	}
	return x
}

func predictIter[T comparable](tree predictableTree[T], query string) func() *Prediction {
	var (
		zero T
		tr   = newTraverser[T](tree, query)
		req  = true
		node T
		end  int
	)
	return func() *Prediction {
		var p *Prediction
		for p == nil {
			if req {
				var valid bool
				node, end, valid = tr.next()
				if !valid {
					tr.close()
					return nil
				}
				req = false
			}
			for !req && p == nil {
				id := tree.nodeId(node)
				if id > 0 {
					st := trailingIndex(query[:end], tree.nodeLevel(node))
					p = &Prediction{Start: st, End: end, ID: id}
				}
				node = tree.nodeFail(node)
				if id == 0 && node == zero {
					req = true
				}
			}
		}
		return p
	}
}
