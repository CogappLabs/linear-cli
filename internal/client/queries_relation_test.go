package client

import (
	"reflect"
	"testing"
)

func TestRelationIDsBetween(t *testing.T) {
	is := &Issue{}
	is.Relations.Nodes = []Relation{
		{ID: "r1", Type: "blocks", RelatedIssue: &RelIssue{Identifier: "WYE-2"}},
		{ID: "r2", Type: "related", RelatedIssue: &RelIssue{Identifier: "WYE-3"}},
	}
	is.InverseRelations.Nodes = []Relation{
		{ID: "r3", Type: "blocks", Issue: &RelIssue{Identifier: "WYE-2"}},
	}
	got := RelationIDsBetween(is, "wye-2")
	if !reflect.DeepEqual(got, []string{"r1", "r3"}) {
		t.Errorf("got %v", got)
	}
	if len(RelationIDsBetween(is, "WYE-9")) != 0 {
		t.Error("expected no matches")
	}
}
