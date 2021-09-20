/*
Package trie provides trie (prefix tree) algorithm.
*/
package trie

import (
	"context"
	"errors"
	"io"

	"github.com/koron-go/trietree"
)

// ErrFreezedAlready is that trie is freezed, can't be modified.
var ErrFreezedAlready = errors.New("freezed already")

// Trie provides trie tree.
type Trie struct {
	dy *trietree.DTree
	st *trietree.STree
}

// New creates a trie tree.
func New() *Trie {
	return &Trie{}
}

// Unmarshal reads and creates (unmarshals) a trie tree from io.Reader.
func Unmarshal(r io.Reader) (*Trie, error) {
	st, err := trietree.Read(r)
	if err != nil {
		return nil, err
	}
	return &Trie{st: st}, nil
}

// Put puts a keyword to a trie, if not be freezed.
func (tr *Trie) Put(s string) (int, error) {
	if tr.st != nil {
		return 0, ErrFreezedAlready
	}
	if tr.dy == nil {
		tr.dy = &trietree.DTree{}
	}
	return tr.dy.Put(s), nil
}

// Scan scans a string and found all keywords in it.
func (tr *Trie) Scan(ctx context.Context, s string, r Reporter) error {
	wrap := trietree.ScanReportFunc(func(src trietree.ScanEvent) {
		ev := ReportEvent{
			Index: src.Index,
			Label: src.Label,
		}
		if len(src.Nodes) > 0 {
			ev.Nodes = make([]ReportNode, len(src.Nodes))
			for i, n := range src.Nodes {
				ev.Nodes[i] = ReportNode{ID: n.ID, Level: n.Level}
			}
		}
		r.Report(ctx, ev)
	})
	if tr.st != nil {
		return tr.st.ScanContext(ctx, s, wrap)
	}
	return tr.dy.ScanContext(ctx, s, wrap)
}

// Freeze freezes a trie tree. Freezed trie can't be modified.
func (tr *Trie) Freeze() error {
	if tr.st != nil {
		return ErrFreezedAlready
	}
	tr.st = trietree.Freeze(tr.dy)
	tr.dy = nil
	return nil
}

// Marshal writes (marshals) a trie to io.Writer.
func (tr *Trie) Marshal(w io.Writer) error {
	if tr.st != nil {
		return tr.st.Write(w)
	}
	return trietree.Freeze(tr.dy).Write(w)
}

// Reporter receive reports of scan.
type Reporter interface {
	Report(ctx context.Context, ev ReportEvent)
}

// ReporterFunc is a utility type to implements Reporter.
type ReporterFunc func(ctx context.Context, ev ReportEvent)

// Report implements a method of Reporter.
func (f ReporterFunc) Report(ctx context.Context, ev ReportEvent) {
	f(ctx, ev)
}

// ReportEvent is an event which detected in Scan.
type ReportEvent struct {
	Index int
	Label rune
	Nodes []ReportNode
}

// ReportNode is a scanned node information.
type ReportNode struct {
	ID    int
	Level int
}
