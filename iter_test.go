//go:build goexperiment.rangefunc

package trietree_test

import (
	"iter"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/koron-go/trietree"
)

type predictor interface {
	Predict(string) iter.Seq[trietree.Prediction]
}

func testPredict(t *testing.T, ptor predictor, q string, want []trietree.Prediction) {
	t.Helper()
	got := make([]trietree.Prediction, 0, 10)
	for p := range ptor.Predict(q) {
		got = append(got, p)
	}
	if d := cmp.Diff(want, got); d != "" {
		t.Errorf("unexpected predictions: -want +got\n%s", d)
	}
}

type predictorBuilder func(t *testing.T, keys ...string) predictor

func testPredictSingle(t *testing.T, build predictorBuilder) {
	ptor := build(t, "1", "2", "3", "4", "5")
	testPredict(t, ptor, "1", []trietree.Prediction{
		{Index: 0, ID: 1, Depth: 1, Label: '1'},
	})
	testPredict(t, ptor, "2", []trietree.Prediction{
		{Index: 0, ID: 2, Depth: 1, Label: '2'},
	})
	testPredict(t, ptor, "3", []trietree.Prediction{
		{Index: 0, ID: 3, Depth: 1, Label: '3'},
	})
	testPredict(t, ptor, "4", []trietree.Prediction{
		{Index: 0, ID: 4, Depth: 1, Label: '4'},
	})
	testPredict(t, ptor, "5", []trietree.Prediction{
		{Index: 0, ID: 5, Depth: 1, Label: '5'},
	})
	testPredict(t, ptor, "6", []trietree.Prediction{})
}

func testPredictMultiple(t *testing.T, build predictorBuilder) {
	ptor := build(t, "1", "2", "3", "4", "5")
	testPredict(t, ptor, "1234567890", []trietree.Prediction{
		{Index: 0, ID: 1, Depth: 1, Label: '1'},
		{Index: 1, ID: 2, Depth: 1, Label: '2'},
		{Index: 2, ID: 3, Depth: 1, Label: '3'},
		{Index: 3, ID: 4, Depth: 1, Label: '4'},
		{Index: 4, ID: 5, Depth: 1, Label: '5'},
	})
}

func testPredictBasic(t *testing.T, build predictorBuilder) {
	ptor := build(t, "ab", "bc", "bab", "d", "abcde")
	testPredict(t, ptor, "ab", []trietree.Prediction{
		{Index: 1, ID: 1, Depth: 2, Label: 'b'},
	})
	testPredict(t, ptor, "bc", []trietree.Prediction{
		{Index: 1, ID: 2, Depth: 2, Label: 'c'},
	})
	testPredict(t, ptor, "bab", []trietree.Prediction{
		{Index: 2, ID: 3, Depth: 3, Label: 'b'},
		{Index: 2, ID: 1, Depth: 2, Label: 'b'},
	})
	testPredict(t, ptor, "d", []trietree.Prediction{
		{Index: 0, ID: 4, Depth: 1, Label: 'd'},
	})
	testPredict(t, ptor, "abcde", []trietree.Prediction{
		{Index: 1, ID: 1, Depth: 2, Label: 'b'},
		{Index: 2, ID: 2, Depth: 2, Label: 'c'},
		{Index: 3, ID: 4, Depth: 1, Label: 'd'},
		{Index: 4, ID: 5, Depth: 5, Label: 'e'},
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

func TestPredict(t *testing.T) {
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
