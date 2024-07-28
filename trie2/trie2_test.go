package trie2

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type Data struct {
	N int
	S string
}

func TestMarshal0(t *testing.T) {
	dt := DTrie[Data]{}
	bb := &bytes.Buffer{}
	if err := dt.Freeze(false).Marshal(bb, nil); err != nil {
		t.Fatalf("failed to marshal: %s", err)
	}
	st, err := Unmarshal[Data](bb, nil)
	if err != nil {
		t.Fatalf("failed to unmarshal: %s", err)
	}
	if d := cmp.Diff(dt.values, st.values); d != "" {
		t.Errorf("failed unmarshal values: -want +got\n%s", d)
	}
}

func TestMarshalValue(t *testing.T) {
	dt := DTrie[Data]{}
	dt.Put("a", Data{111, "aaa"})
	dt.Put("ab", Data{222, "bbb"})
	dt.Put("abc", Data{333, "ccc"})
	dt.Put("d", Data{444, "ddd"})
	dt.Put("de", Data{555, "eee"})
	bb := &bytes.Buffer{}
	if err := dt.Freeze(false).Marshal(bb, nil); err != nil {
		t.Fatalf("failed to marshal: %s", err)
	}
	st, err := Unmarshal[Data](bb, nil)
	if err != nil {
		t.Fatalf("failed to unmarshal: %s", err)
	}
	if d := cmp.Diff(dt.values, st.values); d != "" {
		t.Errorf("failed unmarshal values: -want +got\n%s", d)
	}
}

func TestMarshalPointer(t *testing.T) {
	dt := DTrie[*Data]{}
	dt.Put("a", &Data{111, "aaa"})
	dt.Put("ab", &Data{222, "bbb"})
	dt.Put("abc", &Data{333, "ccc"})
	dt.Put("d", &Data{444, "ddd"})
	dt.Put("de", &Data{555, "eee"})
	bb := &bytes.Buffer{}
	if err := dt.Freeze(false).Marshal(bb, nil); err != nil {
		t.Fatalf("failed to marshal: %s", err)
	}
	st, err := Unmarshal[*Data](bb, nil)
	if err != nil {
		t.Fatalf("failed to unmarshal: %s", err)
	}
	if d := cmp.Diff(dt.values, st.values); d != "" {
		t.Errorf("failed unmarshal values: -want +got\n%s", d)
	}
}

type DataJSON struct {
	N int    `json:"n"`
	S string `json:"s"`
}

func TestMarshalCustom(t *testing.T) {
	dt := DTrie[DataJSON]{}
	dt.Put("a", DataJSON{111, "aaa"})
	dt.Put("ab", DataJSON{222, "bbb"})
	dt.Put("abc", DataJSON{333, "ccc"})
	dt.Put("d", DataJSON{444, "ddd"})
	dt.Put("de", DataJSON{555, "eee"})
	bb := &bytes.Buffer{}
	err := dt.Freeze(true).Marshal(bb, func(w io.Writer, values []DataJSON) error {
		return json.NewEncoder(w).Encode(values)
	})
	if err != nil {
		t.Fatalf("failed to marshal: %s", err)
	}
	st, err := Unmarshal[DataJSON](bb, func(r io.Reader, n int) ([]DataJSON, error) {
		values := make([]DataJSON, 0, n)
		if err := json.NewDecoder(r).Decode(&values); err != nil {
			return nil, err
		}
		return values, nil
	})
	if err != nil {
		t.Fatalf("failed to unmarshal: %s", err)
	}
	if d := cmp.Diff(dt.values, st.values); d != "" {
		t.Errorf("failed unmarshal values: -want +got\n%s", d)
	}
}

func TestLongestPrefixDTree(t *testing.T) {
	dt := DTrie[Data]{}
	dt.Put("a", Data{111, "aaa"})
	dt.Put("ab", Data{222, "bbb"})
	dt.Put("abc", Data{333, "ccc"})
	dt.Put("d", Data{444, "ddd"})
	dt.Put("de", Data{555, "eee"})
	for i, c := range []struct {
		query string
		wantV Data
		wantP string
		wantF bool
	}{
		{"az", Data{111, "aaa"}, "a", true},
		{"za", Data{}, "", false},
		{"abcde", Data{333, "ccc"}, "abc", true},
		{"ababc", Data{222, "bbb"}, "ab", true},
	} {
		gotV, gotP, gotF := dt.LongestPrefix(c.query)
		if gotF != c.wantF {
			t.Errorf("existence unmatch #%d: want=%t got=%t", i, c.wantF, gotF)
			continue
		}
		if gotP != c.wantP {
			t.Errorf("prefix unmatch #%d: want=%s got=%s", i, c.wantP, gotP)
		}
		if d := cmp.Diff(c.wantV, gotV); d != "" {
			t.Errorf("values unmatch #%d: -want +got\n%s", i, d)
		}
	}
}

func TestLongestPrefixSTree(t *testing.T) {
	dt := DTrie[Data]{}
	dt.Put("a", Data{111, "aaa"})
	dt.Put("ab", Data{222, "bbb"})
	dt.Put("abc", Data{333, "ccc"})
	dt.Put("d", Data{444, "ddd"})
	dt.Put("de", Data{555, "eee"})
	st := dt.Freeze(false)
	for i, c := range []struct {
		query string
		wantV Data
		wantP string
		wantF bool
	}{
		{"az", Data{111, "aaa"}, "a", true},
		{"za", Data{}, "", false},
		{"abcde", Data{333, "ccc"}, "abc", true},
		{"ababc", Data{222, "bbb"}, "ab", true},
	} {
		gotV, gotP, gotF := st.LongestPrefix(c.query)
		if gotF != c.wantF {
			t.Errorf("existence unmatch #%d: want=%t got=%t", i, c.wantF, gotF)
			continue
		}
		if gotP != c.wantP {
			t.Errorf("prefix unmatch #%d: want=%s got=%s", i, c.wantP, gotP)
		}
		if d := cmp.Diff(c.wantV, gotV); d != "" {
			t.Errorf("values unmatch #%d: -want +got\n%s", i, d)
		}
	}
}
