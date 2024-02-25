//go:build goexperiment.rangefunc

package trietree

import (
	"iter"
)

// Predict returns an iterator which enumerates Prediction: key suggestions
// that match the query in the tree.
func (dt *DTree) Predict(query string) iter.Seq[Prediction] {
	return predict[*DNode](dt, query)
}

// Predict returns an iterator which enumerates Prediction: key suggestions
// that match the query in the tree.
func (st *STree) Predict(query string) iter.Seq[Prediction] {
	return predict[int](st, query)
}

func predict[T comparable](tree predictableTree[T], query string) iter.Seq[Prediction] {
	var zero T
	tr := newTraverser[T](tree, query)
	return func(yield func(Prediction) bool) {
		for {
			node, end := tr.next()
			if node == zero {
				return
			}
			for node != zero {
				if id := tree.nodeId(node); id > 0 {
					st := trailingIndex(query[:end], tree.nodeLevel(node))
					if !yield(Prediction{Start: st, End: end, ID: id}) {
						tr.close()
						return
					}
				}
				node = tree.nodeFail(node)
			}
		}
	}
}
