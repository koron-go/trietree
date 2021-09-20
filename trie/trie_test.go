package trie

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func testPutAll(t *testing.T, tr *Trie, keys ...string) {
	t.Helper()
	for _, k := range keys {
		_, err := tr.Put(k)
		if err != nil {
			t.Fatalf("failed to put %q: %v", k, err)
		}
	}
}

func testScan(t *testing.T, tr *Trie, s string, expected []ReportEvent) {
	t.Helper()
	var actual []ReportEvent
	err := tr.Scan(context.Background(), s, ReporterFunc(func(ctx context.Context, ev ReportEvent) {
		actual = append(actual, ev)
	}))
	if err != nil {
		t.Errorf("scan failed: %v", err)
	}
	if d := cmp.Diff(expected, actual); d != "" {
		t.Errorf("unexpected scan reports: -want +got\n%s", d)
	}
}

func TestTrie_Basic(t *testing.T) {
	tr := New()
	testPutAll(t, tr, "ab", "bc", "bab", "d", "abcde")
	testScan(t, tr, "ab", []ReportEvent{
		{Index: 0, Label: 'a'},
		{Index: 1, Label: 'b', Nodes: []ReportNode{
			{ID: 1, Level: 2},
		}},
	})
	testScan(t, tr, "bc", []ReportEvent{
		{Index: 0, Label: 'b'},
		{Index: 1, Label: 'c', Nodes: []ReportNode{
			{ID: 2, Level: 2},
		}},
	})
	testScan(t, tr, "bab", []ReportEvent{
		{Index: 0, Label: 'b'},
		{Index: 1, Label: 'a'},
		{Index: 2, Label: 'b', Nodes: []ReportNode{
			{ID: 3, Level: 3},
			//{ID: 1, Level: 2},
		}},
	})
	testScan(t, tr, "d", []ReportEvent{
		{Index: 0, Label: 'd', Nodes: []ReportNode{
			{ID: 4, Level: 1},
		}},
	})
	testScan(t, tr, "abcde", []ReportEvent{
		{Index: 0, Label: 'a'},
		{Index: 1, Label: 'b', Nodes: []ReportNode{{ID: 1, Level: 2}}},
		{Index: 2, Label: 'c'},
		{Index: 3, Label: 'd'},
		{Index: 4, Label: 'e', Nodes: []ReportNode{{ID: 5, Level: 5}}},
	})
}

func TestTrie_Freeze(t *testing.T) {
	tr := New()
	testPutAll(t, tr, "ab", "bc", "bab", "d", "abcde")
	err := tr.Freeze()
	if err != nil {
		t.Fatalf("failed to freeze: %v", err)
	}
	testScan(t, tr, "ab", []ReportEvent{
		{Index: 0, Label: 'a'},
		{Index: 1, Label: 'b', Nodes: []ReportNode{
			{ID: 1, Level: 2},
		}},
	})
	testScan(t, tr, "bc", []ReportEvent{
		{Index: 0, Label: 'b'},
		{Index: 1, Label: 'c', Nodes: []ReportNode{
			{ID: 2, Level: 2},
		}},
	})
	testScan(t, tr, "bab", []ReportEvent{
		{Index: 0, Label: 'b'},
		{Index: 1, Label: 'a'},
		{Index: 2, Label: 'b', Nodes: []ReportNode{
			{ID: 3, Level: 3},
			{ID: 1, Level: 2},
		}},
	})
	testScan(t, tr, "d", []ReportEvent{
		{Index: 0, Label: 'd', Nodes: []ReportNode{
			{ID: 4, Level: 1},
		}},
	})
	testScan(t, tr, "abcde", []ReportEvent{
		{Index: 0, Label: 'a'},
		{Index: 1, Label: 'b', Nodes: []ReportNode{{ID: 1, Level: 2}}},
		{Index: 2, Label: 'c', Nodes: []ReportNode{{ID: 2, Level: 2}}},
		{Index: 3, Label: 'd', Nodes: []ReportNode{{ID: 4, Level: 1}}},
		{Index: 4, Label: 'e', Nodes: []ReportNode{{ID: 5, Level: 5}}},
	})
}

func TestTrie_ErrFreezedAlready(t *testing.T) {
	tr := New()
	testPutAll(t, tr, "ab", "bc", "bab", "d", "abcde")
	err := tr.Freeze()
	if err != nil {
		t.Fatalf("failed to freeze: %v", err)
	}

	n, err := tr.Put("foo")
	if n != 0 {
		t.Errorf("putting on freezed trie should be return 0: got=%d", n)
	}
	if !errors.Is(err, ErrFreezedAlready) {
		t.Errorf("unexpected failure to put: want=%v got=%v", ErrFreezedAlready, err)
	}

	err = tr.Freeze()
	if !errors.Is(err, ErrFreezedAlready) {
		t.Errorf("unexpected failure to freeze: want=%v got=%v", ErrFreezedAlready, err)
	}
}

func TestTrie_Unmarshal(t *testing.T) {
	tr0 := New()
	testPutAll(t, tr0, "ab", "bc", "bab", "d", "abcde")

	var bb = &bytes.Buffer{}
	err := tr0.Marshal(bb)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	tr, err := Unmarshal(bb)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	testScan(t, tr, "ab", []ReportEvent{
		{Index: 0, Label: 'a'},
		{Index: 1, Label: 'b', Nodes: []ReportNode{
			{ID: 1, Level: 2},
		}},
	})
	testScan(t, tr, "bc", []ReportEvent{
		{Index: 0, Label: 'b'},
		{Index: 1, Label: 'c', Nodes: []ReportNode{
			{ID: 2, Level: 2},
		}},
	})
	testScan(t, tr, "bab", []ReportEvent{
		{Index: 0, Label: 'b'},
		{Index: 1, Label: 'a'},
		{Index: 2, Label: 'b', Nodes: []ReportNode{
			{ID: 3, Level: 3},
			{ID: 1, Level: 2},
		}},
	})
	testScan(t, tr, "d", []ReportEvent{
		{Index: 0, Label: 'd', Nodes: []ReportNode{
			{ID: 4, Level: 1},
		}},
	})
	testScan(t, tr, "abcde", []ReportEvent{
		{Index: 0, Label: 'a'},
		{Index: 1, Label: 'b', Nodes: []ReportNode{{ID: 1, Level: 2}}},
		{Index: 2, Label: 'c', Nodes: []ReportNode{{ID: 2, Level: 2}}},
		{Index: 3, Label: 'd', Nodes: []ReportNode{{ID: 4, Level: 1}}},
		{Index: 4, Label: 'e', Nodes: []ReportNode{{ID: 5, Level: 5}}},
	})
}

func TestTrie_UnmarshalFreezed(t *testing.T) {
	tr0 := New()
	testPutAll(t, tr0, "ab", "bc", "bab", "d", "abcde")
	err := tr0.Freeze()
	if err != nil {
		t.Errorf("failed to freeze: %v", err)
	}

	var bb = &bytes.Buffer{}
	err = tr0.Marshal(bb)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	tr, err := Unmarshal(bb)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	testScan(t, tr, "ab", []ReportEvent{
		{Index: 0, Label: 'a'},
		{Index: 1, Label: 'b', Nodes: []ReportNode{
			{ID: 1, Level: 2},
		}},
	})
	testScan(t, tr, "bc", []ReportEvent{
		{Index: 0, Label: 'b'},
		{Index: 1, Label: 'c', Nodes: []ReportNode{
			{ID: 2, Level: 2},
		}},
	})
	testScan(t, tr, "bab", []ReportEvent{
		{Index: 0, Label: 'b'},
		{Index: 1, Label: 'a'},
		{Index: 2, Label: 'b', Nodes: []ReportNode{
			{ID: 3, Level: 3},
			{ID: 1, Level: 2},
		}},
	})
	testScan(t, tr, "d", []ReportEvent{
		{Index: 0, Label: 'd', Nodes: []ReportNode{
			{ID: 4, Level: 1},
		}},
	})
	testScan(t, tr, "abcde", []ReportEvent{
		{Index: 0, Label: 'a'},
		{Index: 1, Label: 'b', Nodes: []ReportNode{{ID: 1, Level: 2}}},
		{Index: 2, Label: 'c', Nodes: []ReportNode{{ID: 2, Level: 2}}},
		{Index: 3, Label: 'd', Nodes: []ReportNode{{ID: 4, Level: 1}}},
		{Index: 4, Label: 'e', Nodes: []ReportNode{{ID: 5, Level: 5}}},
	})
}

func TestTrie_UnmarshalError(t *testing.T) {
	bb := bytes.NewBuffer([]byte{})
	tr, err := Unmarshal(bb)
	if err == nil {
		t.Error("unexpected success")
	}
	if tr != nil{
		t.Errorf("unexpected trie is unmarshaled: %+v",tr)
	}
}
