package trie2

import "github.com/koron-go/trietree"

type Prediction[T any] struct {
	Start int    // Start is the start index of Key in the query.
	End   int    // End is the end index of Key in the query.
	Key   string // Key is matched string.
	Value T      // Value is the value corresponding to the key.
}

// PredictionIter is the iterator of Prediction.
type PredictionIter[T any] func() *Prediction[T]

func predictIter[T any](query string, iter trietree.PredictionIter, values []T) func() *Prediction[T] {
	return func() *Prediction[T] {
		p := iter()
		if p == nil {
			return nil
		}
		return &Prediction[T]{
			Start: p.Start,
			End:   p.End,
			Key:   query[p.Start:p.End],
			Value: values[p.ID-1],
		}
	}
}

func (dt *DTrie[T]) PredictIter(query string) PredictionIter[T] {
	return predictIter(query, dt.tree.PredictIter(query), dt.values)
}

func (st *STrie[T]) PredictIter(query string) PredictionIter[T] {
	return predictIter(query, st.tree.PredictIter(query), st.values)
}
