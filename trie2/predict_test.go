package trie2

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type predictIterator[T any] interface {
	PredictIter(string) PredictionIter[T]
}

func testPredictIter[T any](t *testing.T, ptor predictIterator[T], q string, want []Prediction[T]) {
	iter := ptor.PredictIter(q)
	got := make([]Prediction[T], 0, len(want))
	for {
		p := iter()
		if p == nil {
			break
		}
		got = append(got, *p)
	}
	if d := cmp.Diff(want, got); d != "" {
		t.Errorf("unexpected predictions: -want +got\n%s", d)
	}
}

func TestPredict(t *testing.T) {
	dt := &DTrie[Data]{}
	dt.Put("a", Data{111, "aaa"})
	dt.Put("ab", Data{222, "bbb"})
	dt.Put("abc", Data{333, "ccc"})
	dt.Put("d", Data{444, "ddd"})
	dt.Put("de", Data{555, "eee"})
	dt.FillFailure()
	st := dt.Freeze(false)

	for i, c := range []struct {
		q    string
		want []Prediction[Data]
	}{
		{"azd", []Prediction[Data]{
			{Start: 0, End: 1, Key: "a", Value: Data{111, "aaa"}},
			{Start: 2, End: 3, Key: "d", Value: Data{444, "ddd"}},
		}},
	} {
		t.Run(fmt.Sprintf("DTrie i:%d q:%s", i, c.q), func(t *testing.T) {
			testPredictIter(t, dt, c.q, c.want)
		})
		t.Run(fmt.Sprintf("STrie i:%d q:%s", i, c.q), func(t *testing.T) {
			testPredictIter(t, st, c.q, c.want)
		})
	}
}
