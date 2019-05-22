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
		{1, 'b', []node{{1, 2}}},
	})
	testSTreeScan(t, st, "bc", reports{
		{0, 'b', nil},
		{1, 'c', []node{{2, 2}}},
	})
	testSTreeScan(t, st, "bab", reports{
		{0, 'b', nil},
		{1, 'a', nil},
		{2, 'b', []node{{3, 3}, {1, 2}}},
	})
	testSTreeScan(t, st, "d", reports{
		{0, 'd', []node{{4, 1}}},
	})
	testSTreeScan(t, st, "abcde", reports{
		{0, 'a', nil},
		{1, 'b', []node{{1, 2}}},
		{2, 'c', []node{{2, 2}}},
		{3, 'd', []node{{4, 1}}},
		{4, 'e', []node{{5, 5}}},
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
		{1, 'b', []node{{1, 2}}},
	})
	testSTreeScan(t, st, "bc", reports{
		{0, 'b', nil},
		{1, 'c', []node{{2, 2}}},
	})
	testSTreeScan(t, st, "bab", reports{
		{0, 'b', nil},
		{1, 'a', nil},
		{2, 'b', []node{{3, 3}, {1, 2}}},
	})
	testSTreeScan(t, st, "d", reports{
		{0, 'd', []node{{4, 1}}},
	})
	testSTreeScan(t, st, "abcde", reports{
		{0, 'a', nil},
		{1, 'b', []node{{1, 2}}},
		{2, 'c', []node{{2, 2}}},
		{3, 'd', []node{{4, 1}}},
		{4, 'e', []node{{5, 5}}},
	})
}
