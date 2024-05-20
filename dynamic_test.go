package trietree_test

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/koron-go/trietree"
)

type node struct {
	id int
	lv int
}

type report struct {
	i     int
	c     rune
	nodes []node
}

type reports []*report

func (r *reports) ScanReport(ev trietree.ScanEvent) {
	var nodes []node
	if len(ev.Nodes) > 0 {
		nodes = make([]node, 0, len(ev.Nodes))
		for _, n := range ev.Nodes {
			nodes = append(nodes, node{id: n.ID, lv: n.Level})
		}
	}
	*r = append(*r, &report{
		i:     ev.Index,
		c:     ev.Label,
		nodes: nodes,
	})
}

func (r reports) compare(t *testing.T, exp reports) {
	t.Helper()
	if len(r) != len(exp) {
		t.Fatalf("reports.length mismatch: actual=%d expected=%d",
			len(r), len(exp))
	}
	for i := range r {
		a, e := r[i], exp[i]
		if a.i != e.i || a.c != e.c || !reflect.DeepEqual(a.nodes, e.nodes) {
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
	err := dt.Scan(s, &act)
	if err != nil {
		t.Fatalf("scan is failed: %v", err)
	}
	act.compare(t, exp)
}

func TestDTree_simple_single(t *testing.T) {
	dt := testDTreePut(t, &trietree.DTree{}, "1", "2", "3", "4", "5")
	testDTreeScan(t, dt, "1", reports{{0, '1', []node{{1, 1}}}})
	testDTreeScan(t, dt, "2", reports{{0, '2', []node{{2, 1}}}})
	testDTreeScan(t, dt, "3", reports{{0, '3', []node{{3, 1}}}})
	testDTreeScan(t, dt, "4", reports{{0, '4', []node{{4, 1}}}})
	testDTreeScan(t, dt, "5", reports{{0, '5', []node{{5, 1}}}})
	testDTreeScan(t, dt, "6", reports{{0, '6', nil}})
}

func TestDTree_simple_multiple(t *testing.T) {
	dt := testDTreePut(t, &trietree.DTree{}, "1", "2", "3", "4", "5")
	testDTreeScan(t, dt, "1234567890", reports{
		{0, '1', []node{{1, 1}}},
		{1, '2', []node{{2, 1}}},
		{2, '3', []node{{3, 1}}},
		{3, '4', []node{{4, 1}}},
		{4, '5', []node{{5, 1}}},
		{5, '6', nil},
		{6, '7', nil},
		{7, '8', nil},
		{8, '9', nil},
		{9, '0', nil},
	})
}

func TestDTree_basic(t *testing.T) {
	dt := testDTreePut(t, &trietree.DTree{}, "ab", "bc", "bab", "d", "abcde")
	testDTreeScan(t, dt, "ab", reports{
		{0, 'a', nil},
		{1, 'b', []node{{1, 2}}},
	})
	testDTreeScan(t, dt, "bc", reports{
		{0, 'b', nil},
		{1, 'c', []node{{2, 2}}},
	})
	testDTreeScan(t, dt, "bab", reports{
		{0, 'b', nil},
		{1, 'a', nil},
		{2, 'b', []node{{3, 3}, {1, 2}}},
	})
	testDTreeScan(t, dt, "d", reports{
		{0, 'd', []node{{4, 1}}},
	})
	testDTreeScan(t, dt, "abcde", reports{
		{0, 'a', nil},
		{1, 'b', []node{{1, 2}}},
		{2, 'c', []node{{2, 2}}},
		{3, 'd', []node{{4, 1}}},
		{4, 'e', []node{{5, 5}}},
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

func TestDTree_get(t *testing.T) {
	dt := testDTreePut(t, &trietree.DTree{}, "bc", "bab", "d", "abcde", "ab")
	n := dt.Get("bab")
	if n == nil {
		t.Error("not found nodes for \"bab\"")
	}
	n1 := dt.Get("cab")
	if n1 != nil {
		t.Errorf("unexpected node found: %+v", n1)
	}

}

func TestDTree_MatchLongest(t *testing.T) {
	dt := testDTreePut(t, &trietree.DTree{},
		"ab", "abcde",
		"bab", "bc",
		"d",
	)
	for i, c := range []struct{ query, want string }{
		{"a", ""},
		{"ab", "ab"},
		{"abcdefg", "abcde"},
		{"b", ""},
		{"bc", "bc"},
		{"bcdzzz", "bc"},
		{"babbab", "bab"},
		{"bac", ""},
		{"bbc", ""},
		{"zzz", ""},
	} {
		got, _ := dt.LongestPrefix(c.query)
		if d := cmp.Diff(c.want, got); d != "" {
			t.Errorf("unexpected #%d %+v: -want +got\n%s", i, c, d)
		}
	}
}

func TestDTree_ScanMultiple(t *testing.T) {
	dt := testDTreePut(t, &trietree.DTree{}, "a", "ab", "abc", "d", "de")
	testDTreeScan(t, dt, "azd", reports{
		{0, 'a', []node{{1, 1}}},
		{1, 'z', nil},
		{2, 'd', []node{{4, 1}}},
	})
}
