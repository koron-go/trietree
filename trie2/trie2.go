package trie2

import (
	"encoding/gob"
	"fmt"
	"io"

	"github.com/koron-go/trietree"
)

type DTrie[T any] struct {
	tree   trietree.DTree
	values []T
}

func (dt *DTrie[T]) FillFailure() {
	dt.tree.FillFailure()
}

func (dt *DTrie[T]) Put(k string, v T) {
	id := dt.tree.Put(k)
	if id-1 == len(dt.values) {
		dt.values = append(dt.values, v)
		return
	}
	dt.values[id-1] = v
}

func (dt *DTrie[T]) LongestPrefix(s string) (v T, prefix string, ok bool) {
	prefix, id := dt.tree.LongestPrefix(s)
	if id == 0 {
		var zero T
		return zero, "", false
	}
	return dt.values[id-1], prefix, true
}

type STrie[T any] struct {
	tree   trietree.STree
	values []T
}

func (dt *DTrie[T]) Freeze(copyValues bool) *STrie[T] {
	tree := trietree.Freeze(&dt.tree)
	var values []T
	if copyValues {
		values = make([]T, len(values))
		copy(values, dt.values)
	} else {
		values = dt.values
	}
	return &STrie[T]{tree: *tree, values: values}
}

func (st *STrie[T]) Marshal(w io.Writer) error {
	if len(st.values) != len(st.tree.Levels) {
		return fmt.Errorf("number of values and levels unmatched: value=%d levels=%d", len(st.values), len(st.tree.Levels))
	}
	if err := st.tree.Write(w); err != nil {
		return err
	}
	if err := gob.NewEncoder(w).Encode(st.values); err != nil {
		return err
	}
	return nil
}

func Unmarshal[T any](r io.Reader) (*STrie[T], error) {
	tree, err := trietree.Read(r)
	if err != nil {
		return nil, err
	}
	if len(tree.Levels) == 0 {
		return &STrie[T]{tree: *tree}, nil
	}
	// read v from r then append it to values.
	values := make([]T, 0, len(tree.Levels))
	if err := gob.NewDecoder(r).Decode(&values); err != nil {
		return nil, err
	}
	return &STrie[T]{tree: *tree, values: values}, nil
}

func (st *STrie[T]) LongestPrefix(s string) (v T, prefix string, ok bool) {
	prefix, id := st.tree.LongestPrefix(s)
	if id == 0 {
		var zero T
		return zero, "", false
	}
	return st.values[id-1], prefix, true
}
