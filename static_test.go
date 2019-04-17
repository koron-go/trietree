package trietree_test

import (
	"bytes"
	"testing"

	"github.com/koron-go/trietree"
)

func testSTreeScan(t *testing.T, st *trietree.STree, s string, exp reports) {
	t.Helper()
	var act reports
	err := st.Scan(s, (&act))
	if err != nil {
		t.Fatalf("scan is failed: %v", err)
	}
	act.compare(t, exp)
}

func TestSTree_freeze(t *testing.T) {
	dt := testDTreePut(t, &trietree.DTree{}, "ab", "bc", "bab", "d", "abcde")
	st := trietree.Freeze(dt)
	testSTreeScan(t, st, "ab", reports{
		{0, 'a', nil},
		{1, 'b', []int{1}},
	})
	testSTreeScan(t, st, "bc", reports{
		{0, 'b', nil},
		{1, 'c', []int{2}},
	})
	testSTreeScan(t, st, "bab", reports{
		{0, 'b', nil},
		{1, 'a', nil},
		{2, 'b', []int{3, 1}},
	})
	testSTreeScan(t, st, "d", reports{
		{0, 'd', []int{4}},
	})
	testSTreeScan(t, st, "abcde", reports{
		{0, 'a', nil},
		{1, 'b', []int{1}},
		{2, 'c', []int{2}},
		{3, 'd', []int{4}},
		{4, 'e', []int{5}},
	})
}

func TestSTree_serialize(t *testing.T) {
	dt := testDTreePut(t, &trietree.DTree{}, "ab", "bc", "bab", "d", "abcde")
	st0 := trietree.Freeze(dt)

	b := &bytes.Buffer{}
	err := st0.Write(b)
	if err != nil {
		t.Fatalf("write failed: %s", err)
	}
	st, err := trietree.Read(b)
	if err != nil {
		t.Fatalf("read failed: %s", err)
	}

	testSTreeScan(t, st, "ab", reports{
		{0, 'a', nil},
		{1, 'b', []int{1}},
	})
	testSTreeScan(t, st, "bc", reports{
		{0, 'b', nil},
		{1, 'c', []int{2}},
	})
	testSTreeScan(t, st, "bab", reports{
		{0, 'b', nil},
		{1, 'a', nil},
		{2, 'b', []int{3, 1}},
	})
	testSTreeScan(t, st, "d", reports{
		{0, 'd', []int{4}},
	})
	testSTreeScan(t, st, "abcde", reports{
		{0, 'a', nil},
		{1, 'b', []int{1}},
		{2, 'c', []int{2}},
		{3, 'd', []int{4}},
		{4, 'e', []int{5}},
	})
}
