package client

import (
	"reflect"
	"testing"
)

func TestLabelFilter(t *testing.T) {
	if got := labelFilter("key", ""); len(got) != 0 {
		t.Errorf("no team should be {}, got %v", got)
	}

	got := labelFilter("key", "WYE")
	want := map[string]any{"or": []map[string]any{
		{"team": map[string]any{"key": map[string]any{"eq": "WYE"}}},
		{"team": map[string]any{"null": true}},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
