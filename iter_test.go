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
		{Index: 0, Label: '1', ID: 1, Depth: 1},
	})
	testPredict(t, ptor, "2", []trietree.Prediction{
		{Index: 0, Label: '2', ID: 2, Depth: 1},
	})
	testPredict(t, ptor, "3", []trietree.Prediction{
		{Index: 0, Label: '3', ID: 3, Depth: 1},
	})
	testPredict(t, ptor, "4", []trietree.Prediction{
		{Index: 0, Label: '4', ID: 4, Depth: 1},
	})
	testPredict(t, ptor, "5", []trietree.Prediction{
		{Index: 0, Label: '5', ID: 5, Depth: 1},
	})
	testPredict(t, ptor, "6", []trietree.Prediction{})
}

func testPredictMultiple(t *testing.T, build predictorBuilder) {
	ptor := build(t, "1", "2", "3", "4", "5")
	testPredict(t, ptor, "1234567890", []trietree.Prediction{
		{Index: 0, Label: '1', ID: 1, Depth: 1},
		{Index: 1, Label: '2', ID: 2, Depth: 1},
		{Index: 2, Label: '3', ID: 3, Depth: 1},
		{Index: 3, Label: '4', ID: 4, Depth: 1},
		{Index: 4, Label: '5', ID: 5, Depth: 1},
	})
}

func testPredictBasic(t *testing.T, build predictorBuilder) {
	ptor := build(t, "ab", "bc", "bab", "d", "abcde")
	testPredict(t, ptor, "ab", []trietree.Prediction{
		{Index: 1, Label: 'b', ID: 1, Depth: 2},
	})
	testPredict(t, ptor, "bc", []trietree.Prediction{
		{Index: 1, Label: 'c', ID: 2, Depth: 2},
	})
	testPredict(t, ptor, "bab", []trietree.Prediction{
		{Index: 2, Label: 'b', ID: 3, Depth: 3},
		{Index: 2, Label: 'b', ID: 1, Depth: 2},
	})
	testPredict(t, ptor, "d", []trietree.Prediction{
		{Index: 0, Label: 'd', ID: 4, Depth: 1},
	})
	testPredict(t, ptor, "abcde", []trietree.Prediction{
		{Index: 1, Label: 'b', ID: 1, Depth: 2},
		{Index: 2, Label: 'c', ID: 2, Depth: 2},
		{Index: 3, Label: 'd', ID: 4, Depth: 1},
		{Index: 4, Label: 'e', ID: 5, Depth: 5},
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

func buildDynamicPredictor(t *testing.T, keys ...string) predictor {
	return testDTreePut(t, &trietree.DTree{}, keys...)
}

func buildStaticPredictor(t *testing.T, keys ...string) predictor {
	dt := testDTreePut(t, &trietree.DTree{}, keys...)
	return trietree.Freeze(dt)
}

func TestDynanicPredict(t *testing.T) {
	testPredictAll(t, buildDynamicPredictor)
}

func TestStaticPredict(t *testing.T) {
	testPredictAll(t, buildStaticPredictor)
}
