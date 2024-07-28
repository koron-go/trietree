package trie2

import (
	"encoding/gob"
	"fmt"
	"io"

	"github.com/koron-go/trietree"
)

// DTrie is dynamic tree.  This can be added a node (pair of key and value).
type DTrie[T any] struct {
	tree   trietree.DTree
	values []T
}

// FillFailure fill Failure field with Aho-Corasick algorithm.
func (dt *DTrie[T]) FillFailure() {
	dt.tree.FillFailure()
}

// Put adds a pair of key and value.
func (dt *DTrie[T]) Put(k string, v T) {
	id := dt.tree.Put(k)
	if id-1 == len(dt.values) {
		dt.values = append(dt.values, v)
		return
	}
	// update an existed value.
	dt.values[id-1] = v
}

// LongestPrefix performs "logest prefix match" with s.  It will return a
// corresponding value and prefix when s found in the trie-tree.
func (dt *DTrie[T]) LongestPrefix(s string) (v T, prefix string, ok bool) {
	prefix, id := dt.tree.LongestPrefix(s)
	if id == 0 {
		var zero T
		return zero, "", false
	}
	return dt.values[id-1], prefix, true
}

// STrie is static tree, which provides compact form of trie-tree.
type STrie[T any] struct {
	tree   trietree.STree
	values []T
}

// Freeze creates a STrie from DTrie.
// The generated STrie is equivalent to the original DTrie. STrie cannot add
// any pairs, but it can be marshaled (serialized) and unmarshaled
// (deserialized). If the argument copyValues is false, the array of values
// held by DTrie will be used as is. If true, create and use a copy.
func (dt *DTrie[T]) Freeze(copyValues bool) *STrie[T] {
	tree := trietree.Freeze(&dt.tree)
	var values []T
	if copyValues {
		values = make([]T, len(dt.values))
		copy(values, dt.values)
	} else {
		values = dt.values
	}
	return &STrie[T]{tree: *tree, values: values}
}

// Marshal serializes STrie on w.
// You can marshal values using the marshalValues function.
// encoding/gob is used to marshal values when marshalValues is nil.
func (st *STrie[T]) Marshal(w io.Writer, marshalValues func(io.Writer, []T) error) error {
	if len(st.values) != len(st.tree.Levels) {
		return fmt.Errorf("number of values and levels unmatched: value=%d levels=%d", len(st.values), len(st.tree.Levels))
	}
	if err := st.tree.Write(w); err != nil {
		return err
	}
	if marshalValues != nil {
		if err := marshalValues(w, st.values); err != nil {
			return fmt.Errorf("failed to marshal values: %w", err)
		}
		return nil
	}
	if err := gob.NewEncoder(w).Encode(st.values); err != nil {
		return err
	}
	return nil
}

// Unmarshal deserializes a STrie from r.
// You can unmarshal values using the unmarshalValues function.
// encoding/gob is used to unmarshal values when unmarshalValues is nil.
func Unmarshal[T any](r io.Reader, unmarshalValues func(io.Reader, int) ([]T, error)) (*STrie[T], error) {
	tree, err := trietree.Read(r)
	if err != nil {
		return nil, err
	}
	if len(tree.Levels) == 0 {
		return &STrie[T]{tree: *tree}, nil
	}
	// read values from r with unmarshalValues.
	if unmarshalValues != nil {
		values, err := unmarshalValues(r, len(tree.Levels))
		if err != nil {
			return nil, err
		}
		return &STrie[T]{tree: *tree, values: values}, nil
	}
	// read values from r without unmarshalValues.
	values := make([]T, 0, len(tree.Levels))
	if err := gob.NewDecoder(r).Decode(&values); err != nil {
		return nil, err
	}
	return &STrie[T]{tree: *tree, values: values}, nil
}

// LongestPrefix performs "logest prefix match" with s.  It will return a
// corresponding value and prefix when s found in the trie-tree.
func (st *STrie[T]) LongestPrefix(s string) (v T, prefix string, ok bool) {
	prefix, id := st.tree.LongestPrefix(s)
	if id == 0 {
		var zero T
		return zero, "", false
	}
	return st.values[id-1], prefix, true
}
