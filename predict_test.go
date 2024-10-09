package trietree_test

import (
	"iter"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/koron-go/trietree"
)

type prediction struct {
	Start int
	End   int
	ID    int
	Key   string
}

type predictIterator interface {
	PredictIter(string) trietree.PredictionIter
}

func testPredictIter(t *testing.T, ptor predictIterator, q string, want []prediction) {
	t.Helper()
	got := make([]prediction, 0, 10)
	iter := ptor.PredictIter(q)
	for {
		p := iter()
		if p == nil {
			break
		}
		got = append(got, prediction{
			Start: p.Start,
			End:   p.End,
			ID:    p.ID,
			Key:   q[p.Start:p.End],
		})
	}
	if d := cmp.Diff(want, got); d != "" {
		t.Errorf("unexpected predictions: -want +got\n%s", d)
	}
}

type predictIteratorBuilder func(t *testing.T, keys ...string) predictIterator

func testPredictIterSingle(t *testing.T, build predictIteratorBuilder) {
	ptor := build(t, "1", "2", "3", "4", "5")
	testPredictIter(t, ptor, "1", []prediction{
		{Start: 0, End: 1, ID: 1, Key: "1"},
	})
	testPredictIter(t, ptor, "2", []prediction{
		{Start: 0, End: 1, ID: 2, Key: "2"},
	})
	testPredictIter(t, ptor, "3", []prediction{
		{Start: 0, End: 1, ID: 3, Key: "3"},
	})
	testPredictIter(t, ptor, "4", []prediction{
		{Start: 0, End: 1, ID: 4, Key: "4"},
	})
	testPredictIter(t, ptor, "5", []prediction{
		{Start: 0, End: 1, ID: 5, Key: "5"},
	})
	testPredictIter(t, ptor, "6", []prediction{})
}

func testPredictIterMultiple(t *testing.T, build predictIteratorBuilder) {
	ptor := build(t, "1", "2", "3", "4", "5")
	testPredictIter(t, ptor, "1234567890", []prediction{
		{Start: 0, End: 1, ID: 1, Key: "1"},
		{Start: 1, End: 2, ID: 2, Key: "2"},
		{Start: 2, End: 3, ID: 3, Key: "3"},
		{Start: 3, End: 4, ID: 4, Key: "4"},
		{Start: 4, End: 5, ID: 5, Key: "5"},
	})
}

func testPredictIterBasic(t *testing.T, build predictIteratorBuilder) {
	ptor := build(t, "ab", "bc", "bab", "d", "abcde")
	testPredictIter(t, ptor, "ab", []prediction{
		{Start: 0, End: 2, ID: 1, Key: "ab"},
	})
	testPredictIter(t, ptor, "bc", []prediction{
		{Start: 0, End: 2, ID: 2, Key: "bc"},
	})
	testPredictIter(t, ptor, "bab", []prediction{
		{Start: 0, End: 3, ID: 3, Key: "bab"},
		{Start: 1, End: 3, ID: 1, Key: "ab"},
	})
	testPredictIter(t, ptor, "d", []prediction{
		{Start: 0, End: 1, ID: 4, Key: "d"},
	})
	testPredictIter(t, ptor, "abcde", []prediction{
		{Start: 0, End: 2, ID: 1, Key: "ab"},
		{Start: 1, End: 3, ID: 2, Key: "bc"},
		{Start: 3, End: 4, ID: 4, Key: "d"},
		{Start: 0, End: 5, ID: 5, Key: "abcde"},
	})
}

func testPredictIterAll(t *testing.T, builder predictIteratorBuilder) {
	t.Run("single", func(t *testing.T) {
		testPredictIterSingle(t, builder)
	})
	t.Run("multiple", func(t *testing.T) {
		testPredictIterMultiple(t, builder)
	})
	t.Run("basic", func(t *testing.T) {
		testPredictIterBasic(t, builder)
	})
}

func TestPredictIter(t *testing.T) {
	t.Run("dynamic", func(t *testing.T) {
		testPredictIterAll(t, func(t *testing.T, keys ...string) predictIterator {
			return testDTreePut(t, &trietree.DTree{}, keys...)
		})
	})
	t.Run("static", func(t *testing.T) {
		testPredictIterAll(t, func(t *testing.T, keys ...string) predictIterator {
			dt := testDTreePut(t, &trietree.DTree{}, keys...)
			return trietree.Freeze(dt)
		})
	})
}

func TestDTree_PredictMultiple(t *testing.T) {
	dt := testDTreePut(t, &trietree.DTree{}, "a", "ab", "abc", "d", "de")
	testPredictIter(t, dt, "azd", []prediction{
		{Start: 0, End: 1, ID: 1, Key: "a"},
		{Start: 2, End: 3, ID: 4, Key: "d"},
	})
}

func TestSTree_PredictMultiple(t *testing.T) {
	dt := testDTreePut(t, &trietree.DTree{}, "a", "ab", "abc", "d", "de")
	st := trietree.Freeze(dt)
	testPredictIter(t, st, "azd", []prediction{
		{Start: 0, End: 1, ID: 1, Key: "a"},
		{Start: 2, End: 3, ID: 4, Key: "d"},
	})
}

type predictor interface {
	Predict(string) iter.Seq[trietree.Prediction]
}

func testPredict(t *testing.T, ptor predictor, q string, want []prediction) {
	t.Helper()
	got := make([]prediction, 0, 10)
	for p := range ptor.Predict(q) {
		got = append(got, prediction{
			Start: p.Start,
			End:   p.End,
			ID:    p.ID,
			Key:   q[p.Start:p.End],
		})
	}
	if d := cmp.Diff(want, got); d != "" {
		t.Errorf("unexpected predictions: -want +got\n%s", d)
	}
}

type predictorBuilder func(t *testing.T, keys ...string) predictor

func testPredictSingle(t *testing.T, build predictorBuilder) {
	ptor := build(t, "1", "2", "3", "4", "5")
	testPredict(t, ptor, "1", []prediction{
		{Start: 0, End: 1, ID: 1, Key: "1"},
	})
	testPredict(t, ptor, "2", []prediction{
		{Start: 0, End: 1, ID: 2, Key: "2"},
	})
	testPredict(t, ptor, "3", []prediction{
		{Start: 0, End: 1, ID: 3, Key: "3"},
	})
	testPredict(t, ptor, "4", []prediction{
		{Start: 0, End: 1, ID: 4, Key: "4"},
	})
	testPredict(t, ptor, "5", []prediction{
		{Start: 0, End: 1, ID: 5, Key: "5"},
	})
	testPredict(t, ptor, "6", []prediction{})
}

func testPredictMultiple(t *testing.T, build predictorBuilder) {
	ptor := build(t, "1", "2", "3", "4", "5")
	testPredict(t, ptor, "1234567890", []prediction{
		{Start: 0, End: 1, ID: 1, Key: "1"},
		{Start: 1, End: 2, ID: 2, Key: "2"},
		{Start: 2, End: 3, ID: 3, Key: "3"},
		{Start: 3, End: 4, ID: 4, Key: "4"},
		{Start: 4, End: 5, ID: 5, Key: "5"},
	})
}

func testPredictBasic(t *testing.T, build predictorBuilder) {
	ptor := build(t, "ab", "bc", "bab", "d", "abcde")
	testPredict(t, ptor, "ab", []prediction{
		{Start: 0, End: 2, ID: 1, Key: "ab"},
	})
	testPredict(t, ptor, "bc", []prediction{
		{Start: 0, End: 2, ID: 2, Key: "bc"},
	})
	testPredict(t, ptor, "bab", []prediction{
		{Start: 0, End: 3, ID: 3, Key: "bab"},
		{Start: 1, End: 3, ID: 1, Key: "ab"},
	})
	testPredict(t, ptor, "d", []prediction{
		{Start: 0, End: 1, ID: 4, Key: "d"},
	})
	testPredict(t, ptor, "abcde", []prediction{
		{Start: 0, End: 2, ID: 1, Key: "ab"},
		{Start: 1, End: 3, ID: 2, Key: "bc"},
		{Start: 3, End: 4, ID: 4, Key: "d"},
		{Start: 0, End: 5, ID: 5, Key: "abcde"},
	})
}

func testPredictAll(t *testing.T, builder predictorBuilder) {
	t.Run("single", func(t *testing.T) {
		testPredictSingle(t, builder)
	})
	t.Run("multiple", func(t *testing.T) {
		testPredictMultiple(t, builder)
	})
	t.Run("basic", func(t *testing.T) {
		testPredictBasic(t, builder)
	})
}

func TestPredictSeq(t *testing.T) {
	t.Run("dynamic", func(t *testing.T) {
		testPredictAll(t, func(t *testing.T, keys ...string) predictor {
			return testDTreePut(t, &trietree.DTree{}, keys...)
		})
	})
	t.Run("static", func(t *testing.T) {
		testPredictAll(t, func(t *testing.T, keys ...string) predictor {
			dt := testDTreePut(t, &trietree.DTree{}, keys...)
			return trietree.Freeze(dt)
		})
	})
}
