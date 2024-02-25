//go:build goexperiment.rangefunc

package trietree

import (
	"iter"
	"unicode/utf8"
)

type Prediction struct {
	Index int
	Label rune
	ID    int
	Depth int
}

func (dt *DTree) Predict(s string) iter.Seq[Prediction] {
	var (
		query    = s
		idx      = 0
		pivot    = &dt.Root // won't be nil
		nextNode = func() (*DNode, int, rune) {
			x := idx
			r, sz := utf8.DecodeRuneInString(query)
			query = query[sz:]
			idx += sz
			pivot = dt.nextNode(pivot, r)
			return pivot, x, r
		}
	)

	return func(yield func(Prediction) bool) {
		var (
			currNode *DNode = nil
			currIdx  int
			currRune rune
		)
		for {
			if currNode == nil {
				if query == "" {
					return
				}
				currNode, currIdx, currRune = nextNode()
				//log.Printf("update: cx=%d cr=%c cn=%v", currIdx, currRune, currNode)
			}
			if currNode.EdgeID > 0 {
				//log.Printf("yield: cx=%d cr=%c cn=%v", currIdx, currRune, currNode)
				if !yield(Prediction{Index: currIdx, Label: currRune, ID: currNode.EdgeID, Depth: currNode.Level}) {
					currNode = nil
					query = ""
					return
				}
			}
			currNode = currNode.Failure
		}
	}
}

func (st *STree) Predict(s string) iter.Seq[Prediction] {
	var (
		query    = s
		idx      = 0
		pivot    = 0
		nextNode = func() (nodeID int, nodeIdx int, nodeRune rune) {
			nodeIdx = idx
			nodeRune, sz := utf8.DecodeRuneInString(query)
			query = query[sz:]
			idx += sz
			pivot = st.nextNode(pivot, nodeRune)
			return pivot, nodeIdx, nodeRune
		}
	)
	return func(yield func(Prediction) bool) {
		var (
			currNode int = 0
			currIdx  int
			currRune rune
		)
		for {
			if currNode == 0 {
				if query == "" {
					return
				}
				currNode, currIdx, currRune = nextNode()
			}
			n := st.Nodes[currNode]
			if id := n.EdgeID; id > 0 {
				if !yield(Prediction{Index: currIdx, Label: currRune, ID: id, Depth: st.Levels[id-1]}) {
					currNode = 0
					query = ""
					return
				}
			}
			currNode = n.Fail
		}
	}
}
