package client

import (
	"reflect"
	"testing"
)

func TestBuildIssueFilter(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		if got := buildIssueFilter(IssueFilter{}); len(got) != 0 {
			t.Errorf("empty filter should be {}, got %v", got)
		}
	})

	t.Run("team and assignee", func(t *testing.T) {
		got := buildIssueFilter(IssueFilter{TeamKey: "ENG", AssigneeID: "u1"})
		want := map[string]any{
			"team":     map[string]any{"key": map[string]any{"eq": "ENG"}},
			"assignee": map[string]any{"id": map[string]any{"eq": "u1"}},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("multiple states use in", func(t *testing.T) {
		got := buildIssueFilter(IssueFilter{StateTypes: []string{"started", "unstarted"}})
		want := map[string]any{
			"state": map[string]any{"type": map[string]any{"in": []string{"started", "unstarted"}}},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("not-done uses nin and overrides states", func(t *testing.T) {
		got := buildIssueFilter(IssueFilter{NotDone: true, StateTypes: []string{"started"}})
		want := map[string]any{
			"state": map[string]any{"type": map[string]any{"nin": []string{"completed", "canceled"}}},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func TestMarkdownImageRe(t *testing.T) {
	md := "text ![alt](https://a.com/x.png) more ![](https://b.com/y.jpg) end"
	matches := markdownImageRe.FindAllStringSubmatch(md, -1)
	if len(matches) != 2 {
		t.Fatalf("expected 2 image matches, got %d", len(matches))
	}
	if matches[0][1] != "https://a.com/x.png" || matches[1][1] != "https://b.com/y.jpg" {
		t.Errorf("captured URLs wrong: %v", matches)
	}
}
