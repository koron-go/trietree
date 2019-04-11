package trietree_test

import (
	"reflect"
	"testing"

	"github.com/koron-go/trietree"
)

type report struct {
	i    int
	c    rune
	root bool
	ids  []int
}

type reports []*report

func (r *reports) Report(i int, c rune, root bool, ids []int) {
	if ids != nil {
		tmp := make([]int, len(ids))
		copy(tmp, ids)
		ids = tmp
	}
	*r = append(*r, &report{i: i, c: c, root: root, ids: ids})
}

func (r reports) compare(t *testing.T, exp reports) {
	t.Helper()
	if len(r) != len(exp) {
		t.Fatalf("reports.length mismatch: actual=%d expected=%d",
			len(r), len(exp))
	}
	for i := range r {
		a, e := r[i], exp[i]
		if a.i != e.i || a.c != e.c || a.root != e.root || !reflect.DeepEqual(a.ids, e.ids) {
			t.Fatalf("report#%d isn't match:\n    actual=%+v\n  expected=%+v", i, a, e)
		}
	}
}

func testDTreePut(t *testing.T, dt *trietree.DTree, keys ...string) *trietree.DTree {
	t.Helper()
	for i, k := range keys {
		exp := i + 1
		act := dt.Put(k)
		if act != exp {
			t.Fatalf("put returns unexpected: actual=%d expected=%d", act, exp)
		}
	}
	dt.FillFailure()
	return dt
}

func testDTreeScan(t *testing.T, dt *trietree.DTree, s string, exp reports) {
	t.Helper()
	var act reports
	err := dt.Scan(s, (&act))
	if err != nil {
		t.Fatalf("scan is failed: %v", err)
	}
	act.compare(t, exp)
}

func TestDTree_simple_single(t *testing.T) {
	dt := testDTreePut(t, &trietree.DTree{}, "1", "2", "3", "4", "5")
	testDTreeScan(t, dt, "1", reports{{0, '1', true, []int{1}}})
	testDTreeScan(t, dt, "2", reports{{0, '2', true, []int{2}}})
	testDTreeScan(t, dt, "3", reports{{0, '3', true, []int{3}}})
	testDTreeScan(t, dt, "4", reports{{0, '4', true, []int{4}}})
	testDTreeScan(t, dt, "5", reports{{0, '5', true, []int{5}}})
	testDTreeScan(t, dt, "6", reports{{0, '6', true, nil}})
}

func TestDTree_simple_multiple(t *testing.T) {
	dt := testDTreePut(t, &trietree.DTree{}, "1", "2", "3", "4", "5")
	testDTreeScan(t, dt, "1234567890", reports{
		{0, '1', true, []int{1}},
		{1, '2', true, []int{2}},
		{2, '3', true, []int{3}},
		{3, '4', true, []int{4}},
		{4, '5', true, []int{5}},
		{5, '6', true, nil},
		{6, '7', true, nil},
		{7, '8', true, nil},
		{8, '9', true, nil},
		{9, '0', true, nil},
	})
}

func TestDTree_basic(t *testing.T) {
	dt := testDTreePut(t, &trietree.DTree{}, "ab", "bc", "bab", "d", "abcde")
	testDTreeScan(t, dt, "ab", reports{
		{0, 'a', true, nil},
		{1, 'b', false, []int{1}},
	})
	testDTreeScan(t, dt, "bc", reports{
		{0, 'b', true, nil},
		{1, 'c', false, []int{2}},
	})
	testDTreeScan(t, dt, "bab", reports{
		{0, 'b', true, nil},
		{1, 'a', false, nil},
		{2, 'b', false, []int{3, 1}},
	})
	testDTreeScan(t, dt, "d", reports{
		{0, 'd', true, []int{4}},
	})
	testDTreeScan(t, dt, "abcde", reports{
		{0, 'a', true, nil},
		{1, 'b', false, []int{1}},
		{2, 'c', false, []int{2}},
		{3, 'd', false, []int{4}},
		{4, 'e', false, []int{5}},
	})
}

func TestDTree_count(t *testing.T) {
	dt := testDTreePut(t, &trietree.DTree{}, "ab", "bc", "bab", "d", "abcde")
	if n := dt.Root.CountChild(); n != 3 {
		t.Fatalf("CountChild()=%d unexpected (expected:3)", n)
	}
	if n := dt.Root.CountAll(); n != 11 {
		t.Fatalf("CountAll()=%d unexpected (expected:11)", n)
	}
}
