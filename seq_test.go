package trietree_test

import (
	"iter"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/koron-go/trietree"
)

type predictor interface {
	PredictSeq(string) iter.Seq[trietree.Prediction]
}

func testPredict(t *testing.T, ptor predictor, q string, want []prediction) {
	t.Helper()
	got := make([]prediction, 0, 10)
	for p := range ptor.PredictSeq(q) {
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
